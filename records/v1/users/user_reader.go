package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/ooqls/go-auth/records/v1/gen"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-log"
	"go.uber.org/zap"
)

var _ Reader = &SQLReader{}

//go:generate go run github.com/golang/mock/mockgen -source=user_reader.go -destination=mocks/mock_user_reader.go -package=mocks -mock_names=UserReader=MockUserReader
type Reader interface {
	GetUser(ctx context.Context, id UserId) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUsers(ctx context.Context, offset, limit int32) ([]User, error)
}

type SQLReader struct {
	cache *cache.Cache[[]User]
	l     *zap.Logger
	q     *gen.Queries
}

func NewSQLUserReader(cache *cache.Cache[[]User], db *sqlx.DB) *SQLReader {
	return &SQLReader{
		cache: cache,
		l:     log.NewLogger("user_reader"),
		q:     gen.New(db),
	}
}

func (r *SQLReader) addCache(ctx context.Context, user *User) error {
	if r.cache == nil {
		return nil
	}

	err := r.cache.Set(ctx, user.Username, []User{*user})
	if err != nil {
		r.l.Error("failed to set cache", zap.Error(err))
	}

	err = r.cache.Set(ctx, user.ID.String(), []User{*user})
	if err != nil {
		r.l.Error("failed to set cache", zap.Error(err))
	}

	return nil
}

func (r *SQLReader) getCache(ctx context.Context, keys ...string) []User {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		user, err := r.cache.Get(ctx, key)
		if err == nil && user != nil {
			return *user
		}

		if err != nil && !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get cache", zap.Error(err))
		}
	}

	return nil
}

func (r *SQLReader) GetUserByUsername(ctx context.Context, username string) (*User, error) {
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

	err = r.addCache(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *SQLReader) GetUser(ctx context.Context, id UserId) (*User, error) {

	cachedUser := r.getCache(ctx, id.String())
	if cachedUser != nil {
		return &cachedUser[0], nil
	}

	user, err := r.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.addCache(ctx, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *SQLReader) GetUsers(ctx context.Context, offset, limit int32) ([]User, error) {
	cacheKey := fmt.Sprintf("users:%d:%d", offset, limit)

	cachedUsers := r.getCache(ctx, cacheKey)
	if cachedUsers != nil {
		return cachedUsers, nil
	}

	users, err := r.q.ListUsers(ctx, gen.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		err = r.cache.Set(ctx, cacheKey, users)
		if err != nil && r.l != nil {
			r.l.Error("failed to set cache", zap.Error(err))
		}

		for _, user := range users {
			err = r.addCache(ctx, &user)
			if err != nil {
				return nil, err
			}
		}
	}

	return users, nil
}
