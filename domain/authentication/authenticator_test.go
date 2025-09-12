package authentication

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/users"
	"github.com/ooqls/go-cache/factory"
	"github.com/ooqls/go-cache/store"
	"github.com/ooqls/go-crypto/crypto"
	"github.com/ooqls/go-crypto/jwt"
	"github.com/ooqls/go-crypto/keys"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticator_IsAuthenticated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	rKey, err := keys.NewRSA()
	assert.Nilf(t, err, "failed to create new key: %v", err)
	jwtKey := keys.NewJWTKey(*rKey)
	issuer := jwt.NewJwtTokenIssuer[UserClaims](&jwt.TokenConfiguration{
		Issuer:                  "test",
		Audience:                []string{"test"},
		ValidityDurationSeconds: 300,
	}, jwtKey)

	refreshIssuer := jwt.NewJwtTokenIssuer[UserClaims](&jwt.TokenConfiguration{
		Issuer:                  "refresh",
		Audience:                []string{"test"},
		ValidityDurationSeconds: 300,
	}, jwtKey)

	type TestCase struct {
		description        string
		solveChallenge     func(t *testing.T, challenge string, key crypto.Algorithm) []byte
		shouldAuthenticate bool
	}
	testCases := []TestCase{
		{
			description: "valid challenge",
			solveChallenge: func(t *testing.T, challenge string, key crypto.Algorithm) []byte {
				b, err := key.Encrypt([]byte(challenge))
				assert.Nilf(t, err, "should not fail to encrypt challenge")
				return b
			},
			shouldAuthenticate: true,
		},
		{
			description: "invalid password hash",
			solveChallenge: func(t *testing.T, challenge string, key crypto.Algorithm) []byte {
				return []byte(challenge)
			},
			shouldAuthenticate: false,
		},
	}

	for _, tc := range testCases {
		salt := generateRandomSalt()
		saltB := []byte(salt)
		key, err := crypto.DeriveAESGCMKey("password", [16]byte(saltB))
		assert.Nilf(t, err, "%s: should not get an error when deriving key", tc.description)
		keyAlgo := crypto.NewAESGCMAlgorithmWithKey(key, [16]byte(saltB))

		user := &users.User{
			ID:       uuid.New(),
			Username: "test",
			Email:    "test",
			Key:      key,
			Salt:     saltB,
		}
		challenger := NewChallengerV1(store.NewMemStore("test", time.Second*10))

		authenticator := NewAuthenticatorV1(
			issuer,
			refreshIssuer,
			&factory.MemCacheFactory{},
			challenger,
			[]string{"test"},
		)

		resp, err := authenticator.ChallengeRequest(ctx, user)
		assert.Nilf(t, err, "ChallengeRequest should not return an error: %v", err)

		response := tc.solveChallenge(t, string(resp.Challenge), keyAlgo)

		okey, rkey, uid, err := authenticator.ChallengeResponse(ctx, resp.ID, response)
		if tc.shouldAuthenticate {
			assert.Nilf(t, err, "ChallengeResponse should not return an error: %v", err)
			assert.NotNilf(t, okey, "okey should not be nil")
			assert.NotNilf(t, rkey, "rkey should not be nil")
			assert.NotNilf(t, uid, "uid should not be nil")
		} else {
			assert.NotNilf(t, err, "ChallengeResponse should return an error: %v", err)
		}
	}
}
