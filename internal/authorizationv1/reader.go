package authorizationv1

//go:generate go run go.uber.org/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authorizationv1/datagen"
	"go.uber.org/zap"
)

type Reader interface {
	CompareUserHierarchy(ctx context.Context, userA, userB uuid.UUID) (int32, error)
}

type SQLReader struct {
	l *zap.Logger
	c *cache.GenericCache
	q *datagen.Queries
}

func NewSQLReader(q *datagen.Queries, c *cache.GenericCache) *SQLReader {
	return &SQLReader{
		q: q,
		c: c,
		l: log.NewLogger("authorization_reader"),
	}
}

func (r *SQLReader) CompareUserHierarchy(ctx context.Context, userA, userB uuid.UUID) (int32, error) {
	cacheKey := fmt.Sprintf("compare_hierarchy:%s:%s", userA, userB)

	if r.c != nil {
		var result int32
		if err := r.c.Get(ctx, cacheKey, &result); err == nil {
			return result, nil
		} else if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get cache", zap.Error(err))
		}
	}

	result, err := r.q.CompareUserRoleHierarchy(ctx, datagen.CompareUserRoleHierarchyParams{
		UserID:   userA,
		UserID_2: userB,
	})
	if err != nil {
		return 0, err
	}

	if r.c != nil {
		if err := r.c.Set(ctx, cacheKey, result); err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return result, nil
}
