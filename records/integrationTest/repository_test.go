package integrationTest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/gen"
	authtestutils "github.com/ooqls/go-auth/records/testutils"
	"github.com/ooqls/go-db/sqlx"
	"github.com/ooqls/go-db/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	timeout := time.Second * 30
	c := testutils.StartPostgres(ctx)
	defer c.Stop(ctx, &timeout)
	authtestutils.SeedDatabase()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()
	db := sqlx.GetSQLX()
	r := gen.New(db)
	user, err := r.CreateUser(ctx, gen.CreateUserParams{
		Username: "user1",
		Email:    "email",
		ID:       uuid.New(),
		Salt:     []byte("salt"),
		Key:      []byte("key"),
	})
	assert.Nilf(t, err, "should not error when creating user")

	u, err := r.GetUser(ctx, user.ID)
	assert.Nilf(t, err, "should not error when getting user")

	assert.Equalf(t, "email", u.Email, "email should be the same")
	assert.Equalf(t, user.ID, u.ID, "userid should be the same")
}
