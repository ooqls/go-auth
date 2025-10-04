package actions

import (
	"context"
	"errors"

	"github.com/eko/gocache/lib/v4/cache"
	authv1 "github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/gen"
	"github.com/ooqls/go-log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UserId = authv1.UserId
type Group = authv1.Group
type Kind = authv1.Kind

//go:generate go run github.com/golang/mock/mockgen -source=actions_reader.go -destination=mocks/mock_actions_reader.go -package=mocks -mock_names=ActionsReader=MockActionsReader
type Reader interface {
	GetActionsForUserByResource(ctx context.Context, id UserId, group Group, kind Kind, name string) ([]string, error)
}

type SQLReader struct {
	cache *cache.Cache[[]string]
	q     *gen.Queries
	l     *zap.Logger
}

func NewSQLReader(q *gen.Queries, cache *cache.Cache[[]string]) *SQLReader {
	return &SQLReader{
		q:     q,
		cache: cache,
		l:     log.NewLogger("ActionsReaderImpl"),
	}
}

func (r *SQLReader) GetActionsForUserByResource(ctx context.Context, id UserId, group Group, kind Kind, name string) ([]string, error) {
	if r.cache != nil {
		cachedActions, err := r.cache.Get(ctx, id.String())
		if err == nil && len(cachedActions) > 0 {
			return cachedActions, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			r.l.Warn("something went wrong when accessing cache", zap.Error(err))
		}
		// Cache miss, fetch from the database
		actions, err := r.q.GetActionsForUserByResource(ctx, gen.GetActionsForUserByResourceParams{
			UserID:        id,
			ResourceGroup: group,
			ResourceKind:  kind,
			ResourceName:  name,
		})
		if err != nil {
			return nil, err
		}

		err = r.cache.Set(ctx, id.String(), actions)
		if err != nil {
			r.l.Warn("failed to set actions in cache", zap.Error(err))
		}

		return actions, nil
	}

	return r.q.GetActionsForUserByResource(ctx, gen.GetActionsForUserByResourceParams{
		UserID:        id,
		ResourceGroup: group,
		ResourceKind:  kind,
		ResourceName:  name,
	})
}
