package integrationTest

import (
	"context"
	"testing"

	"github.com/braumsmilk/go-auth/pg"
	"github.com/braumsmilk/go-auth/pg/userv1"
	"github.com/braumsmilk/go-auth/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutils.InitPostgres()

	m.Run()
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()

	pg.InitDefault()
	r := userv1.PostgresRepository{}
	id, err := r.CreateUser(ctx, "user1", "name", "pw")
	assert.Nilf(t, err, "should not error when creating user")

	user, err := r.GetUser(ctx, id)
	assert.Nilf(t, err, "should not error when getting user")

	assert.Equalf(t, "user1", user.Email, "name should be the same")
	assert.Equalf(t, id, user.UserId, "userid should be the same")

	authed, id, err := r.Authenticate(ctx, user.Email, "pw")
	assert.Nilf(t, err, "should be able to authenticate")
	assert.Truef(t, authed, "authed should be true")
	assert.Greaterf(t, id, userv1.Id(-1), "should have gotten a valid userid")

	authed, id, err = r.Authenticate(ctx, user.Name, "fdnsjk")
	assert.Nilf(t, err, "should not error for using wrong password")
	assert.Falsef(t, authed, "should not be authed using wrong password")

	authed, id, err = r.Authenticate(ctx, "not_found", "fnjdks")
	assert.Nilf(t, err, "should not error when trying to authenticate using the wrong username")
	assert.Falsef(t, authed, "should not be authed for using wrong username")
}
