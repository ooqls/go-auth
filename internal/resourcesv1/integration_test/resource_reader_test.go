package integrationTest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	"github.com/stretchr/testify/assert"
)

func createResource(t *testing.T, group, kind string) []resourcesv1.Resourcev1 {
	var res []resourcesv1.Resourcev1
	for i := 0; i < 10; i++ {
		res = append(res, resourcesv1.Resourcev1{
			Object: corev1.Object{
				Name: fmt.Sprintf("test%d", i),
				Metadata: corev1.Metadata{
					Group: group,
					Kind:  kind,
				},
			},
		})
	}
	return res
}

func populateDatabase(t *testing.T) []resourcesv1.Resourcev1 {
	conn := pgx.GetPGX()
	w := resourcesv1.NewSQLWriter(datagen.New(conn))
	ctx := context.Background()
	resObjs := createResource(t, "test", "object")
	for _, res := range resObjs {
		id := corev1.Object{
			Name: res.Name,
			Metadata: corev1.Metadata{
				Group: res.Group,
				Kind:  res.Kind,
			},
		}
		_, err := w.CreateResource(ctx, id.Group, id.Kind, id.Name, "test")
		assert.Nilf(t, err, "should not error when creating resource")
	}
	time.Sleep(time.Second)
	return resObjs
}

func TestResourceReaderWriter(t *testing.T) {
	resObjs := populateDatabase(t)
	conn := pgx.GetPGX()
	r := resourcesv1.NewSQLReader(cache.NewGenericCache("test", cache.NewMemCache()), datagen.New(conn))

	t.Run("GetResource", func(t *testing.T) {
		ctx := context.Background()
		for _, res := range resObjs {
			retrieved, err := r.GetResource(ctx,
				res.Name,
				res.Metadata,
			)

			assert.Nilf(t, err, "should not error when getting resource")
			assert.NotNilf(t, retrieved, "should not return nil")
			if retrieved != nil && err == nil {
				assert.Equal(t, res.Name, retrieved.Name, "name should be the same")
				assert.Equal(t, res.Metadata.Kind, retrieved.Metadata.Kind, "kind should be the same")
				assert.Equal(t, res.Metadata.Group, retrieved.Metadata.Group, "group should be the same")
			}
		}
	})

	t.Run("GetResources", func(t *testing.T) {
		ctx := context.Background()
		retr, err := r.GetResources(ctx, corev1.Metadata{
			Kind:  "object",
			Group: "test",
		}, 10, 0)
		assert.Nilf(t, err, "should not error when querying resources")
		assert.Equal(t, len(resObjs), len(retr), "should return some resources")

		retr, err = r.GetResources(ctx, corev1.Metadata{
			Kind:  "object",
			Group: "test",
		}, 10, 10)
		assert.Nilf(t, err, "should not error when querying resources")
		assert.Equal(t, 0, len(retr), "should return no resources")

		retr, err = r.GetResources(ctx, corev1.Metadata{
			Group: "group",
			Kind:  "unknown",
		}, 10, 0)
	})
}
