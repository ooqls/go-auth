package authenticationv1

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/cache/store"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthenticator_Challenge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := contexts.NewLoggingContext(gocontext.Background(), zap.L())
	rKey, err := keys.NewRSA()
	assert.Nilf(t, err, "failed to create new key: %v", err)
	jwtKey := keys.NewJWTKey(*rKey)
	issuer := jwt.NewJwtTokenIssuer[claims.UserClaims](&jwt.TokenConfiguration{
		Issuer:                  "test",
		Audience:                []string{"test"},
		ValidityDurationSeconds: 300,
	}, jwtKey)

	refreshIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](&jwt.TokenConfiguration{
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

		user := &usersv1.User{
			Object: corev1.Object{
				Metadata: usersv1.Metadata,
				Id:       uuid.New(),
			},
			Username: "test",
			Email:    "test",
			Key:      key,
			Salt:     saltB,
		}
		challenger := NewChallengerV1(store.NewMemStore("test", time.Second*10))
		userTokenIssuer := NewUserTokenIssuer(issuer, refreshIssuer)
		authenticator := NewAuthenticatorV1(
			userTokenIssuer,
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

			claims, err := authenticator.IsAuthenticated(ctx, okey)
			assert.Nilf(t, err, "IsAuthenticated should not return an error: %v", err)
			assert.NotNilf(t, claims, "claims should not be nil")
			assert.Equal(t, uid, claims.UserID.String())
		} else {
			assert.NotNilf(t, err, "ChallengeResponse should return an error: %v", err)
			claims, err := authenticator.IsAuthenticated(ctx, okey)
			assert.NotNilf(t, err, "IsAuthenticated should return an error: %v", err)
			assert.Nilf(t, claims, "claims should be nil")
		}
	}
}
