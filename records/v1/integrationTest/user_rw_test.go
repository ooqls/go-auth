package integrationTest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/gen"
	"github.com/ooqls/go-auth/records/v1/users"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-db/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestUserRW(t *testing.T) {
	w := users.NewSQLWriter(sqlx.GetSQLX())
	r := users.NewSQLUserReader(cache.New[[]users.User](cache.NewMemCache()), sqlx.GetSQLX())
	t.Run("Create user and read", func(t *testing.T) {
		user1Id := uuid.New()
		assert.Nilf(t, w.CreateUser(context.Background(), gen.Authv1User{
			ID:       user1Id,
			Username: "test",
			Email:    "email",
			Salt:     []byte("123"),
			Key:      []byte("123"),
		}), "CreateUser should not return an error")

		user, err := r.GetUser(context.Background(), user1Id)
		assert.Nilf(t, err, "GetUser should not return an error")
		assert.Equal(t, user1Id, user.ID, "ID should be the same")

	})
}
