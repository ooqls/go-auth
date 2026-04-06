package aggsv1

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/aggsv1/datagen"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

type Reader interface {
	GetRoleAgg(ctx context.Context, roleID uuid.UUID) (*RoleAgg, error)
	GetUserAgg(ctx context.Context, userID uuid.UUID) (*UserAgg, error)
	CompareUserHierarchy(ctx context.Context, userA, userB uuid.UUID) (int32, error)
	GetUserHierarchy(ctx context.Context, userID uuid.UUID) (int32, error)
}

func NewSQLReader(q *datagen.Queries, c *cache.GenericCache) Reader {
	return &SQLReader{
		q:     q,
		cache: c,
		l:     log.NewLogger("agg_reader"),
	}
}

type SQLReader struct {
	q     *datagen.Queries
	cache *cache.GenericCache
	l     *zap.Logger
}

func (r *SQLReader) GetRoleAgg(ctx context.Context, roleID uuid.UUID) (*RoleAgg, error) {
	cacheKey := fmt.Sprintf("role_agg:%s", roleID.String())

	if r.cache != nil {
		var cached RoleAgg
		if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		} else if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get role agg from cache", zap.Error(err))
		}
	}

	rows, err := r.q.GetRoleAggregate(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	first := rows[0]
	agg := &RoleAgg{
		Role: rolesv1.Role{
			Object: corev1.Object{
				Metadata: corev1.Metadata{
					Group: "roles",
					Kind:  rolesv1.Kind,
				},
				Id:   first.RoleID,
				Name: first.RoleName,
			},
			Description: first.Description,
			Hierarchy:   first.Hierarchy,
		},
	}

	for _, row := range rows {
		perm := permissionsv1.Permission{
			Object: corev1.Object{
				Metadata:  permissionsv1.Metadata,
				Id:        row.PermissionID,
				Name:      row.PermissionName,
				CreatedAt: row.PermissionCreatedAt.Time,
				UpdatedAt: row.PermissionUpdatedAt.Time,
			},
			Resource: corev1.Object{
				Metadata: corev1.Metadata{
					Group: row.PermissionGroup,
					Kind:  row.PermissionKind,
				},
				Name: row.PermissionName,
			},
			Actions: row.Actions,
		}
		agg.Permissions = append(agg.Permissions, perm)
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, agg); err != nil {
			r.l.Error("failed to set role agg in cache", zap.Error(err))
		}
	}

	return agg, nil
}

func (r *SQLReader) GetUserAgg(ctx context.Context, userID uuid.UUID) (*UserAgg, error) {
	cacheKey := fmt.Sprintf("user_agg:%s", userID.String())

	if r.cache != nil {
		var cached UserAgg
		if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		} else if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get user agg from cache", zap.Error(err))
		}
	}

	rows, err := r.q.GetUserAggregate(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}

	first := rows[0]
	agg := &UserAgg{
		User: usersv1.User{
			Object: corev1.Object{
				Metadata: usersv1.Metadata,
				Id:       first.UserID,
				Name:     first.Username,
			},
			Username: first.Username,
			Email:    first.Email,
		},
	}

	roleMap := map[uuid.UUID]*RoleAgg{}
	roleOrder := []uuid.UUID{}

	for _, row := range rows {
		if !row.RoleID.Valid {
			continue
		}

		roleID := uuid.UUID(row.RoleID.Bytes)
		role, ok := roleMap[roleID]
		if !ok {
			role = &RoleAgg{
				Role: rolesv1.Role{
					Object: corev1.Object{
						Metadata: corev1.Metadata{
							Group: "roles",
							Kind:  rolesv1.Kind,
						},
						Id:   roleID,
						Name: row.RoleName.String,
					},
					Description: row.RoleDescription.String,
					Hierarchy:   row.Hierarchy.Int32,
				},
			}
			roleMap[roleID] = role
			roleOrder = append(roleOrder, roleID)
		}

		if !row.PermissionID.Valid {
			continue
		}

		perm := permissionsv1.Permission{
			Object: corev1.Object{
				Metadata:  permissionsv1.Metadata,
				Id:        uuid.UUID(row.PermissionID.Bytes),
				Name:      row.PermissionName.String,
				CreatedAt: row.PermissionCreatedAt.Time,
				UpdatedAt: row.PermissionUpdatedAt.Time,
			},
			Resource: corev1.Object{
				Metadata: corev1.Metadata{
					Group: row.PermissionGroup.String,
					Kind:  row.PermissionKind.String,
				},
				Name: row.PermissionName.String,
			},
			Actions: row.Actions.String,
		}
		role.Permissions = append(role.Permissions, perm)
	}

	for _, id := range roleOrder {
		agg.Roles = append(agg.Roles, *roleMap[id])
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, agg); err != nil {
			r.l.Error("failed to set user agg in cache", zap.Error(err))
		}
	}

	return agg, nil
}

func (r *SQLReader) CompareUserHierarchy(ctx context.Context, userA, userB uuid.UUID) (int32, error) {
	cacheKey := fmt.Sprintf("compare_hierarchy:%s:%s", userA, userB)

	if r.cache != nil {
		var cached int32
		if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached, nil
		} else if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get hierarchy comparison from cache", zap.Error(err))
		}
	}

	result, err := r.q.CompareUserRoleHierarchy(ctx, datagen.CompareUserRoleHierarchyParams{
		UserID:   userA,
		UserID_2: userB,
	})
	if err != nil {
		return 0, err
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, result); err != nil {
			r.l.Error("failed to set hierarchy comparison in cache", zap.Error(err))
		}
	}

	return result, nil
}

func (r *SQLReader) GetUserHierarchy(ctx context.Context, userID uuid.UUID) (int32, error) {
	cacheKey := fmt.Sprintf("user_hierarchy:%s", userID)

	if r.cache != nil {
		var cached int32
		if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached, nil
		} else if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get user hierarchy from cache", zap.Error(err))
		}
	}

	result, err := r.q.GetUserHierarchy(ctx, userID)
	if err != nil {
		return 0, err
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey, result); err != nil {
			r.l.Error("failed to set user hierarchy in cache", zap.Error(err))
		}
	}

	return result, nil
}