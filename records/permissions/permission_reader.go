package permissions

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/cache"
	authv1 "github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/gen"
	"go.uber.org/zap"
)

type Permission = authv1.Permission
type PermissionId = authv1.PermissionId
type RoleId = authv1.RoleId

var _ PermissionReader = &SQLPermissionReader{}

//go:generate go run github.com/golang/mock/mockgen -source=permission_reader.go -destination=mocks/mock_permission_reader.go -package=mocks -mock_names=PermissionReader=MockPermissionReader
type PermissionReader interface {
	GetPermission(ctx context.Context, id PermissionId) (*Permission, error)
	GetPermissions(ctx context.Context, limit, offset int32) ([]Permission, error)
	GetPermissionsForRole(ctx context.Context, roleId RoleId) ([]Permission, error)
	GetPermissionsByFilter(ctx context.Context, group, kind, name string, limit, offset int32) ([]Permission, error)
}

func NewSQLPermissionReader(cache *cache.Cache[[]Permission], l *zap.Logger, q *gen.Queries) *SQLPermissionReader {
	return &SQLPermissionReader{
		cache: cache,
		l:     l,
		q:     q,
	}
}

type SQLPermissionReader struct {
	l     *zap.Logger
	cache *cache.Cache[[]Permission]
	q     *gen.Queries
}

func (r *SQLPermissionReader) GetPermission(ctx context.Context, id PermissionId) (*Permission, error) {

	cacheKey := fmt.Sprintf("permission:%s", id)

	if r.cache != nil {
		permissions, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return &permissions[0], nil
		}
	}

	permission, err := r.q.GetPermissionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p := &Permission{
		ID:            permission.ID,
		ResourceKind:  permission.ResourceKind,
		ResourceGroup: permission.ResourceGroup,
		ResourceName:  permission.ResourceName,
		Actions:       permission.Actions,
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, []Permission{*p})
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return p, nil
}

func (r *SQLPermissionReader) GetPermissions(ctx context.Context, limit, offset int32) ([]Permission, error) {

	cacheKey := fmt.Sprintf("permissions:%d:%d", limit, offset)

	if r.cache != nil {
		permissions, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return permissions, nil
		}
	}

	permissions, err := r.q.GetPermissions(ctx, gen.GetPermissionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	var result []Permission
	for _, p := range permissions {
		result = append(result, Permission{
			ID:            p.ID,
			ResourceKind:  p.ResourceKind,
			ResourceGroup: p.ResourceGroup,
			ResourceName:  p.ResourceName,
			Actions:       p.Actions,
		})
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, result)
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return result, nil
}

func (r *SQLPermissionReader) GetPermissionsForRole(ctx context.Context, roleId RoleId) ([]Permission, error) {

	cacheKey := fmt.Sprintf("permissions_for_role:%s", roleId)

	if r.cache != nil {
		permissions, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return permissions, nil
		}
	}

	permissions, err := r.q.GetPermissionsByRoleID(ctx, gen.GetPermissionsByRoleIDParams{
		RoleID: roleId,
	})
	if err != nil {
		return nil, err
	}

	var result []Permission
	for _, p := range permissions {
		result = append(result, Permission{
			ID:            p.ID,
			ResourceKind:  p.ResourceKind,
			ResourceGroup: p.ResourceGroup,
			ResourceName:  p.ResourceName,
			Actions:       p.Actions,
		})
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, result)
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return result, nil
}

func (r *SQLPermissionReader) GetPermissionsByFilter(ctx context.Context, group, kind, name string, limit, offset int32) ([]Permission, error) {

	cacheKey := fmt.Sprintf("permissions_by_filter:%s:%s:%s:%d:%d", group, kind, name, limit, offset)

	if r.cache != nil {
		permissions, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return permissions, nil
		}
	}

	permissions, err := r.q.GetPermissionsByFilter(ctx, gen.GetPermissionsByFilterParams{
		ResourceGroup: group,
		ResourceKind:  kind,
		ResourceName:  name,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}

	var result []Permission
	for _, p := range permissions {
		result = append(result, Permission{
			ID:            p.ID,
			ResourceKind:  p.ResourceKind,
			ResourceGroup: p.ResourceGroup,
			ResourceName:  p.ResourceName,
			Actions:       p.Actions,
		})
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, result)
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return result, nil
}
