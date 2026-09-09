package permissionsv1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

// cacheKeyAllPermissions caches the full permission set as a single entry so
// pagination happens in-memory and there is one key to invalidate on write,
// instead of one entry per distinct (page, pageSize) combination.
const cacheKeyAllPermissions = "permissions:all"

func permissionCacheKey(permission string) string {
	return fmt.Sprintf("permission:%s", permission)
}

//go:generate go run go.uber.org/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks
type Reader interface {
	GetPermissions(ctx contexts.LContext, page, pageSize int) (*corev1.Result[[]Permission], error)
	GetPermission(ctx contexts.LContext, permission string) (*Permission, error)
	SearchPermissions(ctx contexts.LContext, permission string, page, pageSize int) (*corev1.Result[[]Permission], error)
	HasCorePermission(ctx contexts.LContext, userID uuid.UUID, group, kind, action string) (bool, error)
	HasResourcePermission(ctx contexts.LContext, userID uuid.UUID, group, kind, action string) (bool, error)

	GetPermissionsForUser(ctx contexts.LContext, userID uuid.UUID) (*corev1.Result[[]Permission], error)
	GetPermissionsForUserByGroup(ctx contexts.LContext, userID uuid.UUID) (*corev1.Result[[]Permission], error)
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

func (r *SQLReader) SearchPermissions(ctx contexts.LContext, permission string, page int, pageSize int) (*corev1.Result[[]Permission], error) {

	page = max(page, 1)
	pageSize = max(pageSize, 1)

	offset := (page - 1) * pageSize

	items, err := r.q.SearchPermissions(ctx, datagen.SearchPermissionsParams{
		Permission: "%" + permission + "%",
		Limit:      int32(pageSize), // or any default limit you want to set
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{})
	}

	count, err := r.q.CountSearchPermissions(ctx, "%"+permission+"%")
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{})
	}

	var perms []Permission
	for _, p := range items {
		perms = append(perms, Permission{
			Metadata:   corev1.PermissionsV1,
			Permission: p,
		})
	}
	return &corev1.Result[[]Permission]{Items: perms, TotalCount: count}, nil
}

func (r *SQLReader) GetPermissions(ctx contexts.LContext, page, pageSize int) (*corev1.Result[[]Permission], error) {
	perms, err := r.getAllPermissions(ctx)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{})
	}

	total := len(perms)
	start := min(max((page-1)*pageSize, 0), total)
	end := min(start+pageSize, total)
	var permArr []Permission
	for _, p := range perms {
		permArr = append(permArr, p)
	}
	return &corev1.Result[[]Permission]{Items: permArr[start:end], TotalCount: int64(total)}, nil
}

func (r *SQLReader) getAllPermissions(ctx contexts.LContext) ([]Permission, error) {
	if cached := r.getCache(ctx, cacheKeyAllPermissions); cached != nil {
		return cached, nil
	}

	rows, err := r.q.GetAllPermissions(ctx)
	if err != nil {
		return nil, v1.ErrInternal(err, nil)
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   corev1.PermissionsV1,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKeyAllPermissions, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return perms, nil
}

func (r *SQLReader) GetPermission(ctx contexts.LContext, permission string) (*Permission, error) {
	cacheKey := permissionCacheKey(permission)

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
		Metadata:   corev1.PermissionsV1,
		Permission: row,
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, []Permission{p}); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &p, nil
}

func (r *SQLReader) HasResourcePermission(ctx contexts.LContext, userID uuid.UUID, group, kind, action string) (bool, error) {
	return r.q.HasResourcePermission(ctx, datagen.HasResourcePermissionParams{
		UserID: userID,
		Rgroup: group,
		Kind:   kind,
		Action: action,
	})
}

func (r *SQLReader) HasCorePermission(ctx contexts.LContext, userID uuid.UUID, group, kind, action string) (bool, error) {
	return r.q.HasCorePermission(ctx, datagen.HasCorePermissionParams{
		UserID: userID,
		Rgroup: group,
		Kind:   kind,
		Action: action,
	})
}

func (r *SQLReader) GetPermissionsForUserByGroup(ctx contexts.LContext, userID uuid.UUID) (*corev1.Result[[]Permission], error) {
	cacheKey := fmt.Sprintf("permissions:user:%s:bygroup", userID)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &corev1.Result[[]Permission]{Items: cached, TotalCount: int64(len(cached))}, nil
	}

	rows, err := r.q.GetPermissionsForUserByGroup(ctx, userID)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   corev1.PermissionsV1,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &corev1.Result[[]Permission]{Items: perms, TotalCount: int64(len(perms))}, nil
}

func (r *SQLReader) GetPermissionsForUser(ctx contexts.LContext, userID uuid.UUID) (*corev1.Result[[]Permission], error) {
	cacheKey := fmt.Sprintf("permissions:user:%s", userID)

	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &corev1.Result[[]Permission]{Items: cached, TotalCount: int64(len(cached))}, nil
	}

	rows, err := r.q.GetPermissionsForUser(ctx, userID)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"userID": userID})
	}

	var perms []Permission
	for _, row := range rows {
		perms = append(perms, Permission{
			Metadata:   corev1.PermissionsV1,
			Permission: row,
		})
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, perms); err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &corev1.Result[[]Permission]{Items: perms, TotalCount: int64(len(perms))}, nil
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
