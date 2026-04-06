package usersv1

// import (
// 	"database/sql"
// 	"errors"
// 	"fmt"

// 	"github.com/jmoiron/sqlx"
// 	"github.com/ooqls/go-auth/common"
// 	"github.com/ooqls/go-auth/common/context"
// 	"github.com/ooqls/go-auth/records/v1/gen"

// 	"github.com/ooqls/getset/cache/cache"
// 	"github.com/ooqls/go-auth/records"
// 	"github.com/ooqls/go-auth/v1/users/datagen"
// 	"go.uber.org/zap"
// )

// type AggReader interface {
// 	GetUserAgg(ctx common.LContext, id UserId) (*records.UserAgg, error)
// }

// type SQLAggReader struct {
// 	cache *cache.GenericCache
// 	q     *datagen.Queries
// }

// func NewSQLUserAggReader(cache cache.GenericCache, db sqlx.DB) *SQLAggReader {
// 	return &SQLAggReader{
// 		cache: &cache,
// 		q:     gen.New(db),
// 	}
// }

// func (r *SQLAggReader) GetUserAgg(ctx context.LContext, id records.UserId) (*records.UserAgg, error) {
// 	cacheKey := fmt.Sprintf("user_agg:%s", id.String())

// 	cachedUserAgg := r.getCache(ctx, cacheKey)
// 	if cachedUserAgg != nil {
// 		return &cachedUserAgg[0], nil
// 	}

// 	userRows, err := r.q.GetUserAgg(ctx, id)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}

// 	if len(userRows) == 0 {
// 		return nil, nil
// 	}

// 	roleMap := map[string]*records.RoleAgg{}
// 	for _, userRow := range userRows {
// 		role, ok := roleMap[userRow.RoleID.UUID.String()]
// 		if !ok {
// 			role = &records.RoleAgg{
// 				RoleId:      records.RoleId(userRow.RoleID.UUID),
// 				Hierarchy:   userRow.Hierarchy.Int32,
// 				Permissions: make([]records.Permission, 0),
// 			}
// 			roleMap[userRow.RoleID.UUID.String()] = role
// 		}
// 		role.Permissions = append(role.Permissions, records.Permission{
// 			ID:            userRow.ID.UUID,
// 			ResourceKind:  userRow.ResourceKind.String,
// 			ResourceGroup: userRow.ResourceGroup.String,
// 			ResourceName:  userRow.ResourceName.String,
// 			Actions:       userRow.Actions.String,
// 		})
// 	}

// 	roles := make([]records.RoleAgg, 0)
// 	for _, role := range roleMap {
// 		roles = append(roles, *role)
// 	}

// 	return &records.UserAgg{
// 		UserId:    records.UserId(userRows[0].UserID),
// 		Username:  userRows[0].Username,
// 		Email:     userRows[0].Email,
// 		CreatedAt: userRows[0].CreatedAt,
// 		UpdatedAt: userRows[0].UpdatedAt,
// 		Roles:     roles,
// 	}, nil
// }

// func (r *SQLAggReader) getCache(ctx context.LContext, keys ...string) []records.UserAgg {
// 	if r.cache == nil {
// 		return nil
// 	}

// 	for _, key := range keys {
// 		var userAgg []records.UserAgg
// 		err := r.cache.Get(ctx, key, &userAgg)
// 		if err == nil && len(userAgg) > 0 {
// 			return userAgg
// 		}

// 		if err != nil && !cache.IsCacheMissErr(err) {
// 			ctx.L().Error("failed to get cache", zap.Error(err))
// 		}
// 	}

// 	return nil
// }
