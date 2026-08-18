package resourcesv1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReader() Reader {
	c := cache.NewMemCache()
	cc := cache.NewGenericCache("test", c)
	return NewSQLReader(cc, datagen.New(pgx.GetPGX()))
}

func makeResource(t *testing.T, w Writer, group, kind, name string) *Resourcev1 {
	t.Helper()
	res, err := w.CreateResource(context.Background(), group, kind, name)
	require.NoError(t, err)
	require.NotNil(t, res)
	return res
}

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.NewString()[:8])
}

func TestGetResourceByID(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	created := makeResource(t, w, "test", "test_kind", uniqueName("byid"))

	res, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, created.Id, res.Id)
	assert.Equal(t, created.Name, res.Name)
	// FromDatagenResource must populate the metadata
	assert.Equal(t, "test", res.Group)
	assert.Equal(t, "test_kind", res.Kind)
}

func TestGetResourceByID_NotFound(t *testing.T) {
	r := newReader()

	res, err := r.GetResourceByID(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestGetResource(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	name := uniqueName("byname")
	created := makeResource(t, w, "test", "test_kind", name)

	res, err := r.GetResource(ctx, "test", "test_kind", name)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, created.Id, res.Id)
	assert.Equal(t, name, res.Name)
	assert.Equal(t, "test", res.Group)
	assert.Equal(t, "test_kind", res.Kind)
}

func TestGetResource_NotFound(t *testing.T) {
	r := newReader()

	res, err := r.GetResource(context.Background(), "test", "test_kind", uniqueName("missing"))
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestGetResourceByID_CacheHit(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	created := makeResource(t, w, "test", "test_kind", uniqueName("cache"))

	// First read populates the cache.
	first, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)
	require.NotNil(t, first)

	// Remove the row from the database; a cache hit should still return it.
	require.NoError(t, testQueries.DeleteResourceById(ctx, created.Id))

	second, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, first.Id, second.Id)
	assert.Equal(t, first.Name, second.Name)
}

func TestGetResources(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	for i := 0; i < 3; i++ {
		makeResource(t, w, "test", "test_kind", uniqueName("list"))
	}

	res, err := r.GetResources(ctx, 100, 0)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.GreaterOrEqual(t, len(res.Items), 3)
	assert.GreaterOrEqual(t, res.TotalCount, int64(3))
}

func TestSearchResource(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	for i := 0; i < 3; i++ {
		makeResource(t, w, "test", "test_kind", uniqueName("search"))
	}
	search := "search"
	res, err := r.SearchResources(ctx, nil, nil, nil, &search, 100, 0)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.GreaterOrEqual(t, len(res.Items), 3)
	assert.GreaterOrEqual(t, res.TotalCount, int64(3))
}

func TestGetResourcesByGroup(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	group := uniqueName("group")
	for i := 0; i < 3; i++ {
		makeResource(t, w, group, "test_kind", uniqueName("g"))
	}

	res, err := r.GetResourcesByGroup(ctx, group, 100, 0)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 3)
	assert.Equal(t, int64(3), res.TotalCount)
	for _, item := range res.Items {
		assert.Equal(t, group, item.Group)
	}
}

func TestGetResourcesByGroup_Pagination(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	group := uniqueName("paged")
	for i := 0; i < 5; i++ {
		makeResource(t, w, group, "test_kind", uniqueName("p"))
	}

	page1, err := r.GetResourcesByGroup(ctx, group, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1.Items, 2)
	assert.Equal(t, int64(5), page1.TotalCount)

	page2, err := r.GetResourcesByGroup(ctx, group, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2.Items, 2)
	assert.Equal(t, int64(5), page2.TotalCount)

	assert.NotEqual(t, page1.Items[0].Id, page2.Items[0].Id)
}

func TestGetResourcesByGroupAndKind(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	group := uniqueName("gk")
	kind := "special_kind"
	for i := 0; i < 2; i++ {
		makeResource(t, w, group, kind, uniqueName("gk"))
	}
	// Different kind in the same group should be excluded.
	makeResource(t, w, group, "other_kind", uniqueName("gk"))

	res, err := r.GetResourcesByGroupAndKind(ctx, group, kind, 100, 0)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, int64(2), res.TotalCount)
	for _, item := range res.Items {
		assert.Equal(t, group, item.Group)
		assert.Equal(t, kind, item.Kind)
	}
}

func TestClearCache(t *testing.T) {
	ctx := context.Background()
	r := newReader()
	w := NewSQLWriter(testQueries)

	created := makeResource(t, w, "test", "test_kind", uniqueName("clear"))

	// Populate the cache.
	_, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)

	require.NoError(t, r.ClearCache(ctx))

	// Still readable from the database after the cache is cleared.
	res, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, created.Id, res.Id)
}
