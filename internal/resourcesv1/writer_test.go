package resourcesv1

import (
	"context"
	"os"
	"testing"
	"time"

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
	// Create the resourcesv1 table
	testutils.SeedDatabase()
	time.Sleep(time.Second)

	os.Exit(m.Run())
}

func TestSQLWriter_CreateResource(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	tests := []struct {
		name        string
		group       string
		kind        string
		resName     string
		description string
		wantErr     bool
	}{
		{
			name:        "should create resource successfully",
			group:       "test-group",
			kind:        "test-kind",
			resName:     "test-resource-1",
			description: "A test resource",
			wantErr:     false,
		},
		{
			name:        "should create resource with empty description",
			group:       "test-group",
			kind:        "test-kind",
			resName:     "test-resource-2",
			description: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := writer.CreateResource(ctx, tt.group, tt.kind, tt.resName, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify the resource was created by querying it back
				resource, err := testQueries.GetResourceByID(ctx, datagen.GetResourceByIDParams{
					ResourceName:  tt.resName,
					ResourceGroup: tt.group,
					ResourceKind:  tt.kind,
				})
				require.NoError(t, err)
				assert.Equal(t, tt.resName, resource.ResourceName)
				assert.Equal(t, tt.group, resource.ResourceGroup)
				assert.Equal(t, tt.kind, resource.ResourceKind)
				assert.Equal(t, tt.description, resource.Description)
			}
		})
	}
}

func TestSQLWriter_DeleteResource(t *testing.T) {
	ctx := context.Background()
	writer := NewSQLWriter(testQueries)

	// First create a resource to delete
	group := "delete-test-group"
	kind := "delete-test-kind"
	resName := "resource-to-delete"

	_, err := writer.CreateResource(ctx, group, kind, resName, "will be deleted")
	require.NoError(t, err)

	// Delete the resource
	err = writer.DeleteResource(ctx, group, kind, resName)
	require.NoError(t, err)

	// Verify the resource was deleted
	_, err = testQueries.GetResourceByID(ctx, datagen.GetResourceByIDParams{
		ResourceName:  resName,
		ResourceGroup: group,
		ResourceKind:  kind,
	})
	assert.Error(t, err, "should return error when resource not found")
}
