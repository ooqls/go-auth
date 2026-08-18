package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/valkey"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/schema"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	containers.StartPostgres(ctx)
	containers.StartValkey(ctx)

	pgx.InitDefault()
	valkey.InitDefault()

	for _, schemas := range schema.GetSchemaMigrations() {
		_, err := pgx.GetPGX().Exec(ctx, schemas)
		if err != nil {
			panic(err)
		}
	}

	m.Run()
}

func TestRW(t *testing.T) {
	ctx := contexts.NewLoggingContext(t.Context(), zap.L())

	r := newReader(context.Background())
	w := newWriter()
	key, salt := newKey()
	_, err := w.CreateUser(ctx, datagen.CreateUserParams{
		ID:       uuid.New(),
		Username: "test",
		Email:    "email@email.com",
		Salt:     salt,
		Key:      []byte(key),
	})
	assert.Nilf(t, err, "should not have gotten an error when creating a new user")

	user, err := r.GetUserByUsername(ctx, "test")
	assert.Nilf(t, err, "should not get an error when getting user by username")
	assert.Equalf(t, "test", user.Username, "should have gotten the same username")
}
