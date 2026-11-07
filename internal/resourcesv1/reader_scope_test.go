//go:build integration

package resourcesv1

import (
	"context"
	"testing"

	"github.com/ooqls/getset/cache/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetResource_Scope(t *testing.T) {
	ctx := context.Background()
	w := NewSQLWriter(testQueries)
	group, kind, name := uniqueName("scope"), uniqueName("kind"), uniqueName("name")
	created := makeResource(t, w, group, kind, name)
	otherGroup := makeResource(t, w, group+"-other", kind, name)
	otherKind := makeResource(t, w, group, kind+"-other", name)

	for _, tc := range []struct {
		name string
		warm bool
	}{
		{name: "cold cache"},
		{name: "warm cache", warm: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newReader()
			if tc.warm {
				got, err := r.GetResource(ctx, group, kind, name)
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, created.Id, got.Id)
			}

			for _, want := range []*Resourcev1{otherGroup, otherKind, created} {
				got, err := r.GetResource(ctx, want.Group, want.Kind, name)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, want.Id, got.Id)
			}

			for _, scope := range []struct {
				name, group, kind string
			}{
				{name: "wrong group", group: group + "-missing", kind: kind},
				{name: "wrong kind", group: group, kind: kind + "-missing"},
			} {
				t.Run(scope.name, func(t *testing.T) {
					got, err := r.GetResource(ctx, scope.group, scope.kind, name)
					require.NoError(t, err)
					assert.Nil(t, got)
				})
			}
		})
	}
}

func TestGetResource_CacheIdentity(t *testing.T) {
	ctx := context.Background()
	w := NewSQLWriter(testQueries)
	created := makeResource(t, w, uniqueName("group"), "kind", uniqueName("resource"))
	// A resource name may be another resource's UUID, but must not share its cache entry.
	named := makeResource(t, w, created.Group, created.Kind, created.Id.String())
	r := newReader()
	got, err := r.GetResourceByID(ctx, created.Id)
	require.NoError(t, err)
	require.NotNil(t, got)
	got, err = r.GetResource(ctx, named.Group, named.Kind, named.Name)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, named.Id, got.Id)

	// Delimiters in group/kind must not make distinct scopes share cache entries.
	name := uniqueName("delimiter")
	first := makeResource(t, w, created.Group+":kind:x", "y", name)
	second := makeResource(t, w, created.Group, "x:kind:y", name)
	for _, want := range []*Resourcev1{first, second, first} {
		got, err := r.GetResource(ctx, want.Group, want.Kind, name)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, want.Id, got.Id)
	}
}

func TestSearchResources_ScopeAndPagination(t *testing.T) {
	ctx := context.Background()
	w := NewSQLWriter(testQueries)
	group, kind := uniqueName("search-group"), uniqueName("search-kind")
	first := makeResource(t, w, group, kind, uniqueName("a"))
	second := makeResource(t, w, group+"-other", kind, uniqueName("b"))
	makeResource(t, w, group, kind+"-other", uniqueName("c"))

	for _, tc := range []struct {
		name  string
		group *string
		want  []Resourcev1
	}{
		{name: "kind only", want: []Resourcev1{*first, *second}},
		{name: "group and kind", group: &group, want: []Resourcev1{*first}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newReader()
			for _, cacheState := range []string{"cold", "warm"} {
				t.Run(cacheState, func(t *testing.T) {
					for offset, want := range tc.want {
						got, err := r.SearchResources(ctx, tc.group, &kind, nil, nil, 1, int32(offset))
						require.NoError(t, err)
						require.NotNil(t, got)
						require.Len(t, got.Items, 1)
						assert.Equal(t, want.Id, got.Items[0].Id)
						assert.Equal(t, int64(len(tc.want)), got.TotalCount)
					}
				})
			}
		})
	}
}

func TestSearchResources_CacheIdentity(t *testing.T) {
	ctx := context.Background()
	w := NewSQLWriter(testQueries)
	group, kind := uniqueName("cache-group"), uniqueName("cache-kind")
	first := makeResource(t, w, group, kind+":kind:other", uniqueName("first"))
	second := makeResource(t, w, group+":"+kind, "kind:other", uniqueName("second"))
	// Share a cache across readers, as the factory does in production.
	shared := cache.NewGenericCache("scope-test", cache.NewMemCache())
	for _, want := range []*Resourcev1{first, second, first} {
		r := NewSQLReader(shared, testQueries)
		got, err := r.SearchResources(ctx, &want.Group, &want.Kind, nil, nil, 10, 0)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, want.Id, got.Items[0].Id)
	}

	// Missing filters and literal wildcard/empty filters are different requests.
	star := makeResource(t, w, "*", kind, uniqueName("star"))
	empty := makeResource(t, w, "", kind, uniqueName("empty"))
	r := NewSQLReader(shared, testQueries)
	all, err := r.SearchResources(ctx, nil, &kind, nil, nil, 100, 0)
	require.NoError(t, err)
	require.Len(t, all.Items, 2)
	for _, want := range []*Resourcev1{star, empty} {
		got, err := r.SearchResources(ctx, &want.Group, &kind, nil, nil, 100, 0)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, want.Id, got.Items[0].Id)
		assert.Equal(t, int64(1), got.TotalCount)
	}
}
