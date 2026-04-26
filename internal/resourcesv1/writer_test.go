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
