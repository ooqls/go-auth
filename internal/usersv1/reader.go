package usersv1

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_user_reader.go -package=mocks -mock_names=UserReader=MockUserReader
type Reader interface {
	GetUser(ctx contexts.LContext, id uuid.UUID) (*User, error)
	GetUserByUsername(ctx contexts.LContext, username string) (*User, error)
	GetUsers(ctx contexts.LContext, offset, limit int32) ([]User, error)
	ClearCache(ctx contexts.LContext) error
}

type SQLReader struct {
	cache *cache.GenericCache
	q     *datagen.Queries
}

func NewSQLUserReader(cache cache.GenericCache, db *datagen.Queries) *SQLReader {
	return &SQLReader{
		cache: &cache,
		q:     db,
	}
}

func (r *SQLReader) addCache(ctx contexts.LContext, user *User) error {
	if r.cache == nil {
		return nil
	}

	err := r.cache.Set(ctx, user.Username, []User{*user})
	if err != nil {
		ctx.L().Error("failed to set cache", zap.Error(err))
	}

	err = r.cache.Set(ctx, user.Object.Id.String(), []User{*user})
	if err != nil {
		ctx.L().Error("failed to set cache", zap.Error(err))
	}

	return nil
}

func (r *SQLReader) getCache(ctx contexts.LContext, keys ...string) []User {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		var user []User
		err := r.cache.Get(ctx, key, &user)
		if err == nil && len(user) > 0 {
			return user
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
		return &cachedUser[0], nil
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
		return &cachedUsers[0], nil
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

func (r *SQLReader) GetUsers(ctx contexts.LContext, offset, limit int32) ([]User, error) {
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

	var userv1s []User
	for _, user := range users {
		userv1 := FromDatagenUser(user)
		userv1s = append(userv1s, userv1)
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, users)
		if err != nil {
			ctx.L().Error("failed to set cache", zap.Error(err))
		}
	}

	return userv1s, nil
}

func (r *SQLReader) ClearCache(ctx contexts.LContext) error {
	return r.cache.Clear(ctx)
}
