package permissionbindingsv1

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
	"go.uber.org/zap"
)

type Reader interface {
	GetPermissionBindingsForRole(ctx contexts.LContext, roleId uuid.UUID) ([]Permissionbindingv1, error)
	GetPermissionsBindings(ctx contexts.LContext, page, pageSize int) ([]Permissionbindingv1, error)
}

func NewSQLReader(cache *cache.GenericCache, q datagen.Queries) Reader {
	return &SQLReader{
		cache: cache,
		q:     q,
	}
}

type SQLReader struct {
	cache *cache.GenericCache
	q     datagen.Queries
}

func (r *SQLReader) GetPermissionBindingsForRole(ctx contexts.LContext, roleId uuid.UUID) ([]Permissionbindingv1, error) {
	cachKey := roleId.String()
	if bindings := r.getCache(ctx, cachKey); bindings != nil {
		return bindings, nil
	}

	var bindings []Permissionbindingv1
	items, err := r.q.GetPermissionsForRole(ctx, roleId)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		bindings = append(bindings, Permissionbindingv1{
			Metadata:     Metadata,
			RoleID:       item.RoleID,
			Permission: item.Permission,
		})
	}

	r.setCache(ctx, cachKey, bindings)

	return bindings, nil
}

func (r *SQLReader) GetPermissionsBindings(ctx contexts.LContext, page, pageSize int) ([]Permissionbindingv1, error) {
	cacheKey := fmt.Sprintf("permission_bindings:%d:%d", page, pageSize)
	if bindings := r.getCache(ctx, cacheKey); bindings != nil {
		return bindings, nil
	}

	items, err := r.q.GetPermissions(ctx, datagen.GetPermissionsParams{
		Limit:  int32(pageSize),
		Offset: int32((page - 1) * pageSize),
	})
	if err != nil {
		return nil, err
	}

	var bindings []Permissionbindingv1
	for _, item := range items {
		bindings = append(bindings, Permissionbindingv1{
			Metadata:     Metadata,
			RoleID:       item.RoleID,
			Permission: item.Permission,
		})
	}

	r.setCache(ctx, cacheKey, bindings)
	return bindings, nil
}

func (r *SQLReader) getCache(ctx contexts.LContext, key string) []Permissionbindingv1 {
	if r.cache == nil {
		return nil
	}

	var bindings []Permissionbindingv1
	err := r.cache.Get(ctx, key, &bindings)
	if err != nil {
		if !cache.IsCacheMissErr(err) {
			ctx.L().Warn("failed to get cache from reader", zap.Error(err))
		}
		return nil
	}

	return bindings
}

func (r *SQLReader) setCache(ctx contexts.LContext, key string, bindings []Permissionbindingv1) {
	if r.cache != nil {
		err := r.cache.Set(ctx, key, bindings)
		if err != nil && !cache.IsCacheMissErr(err) {
			ctx.L().Warn("failed to set reader cache", zap.Error(err))
		}
	}
}
