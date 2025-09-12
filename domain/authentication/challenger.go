package authentication

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/users"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-cache/store"
	"github.com/ooqls/go-crypto/crypto"
	"go.uber.org/zap"
)

var (
	ErrChallengeExpired error = errors.New("challenge expired")
	ErrChallengeFailed  error = errors.New("challenge failed")
)

type Challenge struct {
	ID        uuid.UUID  `json:"id"`
	User      users.User `json:"user"`
	Challenge []byte     `json:"challenge"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewChallenge(user *users.User) Challenge {
	return Challenge{
		ID:        uuid.New(),
		User:      *user,
		Challenge: []byte(uuid.New().String()),
		CreatedAt: time.Now(),
	}
}

type AuthedResult struct {
	ChallengeID uuid.UUID   `json:"challenge_id"`
	User        *users.User `json:"user"`
}

//go:generate go run github.com/golang/mock/mockgen -source=challenger.go -destination=mocks/mock_challenger.go -package=mocks -mock_names=Challenger=MockChallenger
type Challenger interface {
	IssueChallenge(ctx context.Context, user *users.User) (*Challenge, error)
	VerifyChallenge(ctx context.Context, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error)
	VerifyRegistration(ctx context.Context, username string, secret []byte, key []byte) ([crypto.SALT_SIZE]byte, error)
}

type ChallengerV1 struct {
	store store.GenericInterface
}

func NewChallengerV1(store store.GenericInterface) Challenger {
	return &ChallengerV1{
		store: store,
	}
}

func (c *ChallengerV1) IssueChallenge(ctx context.Context, user *users.User) (*Challenge, error) {
	var cachedChallenge Challenge
	l.Info("issuing challenge for user", zap.String("user_id", user.ID.String()))
	err := c.store.Get(ctx, user.ID.String(), &cachedChallenge)
	if err != nil && !cache.IsCacheMissErr(err) {
		l.Error("failed to get challenge from store", zap.String("user_id", user.ID.String()), zap.Error(err))
		return nil, ErrInternal
	}

	if err == nil {
		return &cachedChallenge, nil
	}

	chal := NewChallenge(user)
	l.Sugar().Infow("storing challenge", "chal", chal)

	err = c.store.Set(ctx, chal.ID.String(), chal)
	if err != nil {
		l.Error("failed to set challenge in store", zap.String("user_id", user.ID.String()), zap.Error(err))
		return nil, ErrInternal
	}

	return &chal, nil
}

func (c *ChallengerV1) VerifyChallenge(ctx context.Context, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error) {
	var challenge Challenge
	err := c.store.Get(ctx, challengeId.String(), &challenge)
	if err != nil {
		if cache.IsCacheMissErr(err) {
			return nil, ErrChallengeExpired
		}

		l.Error("failed to get challenge from store", zap.String("challenge_id", challengeId.String()), zap.Error(err))
		return nil, ErrInternal
	}

	decryptedChallenge, err := crypto.AESGCMDecryptWithKey(challenge.User.Key, solvedChallenge)
	if err != nil {
		l.Warn("failed to decrypt given challenge with user key", zap.Error(err))
		return nil, ErrChallengeFailed
	}

	authed := bytes.Equal(decryptedChallenge, challenge.Challenge)

	if !authed {
		return nil, ErrChallengeFailed
	}

	err = c.store.Delete(ctx, challengeId.String())
	if err != nil {
		l.Error("failed to delete challenge from store", zap.String("challenge_id", challengeId.String()), zap.Error(err))
	}

	return &AuthedResult{
		ChallengeID: challenge.ID,
		User:        &challenge.User,
	}, nil
}

func (c *ChallengerV1) VerifyRegistration(ctx context.Context, username string, secret []byte, key []byte) ([crypto.SALT_SIZE]byte, error) {
	l := l.With(zap.String("username", username))
	err := crypto.VerifyGCMAESKey(key)
	if err != nil {
		l.Warn("failed to verify key", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	decryptedSecret, err := crypto.AESGCMDecryptWithKey(key, secret)
	if err != nil {
		l.Warn("failed to decrypt secret", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	if string(decryptedSecret) != string(username) {
		l.Warn("decrypted secret does not match username", zap.String("decryptedSecret", string(decryptedSecret)), zap.String("username", username))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}
	
    usernameBytes := make([]byte, 8)
	copy(usernameBytes, []byte(username))
	seed := binary.LittleEndian.Uint64(usernameBytes)
	rng := crypto.NewPCG32(seed, 0)

	salt := make([]byte, crypto.SALT_SIZE)
	rng.Read(salt)

	givenSalt, _, _, err := crypto.DecodeAESGCM(secret)
	if err != nil {
		l.Warn("failed to decode secret", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	if !bytes.Equal(salt, givenSalt[:]) {
		l.Warn("given salt does not match expected calculated salt", zap.ByteString("expected", salt), zap.ByteString("given", givenSalt[:]))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	return [crypto.SALT_SIZE]byte(salt), nil
}
