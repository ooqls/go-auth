package loginchallenges

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/gen"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)
//go:generate go run github.com/golang/mock/mockgen -source=challenges.go -destination=mocks/mock_challenges.go -package=mocks -mock_names=ChallengeReader=MockChallengeReader -mock_names=ChallengeWriter=MockChallengeWriter
type Reader interface {
	GetChallenge(ctx context.Context, userID uuid.UUID) (*Challenge, error)
}

type Writer interface {
	CreateChallenge(ctx context.Context, userID uuid.UUID, salt []byte) (*Challenge, error)
}

type SQLReader struct {
	cache cache.Cache[Challenge]
	q     gen.Queries
	l     *zap.Logger
}

func NewSQLReader(q gen.Queries, cache cache.Cache[Challenge]) Reader {
	return &SQLReader{
		q:     q,
		cache: cache,
		l:     log.NewLogger("challenge_reader"),
	}
}

func (r *SQLReader) GetChallenge(ctx context.Context, challengeID uuid.UUID) (*Challenge, error) {
	cachedChallenge, err := r.cache.Get(ctx, challengeID.String())
	if err != nil && !errors.Is(err, redis.Nil) {
		r.l.Warn("failed to get challenge from cache", zap.Error(err))
	}

	if err == nil {
		return cachedChallenge, nil
	}

	challenge, err := r.q.GetChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &challenge, nil
}

type SQLWriter struct {
	q gen.Queries
}

func NewSQLWriter(q gen.Queries) Writer {
	return &SQLWriter{q: q}
}

func (r *SQLWriter) CreateChallenge(ctx context.Context, userID uuid.UUID, salt []byte) (*Challenge, error) {
	challengeStr := uuid.New().String()
	challenge, err := r.q.CreateChallenge(ctx, gen.CreateChallengeParams{
		UserID:    userID,
		Challenge: challengeStr,
		Salt:      salt,
		ExpiresAt: time.Now().Add(time.Minute * 10),
	})

	if err != nil {
		return nil, err
	}

	return &challenge, nil
}
