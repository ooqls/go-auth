package authenticationv1

import (
	gocontext "context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/store"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/redis"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func generateUserKey(t *testing.T) (crypto.Algorithm, []byte, [16]byte) {
	var salt [16]byte
	rand.Read(salt[:])

	userKey := crypto.NewAESGCMAlgorithm("password", salt)
	userKeyBytes, err := userKey.GetKey()
	assert.Nilf(t, err, "should not get an error when getting user key")
	return userKey, userKeyBytes, salt
}

func TestChallenger_IssueChallenge(t *testing.T) {
	challenger := NewChallengerV1(store.NewMemStore("test", time.Second*10))
	ctx := contexts.NewLoggingContext(gocontext.Background(), zap.L())

	// should not get a result because the challenge does not exist
	res, err := challenger.VerifyChallenge(ctx, uuid.New(), []byte("fnjnjekw"))
	assert.Nilf(t, res, "should not get a result")
	assert.NotNilf(t, err, "should get an error")
}

func TestChallenger_VerifyChallenge(t *testing.T) {
	ctx := gocontext.Background()
	containers.StartRedis(ctx)

	type TestCase struct {
		description    string
		solveChallenge func(t *testing.T, challenge string) ([]byte, error)
		user           *usersv1.User
		shouldVerify   bool
	}

	memStore := store.NewRedisStore("test", *redis.GetConnection(ctx), time.Second*10)
	userKeyA, userKeyBytesA, saltA := generateUserKey(t)
	userKeyB, _, _ := generateUserKey(t)

	cases := []TestCase{
		{
			description: "should verify challenge",

			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyA.Encrypt([]byte(challenge))
			},
			user: &usersv1.User{
				Object: corev1.Object{
					Id:        uuid.New(),
					CreatedAt: time.Now(),
				},
				Key:      userKeyBytesA,
				Salt:     saltA[:],
				Username: "testuser",
				Email:    "testuser@test.com",
			},
			shouldVerify: true,
		},
		{
			description: "should not verify challenge because the challenge is incorrect",
			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyA.Encrypt([]byte("fnjnjekw"))
			},
			user: &usersv1.User{
				Object: corev1.Object{
					Id:        uuid.New(),
					CreatedAt: time.Now(),
				},
				Key:      userKeyBytesA,
				Salt:     saltA[:],
				Username: "testuser",
				Email:    "testuser@test.com",
			},
			shouldVerify: false,
		},
		{
			description: "should not verify challenge because the wrong key was used",
			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyB.Encrypt([]byte(challenge))
			},
			user: &usersv1.User{
				Object: corev1.Object{
					Id:        uuid.New(),
					CreatedAt: time.Now(),
				},
				Key:      userKeyBytesA,
				Salt:     saltA[:],
				Username: "testuser",
				Email:    "testuser@test.com",
			},
			shouldVerify: false,
		},
	}

	for _, tc := range cases {

		challenger := NewChallengerV1(memStore)

		lc := contexts.NewLoggingContext(gocontext.Background(), zap.L())

		challenge, err := challenger.IssueChallenge(lc, tc.user)
		assert.Nilf(t, err, "%s: should not get an error when getting key", tc.description)
		assert.NotNilf(t, challenge, "%s: should get a challenge", tc.description)
		assert.Equalf(t, tc.user.Object.Id, challenge.User.Object.Id, "%s: userid should equal challenge user id", tc.description)

		b, err := tc.solveChallenge(t, string(challenge.Challenge))
		assert.Nilf(t, err, "%s: should not fail to solve challenge", tc.description)

		result, err := challenger.VerifyChallenge(lc, challenge.ID, b)

		if tc.shouldVerify {
			assert.NotNilf(t, result, "%s: should be verified", tc.description)
			assert.Equalf(t, tc.user.Object.Id, result.User.Object.Id, "%s: userid should equal challenge user id", tc.description)
			assert.Nilf(t, err, "%s: should not get an error when verifying challenge", tc.description)
		} else {
			assert.Nilf(t, result, "%s: should not be verified", tc.description)
			assert.NotNilf(t, err, "%s: should get an error when verifying challenge", tc.description)
		}
	}
}
