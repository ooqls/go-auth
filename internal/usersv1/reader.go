package usersv1

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

//go:generate go run go.uber.org/mock/mockgen -source=reader.go -destination=mocks/mock_user_reader.go -package=mocks -mock_names=UserReader=MockUserReader
type Reader interface {
	GetUser(ctx contexts.LContext, id uuid.UUID) (*User, error)
	GetUserByUsername(ctx contexts.LContext, username string) (*User, error)
	GetUsers(ctx contexts.LContext, offset, limit int32) (*corev1.Result[[]User], error)
	ClearCache(ctx contexts.LContext) error
}

type SQLReader struct {
	cache *cache.GenericCache
	q     *datagen.Queries
}

func NewSQLReader(cache cache.GenericCache, db *datagen.Queries) *SQLReader {
	return &SQLReader{
		cache: &cache,
		q:     db,
	}
}

func (r *SQLReader) addCache(ctx contexts.LContext, user *User) error {
	if r.cache == nil {
		return nil
	}

	err := r.cache.Set(ctx, user.Username, corev1.Result[[]User]{Items: []User{*user}})
	if err != nil {
		ctx.L().Error("failed to set cache", zap.Error(err))
	}

	err = r.cache.Set(ctx, user.Object.Id.String(), corev1.Result[[]User]{Items: []User{*user}})
	if err != nil {
		ctx.L().Error("failed to set cache", zap.Error(err))
	}

	return nil
}

func (r *SQLReader) getCache(ctx contexts.LContext, keys ...string) *corev1.Result[[]User] {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		var res corev1.Result[[]User]
		err := r.cache.Get(ctx, key, &res)
		if err == nil && len(res.Items) > 0 {
			return &res
		}

		if err != nil && !cache.IsCacheMissErr(err) {
			ctx.L().Error("failed to get cache", zap.Error(err))
		}
	}

	return nil
}

func (r *SQLReader) GetUserByUsername(ctx contexts.LContext, username string) (*User, error) {
	cacheKey := fmt.Sprintf("user_by_username:%s", username)

	cachedUser := r.getCache(ctx, cacheKey)
	if cachedUser != nil {
		return &cachedUser.Items[0], nil
	}

	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	userv1 := FromDatagenUser(user)

	err = r.addCache(ctx, &userv1)
	if err != nil {
		return nil, err
	}

	return &userv1, nil
}

func (r *SQLReader) GetUser(ctx contexts.LContext, id uuid.UUID) (*User, error) {

	cachedUsers := r.getCache(ctx, id.String())
	if cachedUsers != nil {
		return &cachedUsers.Items[0], nil
	}

	user, err := r.q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	userv1 := FromDatagenUser(user)
	err = r.addCache(ctx, &userv1)
	if err != nil {
		return nil, err
	}

	return &userv1, nil
}

func (r *SQLReader) GetUsers(ctx contexts.LContext, offset, limit int32) (*corev1.Result[[]User], error) {
	cacheKey := fmt.Sprintf("users:%d:%d", offset, limit)

	cachedUsers := r.getCache(ctx, cacheKey)
	if cachedUsers != nil {
		return cachedUsers, nil
	}

	users, err := r.q.ListUsers(ctx, datagen.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	count, err := r.q.CountUsers(ctx)
	if err != nil {
		ctx.L().Warn("failed to count users")
		count = -1
	}

	var userv1s []User
	for _, user := range users {
		userv1 := FromDatagenUser(datagen.Usersv1{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Key:       user.Key,
			Salt:      user.Salt,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
		userv1s = append(userv1s, userv1)
	}

	result := &corev1.Result[[]User]{Items: userv1s, TotalCount: count}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, result)
		if err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return result, nil
}

func (r *SQLReader) ClearCache(ctx contexts.LContext) error {
	return r.cache.Clear(ctx)
}
