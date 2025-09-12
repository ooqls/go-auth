package authentication

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ooqls/go-auth/api/v1/gen/gen_authentication"
	"github.com/ooqls/go-crypto/crypto"
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
	usernameBytes := make([]byte, 8)
	copy(usernameBytes, []byte(username))

	seed := binary.LittleEndian.Uint64(usernameBytes)
	salt := make([]byte, crypto.SALT_SIZE)
	rng := crypto.NewPCG32(seed, 0)
	rng.Read(salt)

	key, err := crypto.DeriveAESGCMKey(password, [16]byte(salt))
	if err != nil {
		return nil, err
	}

	encrypted, err := crypto.AESGCMEncryptWithKey(key, [16]byte(salt), []byte(username))
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Register(ctx, gen_authentication.RegisterJSONRequestBody{
		Email:           email,
		Base64Key:       base64.StdEncoding.EncodeToString(key),
		Username:        username,
		EncryptedSecret: base64.StdEncoding.EncodeToString(encrypted),
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
		return nil, nil, nil, err
	}

	if challengeResp.StatusCode != 200 {
		err = unmarshalError(challengeResp)
		return nil, nil, nil, err
	}

	challenge, err := unmarshalResponse[gen_authentication.ChallengeServerResponse](challengeResp)
	if err != nil {
		return nil, nil, nil, err
	}

	challengeStr, err := base64.StdEncoding.DecodeString(challenge.Base64Challenge)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.StdEncoding.DecodeString(challenge.Base64Salt)
	if err != nil {
		return nil, nil, nil, err
	}

	key, err := crypto.DeriveAESGCMKey(password, [16]byte(salt))
	if err != nil {
		return nil, nil, nil, err
	}

	log.Printf("key: %v", key)
	log.Printf("challengeStr: %v", challengeStr)
	log.Printf("salt: %v", salt)

	encrypted, err := crypto.AESGCMEncryptWithKey(key, [16]byte(salt), []byte(challengeStr))
	if err != nil {
		return nil, nil, nil, err
	}

	resp, err := c.c.LoginChallengeResponse(ctx, gen_authentication.LoginChallengeResponseJSONRequestBody{
		Base64Challenge: base64.StdEncoding.EncodeToString(encrypted),
		Id:              challenge.Id,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	if resp.StatusCode != 200 {
		err = unmarshalError(resp)
		return nil, nil, nil, err
	}

	cookies := resp.Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "RKEY" {
			rkey = &cookie.Value
		} else if cookie.Name == "OKEY" {
			okey = &cookie.Value
		} else if cookie.Name == "UID" {
			uid = &cookie.Value
		}
	}

	return
}
