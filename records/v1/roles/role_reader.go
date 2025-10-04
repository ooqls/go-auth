package roles

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/gen"
	"go.uber.org/zap"
)

var _ Reader = &SQLRoleReader{}

//go:generate go run github.com/golang/mock/mockgen -source=role_reader.go -destination=mocks/mock_role_reader.go -package=mocks -mock_names=RoleReader=MockRoleReader
type Reader interface {
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoles(ctx context.Context, limit, offset int32) ([]Role, error)
	GetRolesForUser(ctx context.Context, userId UserId) ([]Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
}

func NewSQLRoleReader(cache *cache.Cache[[]Role], l *zap.Logger, q *gen.Queries) *SQLRoleReader {
	return &SQLRoleReader{
		cache: cache,
		l:     l,
		q:     q,
	}
}

type SQLRoleReader struct {
	cache *cache.Cache[[]Role]
	l     *zap.Logger
	q     *gen.Queries
}

func (r *SQLRoleReader) GetRole(ctx context.Context, id uuid.UUID) (*Role, error) {
	cacheKey := fmt.Sprintf("role:%s", id)

	if r.cache != nil {
		role, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return &role[0], nil
		}
	}

	qRole, err := r.q.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	role := &Role{
		ID:          qRole.ID,
		RoleName:    qRole.RoleName,
		Description: qRole.Description,
		UpdatedAt:   qRole.UpdatedAt,
		CreatedAt:   qRole.CreatedAt,
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, []Role{*role})
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return role, nil
}

func (r *SQLRoleReader) GetRoles(ctx context.Context, limit, offset int32) ([]Role, error) {
	cacheKey := fmt.Sprintf("roles:%d:%d", limit, offset)

	if r.cache != nil {
		roles, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return roles, nil
		}
	}

	roles, err := r.q.GetRoles(ctx, gen.GetRolesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, []Role(roles))
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return []Role(roles), nil
}

func (r *SQLRoleReader) GetRolesForUser(ctx context.Context, userId UserId) ([]Role, error) {
	cacheKey := fmt.Sprintf("roles_for_user:%s", userId)

	if r.cache != nil {
		roles, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return roles, nil
		}
	}

	userRoles, err := r.q.GetRolesForUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	var roles []Role
	for _, role := range userRoles {
		roles = append(roles, Role{
			ID:          role.ID,
			RoleName:    role.RoleName,
			Description: role.Description,
			UpdatedAt:   role.UpdatedAt,
			CreatedAt:   role.CreatedAt,
		})
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, roles)
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return roles, nil
}

func (r *SQLRoleReader) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	cacheKey := fmt.Sprintf("roles_by_name:%s", name)

	if r.cache != nil {
		roles, err := r.cache.Get(ctx, cacheKey)
		if err == nil {
			return &roles[0], nil
		}
	}

	roles, err := r.q.GetRolesByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, nil
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, []Role(roles))
		if err != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}
	}

	return &roles[0], nil
}
