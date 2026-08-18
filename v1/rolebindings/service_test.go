package rolebindings

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	rolesdatagen "github.com/ooqls/go-auth/internal/rolesv1/datagen"
	"github.com/ooqls/go-auth/internal/testutils"
	usersdatagen "github.com/ooqls/go-auth/internal/usersv1/datagen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	timeout := 30 * time.Second

	c := containers.StartPostgres(ctx)
	defer c.Stop(context.Background(), &timeout)

	if err := pgx.InitDefault(); err != nil {
		panic(err)
	}

	testutils.SeedDatabase()
	time.Sleep(time.Second)

	os.Exit(m.Run())
}

func newService() *ServiceImpl {
	f := datav1.NewFactory(pgx.GetPGX(), &factory.MemCacheFactory{})
	return NewServiceImpl(f)
}

func newCtx() *authorizationv1.Context {
	ctx := authorizationv1.NewInternalOperationContext(context.Background())
	return &ctx
}

// createTestUser inserts a minimal user row and returns its UUID.
func createTestUser(t *testing.T, suffix string) uuid.UUID {
	t.Helper()
	conn := pgx.GetPGX()
	q := usersdatagen.New(conn)
	userID := uuid.New()
	_, err := q.CreateUser(context.Background(), usersdatagen.CreateUserParams{
		ID:       userID,
		Username: fmt.Sprintf("rb-user-%s", suffix),
		Email:    fmt.Sprintf("rb-%s@test.com", suffix),
		Salt:     []byte("testsalt"),
		Key:      []byte("testkey0testkey0testkey0testkey0"),
	})
	require.NoError(t, err)
	return userID
}

// createTestRole inserts a minimal role row and returns its UUID.
func createTestRole(t *testing.T, suffix string) uuid.UUID {
	t.Helper()
	conn := pgx.GetPGX()
	q := rolesdatagen.New(conn)
	role, err := q.CreateRole(context.Background(), rolesdatagen.CreateRoleParams{
		Name:        fmt.Sprintf("rb-role-%s", suffix),
		Description: "test role",
		Hierarchy:   0,
	})
	require.NoError(t, err)
	return role.ID
}

func TestAssignRoleToUser(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("assigns a role to a user successfully", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID := createTestRole(t, uuid.New().String()[:8])

		err := svc.AssignRoleToUser(ctx, userID, roleID)
		require.NoError(t, err)
	})

	t.Run("assigns multiple roles to the same user", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID1 := createTestRole(t, uuid.New().String()[:8])
		roleID2 := createTestRole(t, uuid.New().String()[:8])

		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID1))
		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID2))
	})
}

func TestGetRoleBindingsForUser(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("returns empty slice for user with no bindings", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])

		bindings, err := svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, bindings)
	})

	t.Run("returns bindings after role assignment", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID := createTestRole(t, uuid.New().String()[:8])

		err := svc.AssignRoleToUser(ctx, userID, roleID)
		require.NoError(t, err)

		bindings, err := svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		require.Len(t, bindings, 1)
		assert.Equal(t, userID, bindings[0].UserID)
		assert.Equal(t, roleID, bindings[0].RoleID)
	})

	t.Run("returns all bindings when multiple roles assigned", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID1 := createTestRole(t, uuid.New().String()[:8])
		roleID2 := createTestRole(t, uuid.New().String()[:8])

		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID1))
		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID2))

		bindings, err := svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, bindings, 2)
	})
}

func TestUnassignRoleFromUser(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("removes an assigned role", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID := createTestRole(t, uuid.New().String()[:8])

		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID))

		bindings, err := svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		require.Len(t, bindings, 1)

		err = svc.UnassignRoleFromUser(ctx, userID, roleID)
		require.NoError(t, err)

		bindings, err = svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, bindings)
	})

	t.Run("unassigning one role leaves others intact", func(t *testing.T) {
		userID := createTestUser(t, uuid.New().String()[:8])
		roleID1 := createTestRole(t, uuid.New().String()[:8])
		roleID2 := createTestRole(t, uuid.New().String()[:8])

		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID1))
		require.NoError(t, svc.AssignRoleToUser(ctx, userID, roleID2))

		require.NoError(t, svc.UnassignRoleFromUser(ctx, userID, roleID1))

		bindings, err := svc.GetRoleBindingsForUser(ctx, userID)
		require.NoError(t, err)
		require.Len(t, bindings, 1)
		assert.Equal(t, roleID2, bindings[0].RoleID)
	})
}
