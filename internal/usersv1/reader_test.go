package usersv1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/schema"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var totalCount int64 = 0

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	containers.StartPostgres(ctx)
	defer cancel()

	err := pgx.InitDefault()
	if err != nil {
		panic(err)
	}
	for _, schemas := range schema.GetSchemaMigrations() {
		_, err := pgx.GetPGX().Exec(ctx, schemas)
		if err != nil {
			panic(err)
		}
	}

	m.Run()
}

func newCtx() contexts.LContext {
	return contexts.NewLoggingContext(context.Background(), log.NewLogger("test"))
}

func makeUser(t *testing.T, w Writer, username string) datagen.CreateUserParams {
	t.Helper()
	totalCount += 1
	params := datagen.CreateUserParams{
		ID:       uuid.New(),
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Salt:     []byte("salt"),
		Key:      []byte("key"),
	}
	_, err := w.CreateUser(newCtx(), params)
	require.NoError(t, err)
	return params
}

func TestGetUser(t *testing.T) {
	r := newReader()
	w := newWriter()

	params := makeUser(t, w, "getuser_"+uuid.NewString()[:8])

	user, err := r.GetUser(newCtx(), params.ID)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, params.ID, user.Object.Id)
	assert.Equal(t, params.Username, user.Username)
	assert.Equal(t, params.Email, user.Email)
}

func TestGetUser_NotFound(t *testing.T) {
	r := newReader()

	user, err := r.GetUser(newCtx(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetUser_CacheHit(t *testing.T) {
	r := newReader()
	w := newWriter()

	params := makeUser(t, w, "cacheget_"+uuid.NewString()[:8])

	user1, err := r.GetUser(newCtx(), params.ID)
	require.NoError(t, err)
	require.NotNil(t, user1)

	// second call should be served from cache
	user2, err := r.GetUser(newCtx(), params.ID)
	require.NoError(t, err)
	require.NotNil(t, user2)
	assert.Equal(t, user1.Object.Id, user2.Object.Id)
	assert.Equal(t, user1.Username, user2.Username)
}

func TestGetUserByUsername(t *testing.T) {
	r := newReader()
	w := newWriter()

	params := makeUser(t, w, "byusername_"+uuid.NewString()[:8])

	user, err := r.GetUserByUsername(newCtx(), params.Username)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, params.ID, user.Object.Id)
	assert.Equal(t, params.Username, user.Username)
	assert.Equal(t, params.Email, user.Email)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	r := newReader()

	user, err := r.GetUserByUsername(newCtx(), "does_not_exist_"+uuid.NewString())
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetUsers(t *testing.T) {
	r := newReader()
	w := newWriter()

	prefix := uuid.NewString()[:8]
	for i := 0; i < 3; i++ {
		makeUser(t, w, fmt.Sprintf("listuser_%s_%d", prefix, i))
	}

	users, err := r.GetUsers(newCtx(), 0, 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users.Items), 3)
	assert.Equal(t, users.TotalCount, totalCount)
}

func TestGetUsers_Pagination(t *testing.T) {
	r := newReader()
	w := newWriter()

	prefix := uuid.NewString()[:8]
	for i := 0; i < 5; i++ {
		makeUser(t, w, fmt.Sprintf("pageuser_%s_%d", prefix, i))
	}

	page1, err := r.GetUsers(newCtx(), 0, 2)
	require.NoError(t, err)
	assert.Len(t, page1.Items, 2)
	assert.Equal(t, page1.TotalCount, totalCount)

	page2, err := r.GetUsers(newCtx(), 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2.Items, 2)
	assert.Equal(t, page2.TotalCount, totalCount)

	assert.NotEqual(t, page1.Items[0].Object.Id, page2.Items[0].Object.Id)
}

func TestClearCache(t *testing.T) {
	r := newReader()
	w := newWriter()

	params := makeUser(t, w, "clearcache_"+uuid.NewString()[:8])

	// populate cache
	_, err := r.GetUser(newCtx(), params.ID)
	require.NoError(t, err)

	err = r.ClearCache(newCtx())
	require.NoError(t, err)

	// should still read from DB after cache cleared
	user, err := r.GetUser(newCtx(), params.ID)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, params.ID, user.Object.Id)
}
