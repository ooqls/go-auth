package resources

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
	f := datav1.NewFactory(*pgx.GetPGX(), &factory.MemCacheFactory{})
	return NewServiceImpl(f)
}

func newCtx() authorizationv1.Context {
	return authorizationv1.NewInternalOperationContext(context.Background())
}

func TestCreateResource(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("creates a resource successfully", func(t *testing.T) {
		res, err := svc.CreateResource(ctx, "testgroup", "testkind", "create-test-resource", "test description")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "create-test-resource", res.Name)
		assert.Equal(t, "test description", res.Description)
	})

	t.Run("creates multiple resources with different names", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("create-multi-%d", i)
			res, err := svc.CreateResource(ctx, "multigroup", "multikind", name, "desc")
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, name, res.Name)
		}
	})
}

func TestGetResource(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("returns nil for nonexistent resource", func(t *testing.T) {
		res, err := svc.GetResource(ctx, "nonexistent", "nonexistent", "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("retrieves an existing resource by name, group, and kind", func(t *testing.T) {
		_, err := svc.CreateResource(ctx, "getgroup", "getkind", "get-test-resource", "get desc")
		require.NoError(t, err)

		retrieved, err := svc.GetResource(ctx, "getgroup", "getkind", "get-test-resource")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "get-test-resource", retrieved.Name)
		assert.Equal(t, "get desc", retrieved.Description)
	})
}

func TestDeleteResource(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("deletes an existing resource by id", func(t *testing.T) {
		created, err := svc.CreateResource(ctx, "delgroup", "delkind", "delete-by-id-resource", "desc")
		require.NoError(t, err)
		require.NotNil(t, created)

		err = svc.DeleteResource(ctx, created.Id)
		require.NoError(t, err)

		retrieved, err := svc.GetResource(ctx, "delgroup", "delkind", "delete-by-id-resource")
		require.NoError(t, err)
		assert.Nil(t, retrieved, "resource should be deleted")
	})
}

func TestDeleteResourceByName(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("deletes an existing resource by name", func(t *testing.T) {
		_, err := svc.CreateResource(ctx, "delngroup", "delnkind", "delete-by-name-resource", "desc")
		require.NoError(t, err)

		err = svc.DeleteResourceByName(ctx, "delngroup", "delnkind", "delete-by-name-resource")
		require.NoError(t, err)

		retrieved, err := svc.GetResource(ctx, "delngroup", "delnkind", "delete-by-name-resource")
		require.NoError(t, err)
		assert.Nil(t, retrieved, "resource should be deleted")
	})
}

func TestUpdateResourceName(t *testing.T) {
	svc := newService()
	ctx := newCtx()

	t.Run("updates name and description for existing resource", func(t *testing.T) {
		created, err := svc.CreateResource(ctx, "updgroup", "updkind", "update-test-resource", "original desc")
		require.NoError(t, err)
		require.NotNil(t, created)

		newName := "updated-resource-name"
		newDesc := "updated desc"
		updated, err := svc.UpdateResourceName(ctx, created.Id, newName, &newDesc)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, newName, updated.Name)
		assert.Equal(t, newDesc, updated.Description)
	})
}
