//go:build integration

package permissionbindings_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/db/containers"
	pgxdefault "github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	bindingsdata "github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
	rolesdata "github.com/ooqls/go-auth/internal/rolesv1/datagen"
	"github.com/ooqls/go-auth/internal/testutils"
	"github.com/ooqls/go-auth/internal/usersv1"
	usersdata "github.com/ooqls/go-auth/internal/usersv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/permissionbindings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()
	timeout := 30 * time.Second
	c := containers.StartPostgres(ctx)
	defer c.Stop(ctx, &timeout)
	if err := pgxdefault.InitDefault(); err != nil {
		panic(err)
	}
	testutils.SeedDatabase()
	db = pgxdefault.GetPGX()
	defer db.Close()
	return m.Run()
}

// The default memory factory creates independent stores on each call. Share
// each namespace here to exercise invalidation across factory-created readers.
type sharedCacheFactory struct {
	factory.MemCacheFactory
	caches map[string]cache.GenericCache
}

func (f *sharedCacheFactory) NewCache(key string, ttl time.Duration) cache.GenericCache {
	if c, ok := f.caches[key]; ok {
		return c
	}
	c := f.MemCacheFactory.NewCache(key, ttl)
	f.caches[key] = c
	return c
}

type fixture struct {
	factory *datav1.SQLFactory
	svc     permissionbindings.Service
	ctx     *authorizationv1.Context
	role    uuid.UUID
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	f := datav1.NewFactory(db, &sharedCacheFactory{caches: make(map[string]cache.GenericCache)})
	userID := uuid.New()
	_, err := usersdata.New(db).CreateUser(context.Background(), usersdata.CreateUserParams{
		ID: userID, Username: userID.String()[:20], Email: userID.String() + "@test.com", Key: []byte("key"), Salt: []byte("salt"),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(context.Background(), "DELETE FROM usersv1 WHERE id=$1", userID)
		require.NoError(t, err)
	})
	role := createRole(t, 10)
	_, err = db.Exec(context.Background(), "INSERT INTO rolebindingsv1(user_id,role_id) VALUES($1,$2)", userID, role)
	require.NoError(t, err)
	for _, permission := range []string{"core:auth:role:update", "core:auth:permission:assign", "core:auth:permission:unassign"} {
		grant(t, role, permission)
	}
	ctx := authorizationv1.NewAuthorizationContext(context.Background(), usersv1.User{})
	ctx.User.Id = userID
	return fixture{factory: f, svc: permissionbindings.NewServiceImpl(f), ctx: &ctx, role: role}
}

func createRole(t *testing.T, hierarchy int32) uuid.UUID {
	t.Helper()
	role, err := rolesdata.New(db).CreateRole(context.Background(), rolesdata.CreateRoleParams{
		Name: uuid.NewString()[:20], Description: "binding test", Hierarchy: hierarchy,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(context.Background(), "DELETE FROM rolesv1 WHERE id=$1", role.ID)
		require.NoError(t, err)
	})
	return role.ID
}

func grant(t *testing.T, role uuid.UUID, permission string) {
	t.Helper()
	require.NoError(t, bindingsdata.New(db).AssignPermission(context.Background(), bindingsdata.AssignPermissionParams{RoleID: role, Permission: permission}))
}

func hasBinding(t *testing.T, role uuid.UUID, permission string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM permissionbindingsv1 WHERE role_id=$1 AND permission=$2)", role, permission).Scan(&exists))
	return exists
}

func hasPermission(t *testing.T, permission string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM permissionsv1 WHERE permission=$1)", permission).Scan(&exists))
	return exists
}

func TestBindingHierarchy(t *testing.T) {
	for _, operation := range []string{"assign", "unassign"} {
		t.Run(operation, func(t *testing.T) {
			for _, tc := range []struct {
				name      string
				hierarchy int32
				allowed   bool
			}{
				{name: "lower", hierarchy: 9, allowed: true},
				{name: "equal", hierarchy: 10, allowed: true},
				{name: "higher", hierarchy: 11},
			} {
				t.Run(tc.name, func(t *testing.T) {
					f := newFixture(t)
					target := createRole(t, tc.hierarchy)
					permission := "test:" + uuid.NewString()
					var err error
					if operation == "assign" {
						err = f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission})
					} else {
						grant(t, target, permission)
						err = f.svc.UnassignPermission(f.ctx, []uuid.UUID{target}, []string{permission})
					}
					if tc.allowed {
						require.NoError(t, err)
					} else {
						var forbidden *v1.ForbiddenError
						require.ErrorAs(t, err, &forbidden)
					}
					wantBinding := tc.allowed
					if operation == "unassign" {
						wantBinding = !tc.allowed
					}
					assert.Equal(t, wantBinding, hasBinding(t, target, permission))
				})
			}
		})
	}
}

