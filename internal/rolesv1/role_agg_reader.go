package rolesv1

// import (
// 	"database/sql"
// 	"errors"
// 	"fmt"

// 	"github.com/jmoiron/sqlx"
// 	"github.com/ooqls/getset/cache/cache"
// 	"github.com/ooqls/go-auth/common/context"
// 	authv1 "github.com/ooqls/go-auth/records"
// 	"github.com/ooqls/go-auth/records/v1/gen"
// 	"go.uber.org/zap"
// )

// type RoleAgg = authv1.RoleAgg
// type RoleId = authv1.RoleId
// type Role = authv1.Role
// type Permission = authv1.Permission
// type UserId = authv1.UserId

// var _ AggReader = &SQLAggReader{}

// type AggReader interface {
// 	GetRoleAggForUser(ctx context.LContext, id UserId) ([]RoleAgg, error)
// 	GetRoleAgg(ctx context.LContext, id RoleId) (*RoleAgg, error)
// }

// type SQLAggReader struct {
// 	cache *cache.GenericCache
// 	q     *gen.Queries
// }

// func NewSQLAggReader(cache cache.GenericCache, db sqlx.DB) *SQLAggReader {
// 	return &SQLAggReader{
// 		cache: &cache,
// 		q:     gen.New(db),
// 	}
// }

// func (r *SQLAggReader) GetRoleAggForUser(ctx context.LContext, id UserId) ([]RoleAgg, error) {
// 	roleAggs, err := r.q.GetRoleAggregate(ctx, id)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}

// 		return nil, err
// 	}

// 	roleMap := map[string]*RoleAgg{}
// 	roles := make([]RoleAgg, 0)

// 	for _, r := range roleAggs {
// 		aggR, ok := roleMap[r.RoleID.UUID.String()]
// 		if !ok {
// 			newAggRole := &RoleAgg{
// 				RoleId:    RoleId(r.RoleID.UUID),
// 				Hierarchy: int32(r.Hierarchy),
// 				Permissions: []Permission{
// 					{
// 						ID:            r.ID.UUID,
// 						ResourceKind:  r.ResourceKind.String,
// 						ResourceGroup: r.ResourceGroup.String,
// 						ResourceName:  r.ResourceName.String,
// 						Actions:       r.Actions.String,
// 						CreatedAt:     r.CreatedAt.Time,
// 						UpdatedAt:     r.UpdatedAt.Time,
// 					},
// 				},
// 			}
// 			roleMap[r.RoleID.UUID.String()] = newAggRole
// 			roles = append(roles, *newAggRole)
// 		} else {
// 			aggR.Permissions = append(aggR.Permissions, Permission{
// 				ID:            r.ID.UUID,
// 				ResourceKind:  r.ResourceKind.String,
// 				ResourceGroup: r.ResourceGroup.String,
// 				ResourceName:  r.ResourceName.String,
// 				Actions:       r.Actions.String,
// 				CreatedAt:     r.CreatedAt.Time,
// 				UpdatedAt:     r.UpdatedAt.Time,
// 			})
// 		}
// 	}
// 	return roles, nil
// }

// func (r *SQLAggReader) GetRoleAgg(ctx context.LContext, id RoleId) (*RoleAgg, error) {
// 	cacheKey := fmt.Sprintf("role_agg:%s", id.String())

// 	cachedRoleAgg := r.getCache(ctx, cacheKey)
// 	if cachedRoleAgg != nil {
// 		return &cachedRoleAgg[0], nil
// 	}

// 	roleRows, err := r.q.GetRoleAgg(ctx, id)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}

// 	if len(roleRows) == 0 {
// 		return nil, nil
// 	}

// 	permissions := make([]Permission, 0)
// 	for _, roleRow := range roleRows {
// 		permissions = append(permissions, Permission{
// 			ID:            roleRow.ID.UUID,
// 			ResourceKind:  roleRow.ResourceKind.String,
// 			ResourceGroup: roleRow.ResourceGroup.String,
// 			ResourceName:  roleRow.ResourceName.String,
// 			Actions:       roleRow.Actions.String,
// 		})
// 	}

// 	return &RoleAgg{
// 		RoleId:      RoleId(id),
// 		Hierarchy:   int32(roleRows[0].Hierarchy),
// 		Permissions: permissions,
// 	}, nil
// }

// func (r *SQLAggReader) getCache(ctx context.LContext, keys ...string) []RoleAgg {
// 	if r.cache == nil {
// 		return nil
// 	}

// 	for _, key := range keys {
// 		var roleAgg []RoleAgg
// 		err := r.cache.Get(ctx, key, &roleAgg)
// 		if err == nil && len(roleAgg) > 0 {
// 			return roleAgg
// 		}

// 		if err != nil && !cache.IsCacheMissErr(err) {
// 			ctx.L().Error("failed to get cache", zap.Error(err))
// 		}
// 	}

// 	return nil
// }
