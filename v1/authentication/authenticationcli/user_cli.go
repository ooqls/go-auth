package authenticationcli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/v1/authentication/api/gen_authentication"
)

func unmarshalResponse[T any](resp *http.Response) (*T, error) {
	var body T
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &body, nil
}

func unmarshalError(resp *http.Response) error {
	var err gen_authentication.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
		return err
	}
	return errors.New(err.Error)
}

type AuthenticationClient struct {
	c gen_authentication.Client
}

func NewAuthenticationClient(c gen_authentication.Client) *AuthenticationClient {
	return &AuthenticationClient{c: c}
}

func (c *AuthenticationClient) Register(ctx context.Context, email string, password string, username string) (*gen_authentication.RegisterResponse, error) {
	salt := authenticationv1.GenerateSalt(username)
	key, err := crypto.DeriveAESGCMKey(password, [16]byte(salt))
	if err != nil {
		return nil, err
	}

	secret, err := crypto.AESGCMEncryptWithKey(key, salt, []byte(username))
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Register(ctx, gen_authentication.RegisterJSONRequestBody{
		Email:    email,
		Key:      key,
		Username: username,
		Secret:   secret,
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = unmarshalError(resp)
		return nil, err
	}

	body, err := unmarshalResponse[gen_authentication.RegisterResponse](resp)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *AuthenticationClient) Login(ctx context.Context, username string, password string) (uid *string, okey *string, rkey *string, err error) {
	challengeResp, err := c.c.LoginChallenge(ctx, gen_authentication.LoginChallengeJSONRequestBody{
		Username: username,
	})
	if err != nil {

		return nil, nil, nil, fmt.Errorf("failed to get login challenge: %v", err)
	}

	if challengeResp.StatusCode != 200 {
		err = unmarshalError(challengeResp)
		return nil, nil, nil, fmt.Errorf("got a non 200 status from api: %s: %v", challengeResp.Status, err)
	}

	challenge, err := unmarshalResponse[gen_authentication.ChallengeServerResponse](challengeResp)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	challengeStr := challenge.Challenge
	salt := challenge.Salt

	key, err := crypto.DeriveAESGCMKey(password, [16]byte(salt))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to derive key: %v", err)
	}

	log.Printf("key: %v", key)
	log.Printf("challengeStr: %v", challengeStr)
	log.Printf("salt: %v", salt)

	encrypted, err := crypto.AESGCMEncryptWithKey(key, [16]byte(salt), []byte(challengeStr))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encrypt key: %v", err)
	}

	resp, err := c.c.LoginChallengeResponse(ctx, gen_authentication.LoginChallengeResponseJSONRequestBody{
		Challenge: encrypted,
		Id:        challenge.Id,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to send login challenge response: %v", err)
	}

	if resp.StatusCode != 200 {
		err = unmarshalError(resp)
		return nil, nil, nil, err
	}

	cookies := resp.Cookies()

	for _, cookie := range cookies {
		switch cookie.Name {
		case "RKEY":
			rkey = &cookie.Value
		case "OKEY":
			okey = &cookie.Value
		case "UID":
			uid = &cookie.Value
		}
	}

	return
}