func TestBindingRequiredPermissions(t *testing.T) {
	for _, tc := range []struct {
		name, operation, revoke string
	}{
		{name: "assign needs role update", operation: "assign", revoke: "core:auth:role:update"},
		{name: "assign needs assign", operation: "assign", revoke: "core:auth:permission:assign"},
		{name: "unassign needs role update", operation: "unassign", revoke: "core:auth:role:update"},
		{name: "unassign needs unassign", operation: "unassign", revoke: "core:auth:permission:unassign"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			target := createRole(t, 1)
			permission := "test:" + uuid.NewString()
			grant(t, target, permission)
			_, err := f.factory.NewPermissionReader().GetPermissionsForUser(f.ctx, f.ctx.User.Id)
			require.NoError(t, err)
			require.NoError(t, bindingsdata.New(db).UnassignPermission(f.ctx, bindingsdata.UnassignPermissionParams{RoleID: f.role, Permission: tc.revoke}))
			if tc.operation == "assign" {
				err = f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission + "-new"})
			} else {
				err = f.svc.UnassignPermission(f.ctx, []uuid.UUID{target}, []string{permission})
			}
			var forbidden *v1.ForbiddenError
			require.ErrorAs(t, err, &forbidden)
			assert.True(t, hasBinding(t, target, permission))
			assert.False(t, hasPermission(t, permission+"-new"))
		})
	}
}

func TestBindingPreflight(t *testing.T) {
	for _, tc := range []struct {
		name, operation string
		internal        bool
		missing         bool
	}{
		{name: "assign denied later role", operation: "assign"},
		{name: "unassign denied later role", operation: "unassign"},
		{name: "assign missing later role", operation: "assign", missing: true},
		{name: "unassign missing later role", operation: "unassign", missing: true},
		{name: "internal assign missing role", operation: "assign", missing: true, internal: true},
		{name: "internal unassign missing role", operation: "unassign", missing: true, internal: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			first, second := createRole(t, 1), uuid.New()
			if !tc.missing {
				second = createRole(t, 11)
			}
			ctx := f.ctx
			if tc.internal {
				internal := authorizationv1.NewInternalOperationContext(context.Background())
				ctx = &internal
			}
			permission := "test:" + uuid.NewString()
			var err error
			if tc.operation == "assign" {
				err = f.svc.AssignPermission(ctx, []uuid.UUID{first, second}, []string{permission})
			} else {
				grant(t, first, permission)
				err = f.svc.UnassignPermission(ctx, []uuid.UUID{first, second}, []string{permission})
			}
			if tc.missing {
				var notFound *v1.NotFoundError
				require.ErrorAs(t, err, &notFound)
			} else {
				var forbidden *v1.ForbiddenError
				require.ErrorAs(t, err, &forbidden)
			}
			assert.Equal(t, tc.operation == "unassign", hasBinding(t, first, permission))
			if tc.operation == "assign" {
				assert.False(t, hasPermission(t, permission), "denied batches must not create permissions")
			}
		})
	}
}

func TestBindingWildcardPermissions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		grants  []string
		allowed bool
	}{
		{name: "all core actions", grants: []string{"core:*"}, allowed: true},
		{name: "all kinds and actions", grants: []string{"core:auth:*:*"}, allowed: true},
		{name: "field wildcards", grants: []string{"*:auth:role:update", "core:*:permission:*"}, allowed: true},
		{name: "resource grants do not grant core actions", grants: []string{"resource:auth:*:*"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			target := createRole(t, 1)
			require.NoError(t, bindingsdata.New(db).UnassignAllPermissions(f.ctx, f.role))
			for _, grantValue := range tc.grants {
				grant(t, f.role, grantValue)
			}
			permission := "test:" + uuid.NewString()
			err := f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission})
			if !tc.allowed {
				var forbidden *v1.ForbiddenError
				require.ErrorAs(t, err, &forbidden)
				assert.False(t, hasBinding(t, target, permission))
				return
			}
			require.NoError(t, err)
			assert.True(t, hasBinding(t, target, permission))
			require.NoError(t, f.svc.UnassignPermission(f.ctx, []uuid.UUID{target}, []string{permission}))
			assert.False(t, hasBinding(t, target, permission))
		})
	}
}

func TestBindingAssignmentRollback(t *testing.T) {
	f := newFixture(t)
	target := createRole(t, 1)
	permission := "test:" + uuid.NewString()
	err := f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission, strings.Repeat("x", 129)})
	var internal *v1.InternalError
	require.ErrorAs(t, err, &internal)
	assert.False(t, hasBinding(t, target, permission))
	assert.False(t, hasPermission(t, permission), "new permission creation must roll back too")
}

