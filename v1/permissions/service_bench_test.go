package permissions_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/testutils"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/seed"
	"github.com/ooqls/go-auth/v1/users"
)

// Seed size for the benchmark dataset: enough roles/permissions/bindings to
// exercise pagination, search and caching without making the one-time setup
// prohibitively slow.
const (
	benchNumUsers     = 40
	benchPermsPerRole = 15
)

var (
	benchService    permissions.Service
	benchCtx        *authorizationv1.Context // internal operation context, bypasses authz
	benchUserCtx    *authorizationv1.Context // context scoped to a seeded user
	benchPermNames  []string
	benchCategories = []string{"orders", "invoices", "shipments", "billing", "reports"}
	benchActions    = []string{"read", "write", "delete", "list"}
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

	f := datav1.NewFactory(pgx.GetPGX(), &factory.MemCacheFactory{})
	benchService = permissions.NewServiceImpl(f)

	internalCtx := authorizationv1.NewInternalOperationContext(context.Background())
	benchCtx = &internalCtx

	if err := seedBenchmarkData(f); err != nil {
		panic(fmt.Errorf("failed to seed benchmark data: %w", err))
	}

	os.Exit(m.Run())
}

// seedBenchmarkData uses the seed service to populate the database with
// benchNumUsers users, each with one role bound to benchPermsPerRole
// permissions. Permission names are grouped by category so SearchPermissions
// has a realistic number of matches to page through.
func seedBenchmarkData(f datav1.Factory) error {
	seedSvc := seed.NewServiceImpl(f)

	var s seed.Seed
	for i := 0; i < benchNumUsers; i++ {
		category := benchCategories[i%len(benchCategories)]

		perms := make([]string, 0, benchPermsPerRole)
		for j := 0; j < benchPermsPerRole; j++ {
			action := benchActions[j%len(benchActions)]
			name := fmt.Sprintf("%s:%s:%04d", category, action, i*benchPermsPerRole+j)
			perms = append(perms, name)
			benchPermNames = append(benchPermNames, name)
		}

		username := fmt.Sprintf("bench-user-%03d", i)
		s.Users = append(s.Users, seed.UserSeed{
			Email:    username + "@bench.test",
			Password: "benchpassword123",
			Username: username,
			Roles: []seed.RoleSeed{
				{
					Name:        fmt.Sprintf("bench-role-%03d", i),
					Description: "benchmark role",
					Permissions: perms,
					Hierarchy:   int32(i % 5),
				},
			},
		})
	}

	if err := seedSvc.Seed(benchCtx, s); err != nil {
		return err
	}

	usersSvc := users.NewServiceImpl(f)
	seededUser, err := usersSvc.GetUserByUsername(benchCtx, "bench-user-000")
	if err != nil {
		return fmt.Errorf("failed to look up seeded user: %w", err)
	}

	userCtx := authorizationv1.NewAuthorizationContext(context.Background(), *seededUser)
	benchUserCtx = &userCtx

	return nil
}

func BenchmarkAddPermission(b *testing.B) {
	names := make([]string, b.N)
	ts := time.Now().UnixNano()
	for i := range names {
		names[i] = fmt.Sprintf("bench-add:%d:%06d", ts, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := benchService.AddPermission(benchCtx, names[i]); err != nil {
			b.Fatalf("AddPermission failed: %v", err)
		}
	}
}

func BenchmarkDeletePermission(b *testing.B) {
	names := make([]string, b.N)
	ts := time.Now().UnixNano()
	for i := range names {
		names[i] = fmt.Sprintf("bench-delete:%d:%06d", ts, i)
		if err := benchService.AddPermission(benchCtx, names[i]); err != nil {
			b.Fatalf("failed to seed permission for delete benchmark: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := benchService.DeletePermission(benchCtx, names[i]); err != nil {
			b.Fatalf("DeletePermission failed: %v", err)
		}
	}
}

func BenchmarkGetPermission(b *testing.B) {
	name := benchPermNames[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := benchService.GetPermission(benchCtx, name); err != nil {
			b.Fatalf("GetPermission failed: %v", err)
		}
	}
}

func BenchmarkGetPermissions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := benchService.GetPermissions(benchCtx, 1, 25); err != nil {
			b.Fatalf("GetPermissions failed: %v", err)
		}
	}
}

func BenchmarkSearchPermissions(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := benchService.SearchPermissions(benchCtx, "orders", 1, 25); err != nil {
			b.Fatalf("SearchPermissions failed: %v", err)
		}
	}
}

func BenchmarkListPermissionsForUser(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := benchService.ListPermissionsForUser(benchUserCtx); err != nil {
			b.Fatalf("ListPermissionsForUser failed: %v", err)
		}
	}
}
