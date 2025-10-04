package challengeattempts

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/gen"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-log"
	"go.uber.org/zap"
)

//go:generate go run github.com/golang/mock/mockgen -source=challenge_attempts.go -destination=mocks/mock_challenge_attempts.go -package=mocks -mock_names=ChallengeAttemptReader=MockReader -mock_names=ChallengeAttemptWriter=MockWriter
type Reader interface {
	GetChallengeAttempts(ctx context.Context, userID string) ([]ChallengeAttempt, error)
	GetFailedAttempts(ctx context.Context, userID string, minutes int) ([]ChallengeAttempt, error)
}

type Writer interface {
	CreateChallengeAttempt(ctx context.Context, challengeAttempt ChallengeAttempt) error
}

type SQLReader struct {
	cache *cache.Cache[[]ChallengeAttempt]
	q     *gen.Queries
	l     *zap.Logger
}

func NewSQLReader(q *gen.Queries, cache *cache.Cache[[]ChallengeAttempt]) *SQLReader {
	return &SQLReader{q: q, cache: cache}
}

func (r *SQLReader) GetChallengeAttempts(ctx context.Context, userID uuid.UUID) ([]ChallengeAttempt, error) {
	cachedAttempts, err := r.cache.Get(ctx, userID.String())
	if err == nil {
		return *cachedAttempts, nil
	}

	attempts, err := r.q.GetChallengeAttempts(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = r.cache.Set(ctx, userID.String(), attempts)
	if err != nil {
		r.l.Warn("failed to set cache", zap.Error(err))
	}

	return attempts, nil
}

func (r *SQLReader) GetFailedAttempts(ctx context.Context, userID uuid.UUID, minutes int) ([]ChallengeAttempt, error) {
	return r.q.GetFailedAttempts(ctx, gen.GetFailedAttemptsParams{
		UserID:  userID,
		Minutes: sql.NullString{String: strconv.Itoa(minutes), Valid: true},
	})
}

type SQLWriter struct {
	q *gen.Queries
	l *zap.Logger
}

func NewSQLWriter(q *gen.Queries) *SQLWriter {
	return &SQLWriter{q: q, l: log.NewLogger("challenge_attempts_writer")}
}

func (w *SQLWriter) CreateChallengeAttempt(ctx context.Context, userID uuid.UUID, challengeID uuid.UUID, success bool) error {
	return w.q.CreateChallengeAttempt(ctx, gen.CreateChallengeAttemptParams{
		ChallengeID: challengeID,
		UserID:      userID,
		Success:     success,
	})
}