func TestBindingUnassignmentRollback(t *testing.T) {
	f := newFixture(t)
	first, second := createRole(t, 1), createRole(t, 1)
	permission := "test:" + uuid.NewString()
	grant(t, first, permission)
	grant(t, second, permission)
	// Force the second DELETE to fail after the first has already executed.
	name := "fail_binding_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err := db.Exec(context.Background(), fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected binding failure'; END $$;
CREATE TRIGGER %s BEFORE DELETE ON permissionbindingsv1 FOR EACH ROW WHEN (OLD.role_id = '%s'::uuid) EXECUTE FUNCTION %s();`, name, name, second, name))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := db.Exec(context.Background(), fmt.Sprintf("DROP TRIGGER %s ON permissionbindingsv1; DROP FUNCTION %s();", name, name))
		require.NoError(t, err)
	})
	err = f.svc.UnassignPermission(f.ctx, []uuid.UUID{first, second}, []string{permission})
	var internal *v1.InternalError
	require.ErrorAs(t, err, &internal)
	assert.True(t, hasBinding(t, first, permission))
	assert.True(t, hasBinding(t, second, permission))
}

func TestBindingIdempotencyAndInternalOperations(t *testing.T) {
	f := newFixture(t)
	first, second := createRole(t, 100), createRole(t, 200)
	ctx := authorizationv1.NewInternalOperationContext(context.Background())
	roles := []uuid.UUID{first, second, first}
	permissions := []string{"test:" + uuid.NewString(), "test:" + uuid.NewString()}
	for range 2 {
		require.NoError(t, f.svc.AssignPermission(&ctx, roles, permissions))
	}
	for _, role := range roles {
		for _, permission := range permissions {
			assert.True(t, hasBinding(t, role, permission))
		}
	}
	for range 2 {
		require.NoError(t, f.svc.UnassignPermission(&ctx, roles, permissions))
	}
	for _, role := range roles {
		for _, permission := range permissions {
			assert.False(t, hasBinding(t, role, permission))
		}
	}
}

func TestBindingFreshHierarchy(t *testing.T) {
	for _, tc := range []struct {
		name, operation string
		changeRequester bool
	}{
		{name: "assign after requester demotion", operation: "assign", changeRequester: true},
		{name: "unassign after requester demotion", operation: "unassign", changeRequester: true},
		{name: "assign after target promotion", operation: "assign"},
		{name: "unassign after target promotion", operation: "unassign"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			target := createRole(t, 5)
			permission := "test:" + uuid.NewString()
			grant(t, target, permission)
			_, err := f.factory.NewRoleReader().GetRole(f.ctx, target)
			require.NoError(t, err)
			_, err = f.factory.NewAggReader().GetUserHierarchy(f.ctx, f.ctx.User.Id)
			require.NoError(t, err)
			if tc.changeRequester {
				_, err = db.Exec(context.Background(), "UPDATE rolesv1 SET hierarchy=1 WHERE id=$1", f.role)
			} else {
				_, err = db.Exec(context.Background(), "UPDATE rolesv1 SET hierarchy=20 WHERE id=$1", target)
			}
			require.NoError(t, err)
			if tc.operation == "assign" {
				err = f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission + "-new"})
			} else {
				err = f.svc.UnassignPermission(f.ctx, []uuid.UUID{target}, []string{permission})
			}
			var forbidden *v1.ForbiddenError
			require.ErrorAs(t, err, &forbidden)
			assert.True(t, hasBinding(t, target, permission))
			assert.False(t, hasPermission(t, permission+"-new"))
		})
	}
}

func TestBindingDerivedViews(t *testing.T) {
	f := newFixture(t)
	target := createRole(t, 1)
	permission := "test:" + uuid.NewString()
	// Warm both derived views before assigning a newly created permission.
	reader := f.factory.NewPermissionBindingReader()
	_, err := reader.GetPermissionBindingsForRole(f.ctx, target)
	require.NoError(t, err)
	_, err = f.factory.NewPermissionReader().GetPermissions(f.ctx, 1, 10000)
	require.NoError(t, err)
	require.NoError(t, f.svc.AssignPermission(f.ctx, []uuid.UUID{target}, []string{permission}))
	bindings, err := reader.GetPermissionBindingsForRole(f.ctx, target)
	require.NoError(t, err)
	require.Len(t, bindings, 1)
	assert.Equal(t, permission, bindings[0].Permission)
	catalog, err := f.factory.NewPermissionReader().GetPermissions(f.ctx, 1, 10000)
	require.NoError(t, err)
	found := false
	for _, item := range catalog.Items {
		found = found || item.Permission == permission
	}
	require.True(t, found, "assignment-created permissions must appear in the catalog")
	require.NoError(t, f.svc.UnassignPermission(f.ctx, []uuid.UUID{target}, []string{permission}))
	bindings, err = f.factory.NewPermissionBindingReader().GetPermissionBindingsForRole(f.ctx, target)
	require.NoError(t, err)
	assert.Empty(t, bindings)
}

func TestBindingStorageFailure(t *testing.T) {
	f := newFixture(t)
	target := createRole(t, 1)
	ctx, cancel := context.WithCancel(f.ctx.Context)
	cancel()
	cancelled := *f.ctx
	cancelled.Context = ctx
	permission := "test:" + uuid.NewString()
	err := f.svc.AssignPermission(&cancelled, []uuid.UUID{target}, []string{permission})
	var internal *v1.InternalError
	require.ErrorAs(t, err, &internal)
	require.ErrorIs(t, err, context.Canceled)
	assert.False(t, hasPermission(t, permission))
	assert.False(t, hasBinding(t, target, permission))
	// Missing targets use the shared v1 error type for caller classification.
	err = f.svc.AssignPermission(f.ctx, []uuid.UUID{uuid.New()}, []string{permission})
	var notFound *v1.NotFoundError
	require.ErrorAs(t, err, &notFound)
}
