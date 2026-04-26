package permissionsv1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks
type Reader interface {
	GetPermissions(ctx contexts.LContext, page, pageSize int) ([]Permission, error)
	GetPermission(ctx contexts.LContext, permission string) (*Permission, error)
	HasPermission(ctx contexts.LContext, userID uuid.UUID, permission string) (bool, error)
	GetPermissionsForUser(ctx contexts.LContext, userID uuid.UUID) ([]Permission, error)
	GetPermissionsForUserByGroup(ctx contexts.LContext, userID uuid.UUID) ([]Permission, error)
	ClearCache(ctx context.Context) error
}

func NewSQLReader(cache *cache.GenericCache, q *datagen.Queries) *SQLReader {
	return &SQLReader{
		cache: cache,
		q:     q,
	}
}

type SQLReader struct {
	cache *cache.GenericCache
	q     *datagen.Queries
}

func (r *SQLReader) GetPermissions(ctx contexts.LContext, page, pageSize int) ([]Permission, error) {
	cacheKey := fmt.Sprintf("permissions:%d:%d", page, pageSize)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetPermissions(ctx, datagen.GetPermissionsParams{
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"page": page, "pageSize": pageSize})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   Metadata,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) GetPermission(ctx contexts.LContext, permission string) (*Permission, error) {
	cacheKey := fmt.Sprintf("permission:%s", permission)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &cached[0], nil
	}

	row, err := r.q.GetPermissionByName(ctx, permission)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, v1.ErrInternal(err, v1.M{"permission": permission})
	}

	p := Permission{
		Metadata:   Metadata,
		Permission: row,
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, []Permission{p}); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &p, nil
}

func (r *SQLReader) HasPermission(ctx contexts.LContext, userID uuid.UUID, permission string) (bool, error) {
	return r.q.HasPermission(ctx, datagen.HasPermissionParams{
		UserID:     userID,
		Permission: permission,
	})
}

func (r *SQLReader) GetPermissionsForUserByGroup(ctx contexts.LContext, userID uuid.UUID) ([]Permission, error) {
	cacheKey := fmt.Sprintf("permissions:user:%s:bygroup", userID)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetPermissionsForUserByGroup(ctx, userID)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   Metadata,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) GetPermissionsForUser(ctx contexts.LContext, userID uuid.UUID) ([]Permission, error) {
	cacheKey := fmt.Sprintf("permissions:user:%s", userID)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   Metadata,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) getCache(ctx contexts.LContext, keys ...string) []Permission {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		var permissions []Permission
		err := r.cache.Get(ctx, key, &permissions)
		if err == nil && len(permissions) > 0 {
			return permissions
		}

		if err != nil && !cache.IsCacheMissErr(err) {
			ctx.L().Error("failed to get cache", zap.Error(err))
		}
	}

	return nil
}

func (r *SQLReader) ClearCache(ctx context.Context) error {
	r.cache.Clear(ctx)
	return nil
}
