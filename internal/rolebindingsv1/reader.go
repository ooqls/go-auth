package rolebindingsv1

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/rolebindingsv1/datagen"
	"go.uber.org/zap"
)

type Reader interface {
	GetRoleBindingsForUser(ctx contexts.LContext, userId uuid.UUID) ([]Rolebinding, error)
	HasRole(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) (bool, error)
	ClearCacheForUser(ctx contexts.LContext, userId uuid.UUID) error
}

type SQLReader struct {
	l *zap.Logger
	c *cache.GenericCache
	q *datagen.Queries
}

func NewSQLReader(c cache.GenericCache, db *datagen.Queries) *SQLReader {
	return &SQLReader{
		q: db,
		c: &c,
		l: log.NewLogger("role_bindings_reader"),
	}
}

func (r *SQLReader) ClearCacheForUser(ctx contexts.LContext, userId uuid.UUID) error {
	cacheKey := fmt.Sprintf("role_bindings_for_user:%s", userId.String())
	err := r.c.Delete(ctx, cacheKey)
	if err != nil {
		r.l.Error("failed to clear cache for user", zap.String("user_id", userId.String()), zap.Error(err))
		return err
	}

	return nil
}

func (r *SQLReader) GetRoleBindingsForUser(ctx contexts.LContext, userId uuid.UUID) ([]Rolebinding, error) {
	cacheKey := fmt.Sprintf("role_bindings_for_user:%s", userId.String())
	if bindings := r.getCache(ctx, cacheKey); bindings != nil {
		return bindings, nil
	}

	roleBindings, err := r.q.GetRoleBindingsForUser(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	var rbs []Rolebinding
	for _, rb := range roleBindings {
		rbs = append(rbs, fromDatagenRoleBinding(rb))
	}

	r.setCache(ctx, cacheKey, rbs...)
	return rbs, nil
}

func (r *SQLReader) HasRole(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) (bool, error) {
	return r.q.UserHasRole(ctx, datagen.UserHasRoleParams{
		UserID: userId,
		RoleID: roleId,
	})
}

func (r *SQLReader) setCache(ctx contexts.LContext, key string, value ...Rolebinding) {
	if r.c == nil {
		return
	}

	err := r.c.Set(ctx, key, value)
	if err != nil {
		ctx.L().Error("failed to set cache", zap.Error(err))
	}
}

func (r *SQLReader) getCache(ctx contexts.LContext, key string) []Rolebinding {
	if r.c == nil {
		return nil
	}

	var roleBindings []Rolebinding
	err := r.c.Get(ctx, key, &roleBindings)
	if err != nil {
		if !cache.IsCacheMissErr(err) {
			ctx.L().Error("failed to get cache", zap.Error(err))
		}
		return nil
	}

	return roleBindings
}
