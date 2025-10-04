package authentication

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/users"
	"github.com/ooqls/go-cache/store"
	"github.com/ooqls/go-db/redis"
	"github.com/ooqls/go-crypto/crypto"
	"github.com/ooqls/go-db/testutils"
	"github.com/stretchr/testify/assert"
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
	challenger := NewChallengerV1(store.NewMemStore("test", time.Second * 10))
	
	// should not get a result because the challenge does not exist
	res, err := challenger.VerifyChallenge(context.Background(), uuid.New(), []byte("fnjnjekw"))
	assert.Nilf(t, res, "should not get a result")
	assert.NotNilf(t, err, "should get an error")
}

func TestChallenger_VerifyChallenge(t *testing.T) {
    testutils.StartRedis(context.Background())

	type TestCase struct {
		description    string
		solveChallenge func(t *testing.T, challenge string) ([]byte, error)
		user           *users.User
		shouldVerify   bool
	}

	memStore := store.NewRedisStore("test", *redis.GetConnection(), time.Second * 10)
	userKeyA, userKeyBytesA, saltA := generateUserKey(t)
	userKeyB, _, _ := generateUserKey(t)

	cases := []TestCase{
		{
			description: "should verify challenge",

			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyA.Encrypt([]byte(challenge))
			},
			user: &users.User{
				ID: uuid.New(),
				Key:  userKeyBytesA,
				Salt: saltA[:],
				CreatedAt: time.Now(),
				Username: "testuser",
				Email: "testuser@test.com",
			},
			shouldVerify: true,
		},
		{
			description: "should not verify challenge because the challenge is incorrect",
			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyA.Encrypt([]byte("fnjnjekw"))
			},
			user: &users.User{
				ID: uuid.New(),
				Key:  userKeyBytesA,
				Salt: saltA[:],
				CreatedAt: time.Now(),
				Username: "testuser",
				Email: "testuser@test.com",
			},
			shouldVerify: false,
		},
		{
			description: "should not verify challenge because the wrong key was used",
			solveChallenge: func(t *testing.T, challenge string) ([]byte, error) {
				return userKeyB.Encrypt([]byte(challenge))
			},
			user: &users.User{
				ID: uuid.New(),
				Key:  userKeyBytesA,
				Salt: saltA[:],
				CreatedAt: time.Now(),
				Username: "testuser",
				Email: "testuser@test.com",
			},
			shouldVerify: false,
		},
	}

	for _, tc := range cases {

		challenger := NewChallengerV1(memStore)

		challenge, err := challenger.IssueChallenge(context.Background(), tc.user)
		assert.Nilf(t, err, "%s: should not get an error when getting key", tc.description)
		assert.NotNilf(t, challenge, "%s: should get a challenge", tc.description)
		assert.Equalf(t, tc.user.ID, challenge.User.ID, "%s: userid should equal challenge user id", tc.description)

		b, err := tc.solveChallenge(t, string(challenge.Challenge))
		assert.Nilf(t, err, "%s: should not fail to solve challenge", tc.description)

		result, err := challenger.VerifyChallenge(context.Background(), challenge.ID, b)

		if tc.shouldVerify {
			assert.NotNilf(t, result, "%s: should be verified", tc.description)
			assert.Equalf(t, tc.user.ID, result.User.ID, "%s: userid should equal challenge user id", tc.description)
			assert.Nilf(t, err, "%s: should not get an error when verifying challenge", tc.description)
		} else {
			assert.Nilf(t, result, "%s: should not be verified", tc.description)
			assert.NotNilf(t, err, "%s: should get an error when verifying challenge", tc.description)
		}
	}
}
