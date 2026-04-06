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
	GetPermissions(ctx contexts.LContext, page int, pageSize int) ([]Permission, error)
	GetPermission(ctx contexts.LContext, name, group, kind string) (*Permission, error)
	HasPermission(ctx contexts.LContext, userID uuid.UUID, name, group, kind, actions string) (bool, error)
	GetPermissionForUserByGrouo(ctx contexts.LContext, userID uuid.UUID, group string) ([]Permission, error)
	GetPermissionForUser(ctx contexts.LContext, userID uuid.UUID, kind, group string) ([]Permission, error)
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

func (r *SQLReader) GetPermissions(ctx contexts.LContext, page int, pageSize int) ([]Permission, error) {
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
		perms = append(perms, *fromDatagenPermission(row))
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) GetPermission(ctx contexts.LContext, name, group, kind string) (*Permission, error) {
	cacheKey := fmt.Sprintf("permission:%s:%s:%s", name, group, kind)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &cached[0], nil
	}

	row, err := r.q.GetPermission(ctx, datagen.GetPermissionParams{
		Name:  name,
		Group: group,
		Kind:  kind,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, v1.ErrInternal(err, v1.M{"name": name, "group": group, "kind": kind})
	}

	p := fromDatagenPermission(row)
	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, []Permission{*p}); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return p, nil
}

func (r *SQLReader) HasPermission(ctx contexts.LContext, userID uuid.UUID, name, group, kind, actions string) (bool, error) {
	// Reuse GetPermissionForUser so the cache is shared; filter by name in Go.
	perms, err := r.q.HasPermission(ctx, datagen.HasPermissionParams{
		UserID:  userID,
		Name:    name,
		Group:   group,
		Kind:    kind,
		Actions: actions,
	})
	if err != nil {
		return false, err
	}

	return perms, nil
}

func (r *SQLReader) GetPermissionForUserByGrouo(ctx contexts.LContext, userID uuid.UUID, group string) ([]Permission, error) {
	cacheKey := fmt.Sprintf("permissions:user:%s:group:%s", userID, group)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetPermissionsForUserByGroup(ctx, datagen.GetPermissionsForUserByGroupParams{
		UserID: userID,
		Group:  group,
	})
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID, "group": group})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, fromUserPermRow(row.PermissionID, row.PermissionName, row.PermissionGroup, row.PermissionKind, row.Actions))
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) GetPermissionForUser(ctx contexts.LContext, userID uuid.UUID, kind, group string) ([]Permission, error) {
	cacheKey := fmt.Sprintf("permissions:user:%s:group:%s:kind:%s", userID, group, kind)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetPermissionsForUser(ctx, datagen.GetPermissionsForUserParams{
		UserID: userID,
		Group:  group,
		Kind:   kind,
	})
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID, "group": group, "kind": kind})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, fromUserPermRow(row.PermissionID, row.PermissionName, row.PermissionGroup, row.PermissionKind, row.Actions))
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
	// This is a simple implementation that clears all permissions-related cache keys.
	// In a real implementation, you might want to be more selective or use a cache that supports key patterns.
	r.cache.Clear(ctx)
	return nil
}
