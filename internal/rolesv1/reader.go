package rolesv1

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ Reader = &SQLRoleReader{}

//go:generate go run go.uber.org/mock/mockgen -source=reader.go -destination=mocks/mock_role_reader.go -package=mocks -mock_names=RoleReader=MockRoleReader
type Reader interface {
	GetRole(ctx contexts.LContext, id uuid.UUID) (*Role, error)
	GetRoles(ctx contexts.LContext, limit, offset int32) (*corev1.Result[[]Role], error)
	// GetRolesForUser(ctx contexts.LContext, userId uuid.UUID) ([]datagen.Rolesv1, error)
	GetRoleByName(ctx contexts.LContext, name string) (*Role, error)
	ClearCache(ctx context.Context) error
}

func NewSQLRoleReader(cache cache.GenericCache, db *datagen.Queries) *SQLRoleReader {
	return &SQLRoleReader{
		cache: &cache,
		q:     db,
	}
}

type SQLRoleReader struct {
	cache *cache.GenericCache
	q     *datagen.Queries
}

func (r *SQLRoleReader) GetRole(ctx contexts.LContext, id uuid.UUID) (*Role, error) {
	cacheKey := fmt.Sprintf("role:%s", id.String())

	cachedRole := r.getCache(ctx, cacheKey)
	if cachedRole != nil {
		return &cachedRole.Items[0], nil
	}

	qRole, err := r.q.GetRole(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get role")
	}

	role := FromDatagenRole(qRole)

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, corev1.Result[[]Role]{Items: []Role{role}})
		if err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &role, nil
}

func (r *SQLRoleReader) GetRoles(ctx contexts.LContext, limit, offset int32) (*corev1.Result[[]Role], error) {
	cacheKey := fmt.Sprintf("roles:%d:%d", limit, offset)

	cachedRoles := r.getCache(ctx, cacheKey)
	if cachedRoles != nil {
		return cachedRoles, nil
	}

	qroles, err := r.q.GetRoles(ctx, datagen.GetRolesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	var roles []Role
	for _, role := range qroles {
		roles = append(roles, FromDatagenRole(role))
	}

	count, err := r.q.CountRoles(ctx)
	if err != nil {
		ctx.L().Warn("failed to count roles")
		count = -1
	}

	res := &corev1.Result[[]Role]{Items: roles, TotalCount: count}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, res)
		if err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return res, nil
}

// func (r *SQLRoleReader) GetRolesForUser(ctx context.LContext, userId uuid.UUID) ([]datagen.Rolesv1, error) {
// 	cacheKey := fmt.Sprintf("roles_for_user:%s", userId)

// 	cachedRoles := r.getCache(ctx, cacheKey)
// 	if cachedRoles != nil {
// 		return cachedRoles, nil
// 	}

// 	userRoles, err := r.q.GetRolesForUser(ctx, datagen.GetRolesForUserParams{
// 		UserID: userId,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	var roles []Role
// 	for _, role := range userRoles {
// 		roles = append(roles, Role{
// 			ID:          role.ID,
// 			Name:        role.Name,
// 			Description: role.Description,
// 			UpdatedAt:   role.UpdatedAt,
// 			CreatedAt:   role.CreatedAt,
// 		})
// 	}

// 	if r.cache != nil {
// 		err = r.cache.Set(ctx, cacheKey, roles)
// 		if err != nil {
// 			ctx.L().Error("failed to set cache", zap.Error(err))
// 		}
// 	}

// 	return roles, nil
// }

func (r *SQLRoleReader) GetRoleByName(ctx contexts.LContext, name string) (*Role, error) {
	cacheKey := fmt.Sprintf("roles_by_name:%s", name)

	cachedRoles := r.getCache(ctx, cacheKey)
	if cachedRoles != nil {
		return &cachedRoles.Items[0], nil
	}

	qrole, err := r.q.GetRolesByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	role := FromDatagenRole(qrole)
	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, []Role{role})
		if err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return &role, nil
}

func (r *SQLRoleReader) getCache(ctx contexts.LContext, keys ...string) *corev1.Result[[]Role] {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		var roles corev1.Result[[]Role]
		err := r.cache.Get(ctx, key, &roles)
		if err == nil && len(roles.Items) > 0 {
			return &roles
		}

		if err != nil && !cache.IsCacheMissErr(err) {
			ctx.L().Error("failed to get cache", zap.Error(err))
		}
	}

	return nil
}

func (r *SQLRoleReader) ClearCache(ctx context.Context) error {
	return r.cache.Clear(ctx)
}
