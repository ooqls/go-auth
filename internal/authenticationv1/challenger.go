package authenticationv1

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/cache/store"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/usersv1"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

type Challenge struct {
	ID        uuid.UUID    `json:"id"`
	User      usersv1.User `json:"user"`
	Challenge []byte       `json:"challenge"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewChallenge(user *usersv1.User, challenge []byte) Challenge {
	return Challenge{
		ID:        uuid.New(),
		User:      *user,
		Challenge: challenge,
		CreatedAt: time.Now(),
	}
}

type AuthedResult struct {
	ChallengeID uuid.UUID     `json:"challenge_id"`
	User        *usersv1.User `json:"user"`
}

//go:generate go run github.com/golang/mock/mockgen -source=challenger.go -destination=mocks/mock_challenger.go -package=mocks -mock_names=Challenger=MockChallenger
type Challenger interface {
	IssueChallenge(ctx contexts.LContext, user *usersv1.User) (*Challenge, error)
	VerifyChallenge(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error)
	VerifyRegistration(ctx contexts.LContext, username string, secret []byte, key []byte) ([16]byte, error)
}

type ChallengerV1 struct {
	store store.GenericInterface
}

func NewChallengerV1(store store.GenericInterface) Challenger {
	return &ChallengerV1{
		store: store,
	}
}

func loggerFrom(ctx contexts.LContext) *zap.Logger {
	if ctx != nil && ctx.L() != nil {
		return ctx.L()
	}

	return zap.L()
}

func (c *ChallengerV1) IssueChallenge(ctx contexts.LContext, user *usersv1.User) (*Challenge, error) {
	logger := loggerFrom(ctx)
	var cachedChallenge Challenge
	logger.Info("issuing challenge for user", zap.String("user_id", user.Object.Id.String()))
	err := c.store.Get(ctx, user.Object.Id.String(), &cachedChallenge)
	if err != nil && !cache.IsCacheMissErr(err) {
		logger.Error("failed to get challenge from store", zap.String("user_id", user.Object.Id.String()), zap.Error(err))
		return nil, ErrInternal
	}

	if err == nil {
		return &cachedChallenge, nil
	}
	chalBytes := []byte(uuid.New().String())
	logger.Debug("generated challenge bytes", zap.ByteString("challenge", chalBytes))
	chal := NewChallenge(user, chalBytes)
	logger.Sugar().Infow("storing challenge", "chal", chal)

	err = c.store.Set(ctx, chal.ID.String(), chal)
	if err != nil {
		logger.Error("failed to set challenge in store", zap.String("user_id", user.Object.Id.String()), zap.Error(err))
		return nil, v1.ErrInternal(err, v1.M{"user_id": user.Object.Id.String()})
	}

	return &chal, nil
}

func (c *ChallengerV1) VerifyChallenge(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error) {
	logger := loggerFrom(ctx)
	var challenge Challenge
	logger.Info("verifying challenge", zap.String("challenge_id", challengeId.String()))
	err := c.store.Get(ctx, challengeId.String(), &challenge)
	if err != nil {
		if cache.IsCacheMissErr(err) {
			return nil, ErrChallengeExpired
		}

		logger.Error("failed to get challenge from store", zap.String("challenge_id", challengeId.String()), zap.Error(err))
		return nil, v1.ErrInternal(err, v1.M{"challenge_id": challengeId.String()})
	}
	logger.Info("decrypting challenge",
		zap.Any("solvedChallenge", solvedChallenge),
		zap.Any("challenge", challenge.Challenge),
		zap.Binary("userKey", challenge.User.Key),
	)

	decryptedChallenge, err := crypto.AESGCMDecryptWithKey(challenge.User.Key, solvedChallenge)
	if err != nil {
		logger.Warn("failed to decrypt given challenge with user key", zap.Error(err))
		return nil, ErrChallengeFailed
	}

	authed := bytes.Equal(decryptedChallenge, challenge.Challenge)

	if !authed {
		return nil, ErrChallengeFailed
	}

	err = c.store.Delete(ctx, challengeId.String())
	if err != nil {
		logger.Error("failed to delete challenge from store", zap.String("challenge_id", challengeId.String()), zap.Error(err))
	}

	return &AuthedResult{
		ChallengeID: challenge.ID,
		User:        &challenge.User,
	}, nil
}

func (c *ChallengerV1) VerifyRegistration(ctx contexts.LContext, username string, secret []byte, key []byte) ([crypto.SALT_SIZE]byte, error) {
	logger := loggerFrom(ctx).With(zap.String("username", username))

	logger.Info("verifing key and secret for registration", zap.String("username", username))
	err := crypto.VerifyGCMAESKey(key)
	if err != nil {
		logger.Warn("failed to verify key", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	logger.Info(("decrypting secret with provided key"))
	decryptedSecret, err := crypto.AESGCMDecryptWithKey(key, secret)
	if err != nil {
		logger.Warn("failed to decrypt secret", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	if string(decryptedSecret) != string(username) {
		logger.Warn("decrypted secret does not match username", zap.String("decryptedSecret", string(decryptedSecret)), zap.String("username", username))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	logger.Info("calculating expected salt for username", zap.String("username", username))
	salt := GenerateSalt(username)
	givenSalt, _, _, err := crypto.DecodeAESGCM(secret)
	if err != nil {
		logger.Warn("failed to decode secret", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	if !bytes.Equal(salt[:], givenSalt[:]) {
		logger.Warn("given salt does not match expected calculated salt", zap.ByteString("expected", salt[:]), zap.ByteString("given", givenSalt[:]))
		return [crypto.SALT_SIZE]byte{}, ErrInvalidRegistration
	}

	return [crypto.SALT_SIZE]byte(salt), nil
}
