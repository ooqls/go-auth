package permissionsv1

import (
	"context"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

var _ Writer = &SQLWriter{}

//go:generate go run go.uber.org/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks
type Writer interface {
	CreatePermission(ctx context.Context, permission string) (*Permission, error)
	DeletePermission(ctx context.Context, permission string) error
}

func NewSQLWriter(cache *cache.GenericCache, q *datagen.Queries) *SQLWriter {
	return &SQLWriter{cache: cache, q: q}
}

type SQLWriter struct {
	cache *cache.GenericCache
	q     *datagen.Queries
}

func (w *SQLWriter) CreatePermission(ctx context.Context, permission string) (*Permission, error) {
	row, err := w.q.CreatePermission(ctx, permission)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"permission": permission})
	}
	w.invalidate(ctx, cacheKeyAllPermissions)
	return fromDatagenPermission(row), nil
}

func (w *SQLWriter) GetOrCreatePermission(ctx context.Context, permission string) (*Permission, error) {
	row, err := w.q.GetOrCreatePermission(ctx, permission)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"permission": permission})
	}
	w.invalidate(ctx, cacheKeyAllPermissions)
	return fromDatagenPermission(row), nil
}

func (w *SQLWriter) DeletePermission(ctx context.Context, permission string) error {
	err := w.q.DeletePermission(ctx, permission)
	if err != nil {
		return v1.ErrInternal(err, v1.M{"permission": permission})
	}
	w.invalidate(ctx, cacheKeyAllPermissions, permissionCacheKey(permission))
	return nil
}

func (w *SQLWriter) invalidate(ctx context.Context, keys ...string) {
	if w.cache == nil {
		return
	}

	for _, key := range keys {
		if err := w.cache.Delete(ctx, key); err != nil {
			zap.L().Error("failed to invalidate permission cache", zap.String("key", key), zap.Error(err))
		}
	}
}
