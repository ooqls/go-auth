package resourcesv1

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	"github.com/ooqls/go-auth/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testQueries *datagen.Queries

func TestMain(m *testing.M) {
	ctx := context.Background()
	timeout := time.Second * 30

	c := containers.StartPostgres(ctx)
	defer c.Stop(ctx, &timeout)

	err := pgx.InitDefault()
	if err != nil {
		panic(err)
	}

	testQueries = datagen.New(pgx.GetPGX())
	testutils.SeedDatabase()
	time.Sleep(time.Second)

	os.Exit(m.Run())
}

func TestSQLWriter_CreateResource(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	tests := []struct {
		name    string
		resName string
		wantErr bool
	}{
		{
			name:    "should create resource successfully",
			resName: "test-resource-1",
			wantErr: false,
		},
		{
			name:    "should create another resource successfully",
			resName: "test-resource-2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := "test"
			kind := "test_kind"
			_, err := writer.CreateResource(ctx, group, kind, tt.resName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				resource, err := testQueries.GetResourceByName(ctx, tt.resName)
				require.NoError(t, err)
				assert.Equal(t, tt.resName, resource.Name)
			}
		})
	}
}

func TestSQLWriter_CreateResource_PopulatesMetadata(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	group := "test"
	kind := "test_kind"
	res, err := writer.CreateResource(ctx, group, kind, "metadata-create-resource")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, group, res.Group)
	assert.Equal(t, kind, res.Kind)
	assert.NotEqual(t, uuid.Nil, res.Id)
}

func TestSQLWriter_UpdateResource(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	group := "test"
	kind := "test_kind"
	created, err := writer.CreateResource(ctx, group, kind, "resource-before-update")
	require.NoError(t, err)

	updated, err := writer.UpdateResource(ctx, group, kind, "resource-before-update", "resource-after-update")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, created.Id, updated.Id)
	assert.Equal(t, "resource-after-update", updated.Name)
	// FromDatagenResource must carry the group/kind through the update.
	assert.Equal(t, group, updated.Group)
	assert.Equal(t, kind, updated.Kind)

	stored, err := testQueries.GetResourceByName(ctx, "resource-after-update")
	require.NoError(t, err)
	assert.Equal(t, "resource-after-update", stored.Name)
}

func TestSQLWriter_UpdateResource_NotFound(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	_, err := writer.UpdateResource(ctx, "test", "test_kind", "does-not-exist", "new-name")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSQLWriter_DeleteResource(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	resName := "resource-to-delete"
	group := "test"
	kind := "test_kind"
	_, err := writer.CreateResource(ctx, group, kind, resName)
	require.NoError(t, err)

	err = writer.DeleteResource(ctx, group, kind, resName)
	require.NoError(t, err)

	_, err = testQueries.GetResourceByName(ctx, resName)
	assert.Error(t, err, "should return error when resource not found")
}

func TestSQLWriter_DeleteResourceById(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	created, err := writer.CreateResource(ctx, "test", "test_kind", "resource-to-delete-by-id")
	require.NoError(t, err)

	err = writer.DeleteResourceById(ctx, created.Id)
	require.NoError(t, err)

	_, err = testQueries.GetResourceByName(ctx, "resource-to-delete-by-id")
	assert.Error(t, err, "should return error when resource not found")
}
