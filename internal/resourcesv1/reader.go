package resourcesv1

//go:generate go run github.com/golang/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	"go.uber.org/zap"
)

func getCacheKey(name string, metadata corev1.Metadata) string {
	return fmt.Sprintf("%s-%s-%s", name, metadata.Group, metadata.Kind)
}

func getPaginatedKey(metadata corev1.Metadata, offset, limit int32) string {
	return fmt.Sprintf("page-%s-%s-%d-%d", metadata.Group, metadata.Kind, limit, offset)
}

type Reader interface {
	GetResourceByID(ctx context.Context, id uuid.UUID) (*Resourcev1, error)
	GetResource(ctx context.Context, name string, object corev1.Metadata) (*Resourcev1, error)
	GetResources(ctx context.Context, object corev1.Metadata, limit, offset int32) ([]Resourcev1, error)
	ClearCache(ctx context.Context) error
}

type SQLReader struct {
	l *zap.Logger
	q *datagen.Queries
	c *cache.GenericCache
}

func NewSQLReader(c *cache.GenericCache, queries *datagen.Queries) Reader {
	return &SQLReader{
		l: log.NewLogger("resource-reader"),
		q: queries,
		c: c,
	}

}

func (r *SQLReader) GetResource(ctx context.Context, name string, obj corev1.Metadata) (*Resourcev1, error) {
	cacheKey := getCacheKey(name, obj)
	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &cached[0], nil
	}

	res, err := r.q.GetResourceByName(ctx, datagen.GetResourceByNameParams{
		ResourceName:  name,
		ResourceGroup: obj.Group,
		ResourceKind:  obj.Kind,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get resource from database: %v", err)
	}

	resObj := &Resourcev1{
		Object: corev1.Object{
			Metadata: corev1.Metadata{
				Group: res.ResourceGroup,
				Kind:  res.ResourceKind,
			},
			Name: res.ResourceName,
		},
		Description: res.Description,
		CreatedAt:   res.CreatedAt.Time,
		UpdatedAt:   res.UpdatedAt.Time,
	}

	r.setCache(ctx, cacheKey, *resObj)
	return resObj, nil
}

func (r *SQLReader) GetResourceByID(ctx context.Context, id uuid.UUID) (*Resourcev1, error) {
	cacheKey := getCacheKey(id.String(), Metadata)
	if cached := r.getCache(ctx, id.String()); cached != nil {
		return &cached[0], nil
	}

	res, err := r.q.GetResourceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get resource from database: %v", err)
	}

	resObj := &Resourcev1{
		Object: corev1.Object{
			Metadata: corev1.Metadata{
				Group: res.ResourceGroup,
				Kind:  res.ResourceKind,
			},
			Name: res.ResourceName,
		},
		Description: res.Description,
		CreatedAt:   res.CreatedAt.Time,
		UpdatedAt:   res.UpdatedAt.Time,
	}

	r.setCache(ctx, cacheKey, *resObj)
	return resObj, nil
}

func (r *SQLReader) GetResources(ctx context.Context, obj corev1.Metadata, limit, offset int32) ([]Resourcev1, error) {
	cacheKey := getPaginatedKey(corev1.Metadata{Group: obj.Group, Kind: obj.Kind}, offset, limit)
	res := r.getCache(ctx, cacheKey)
	if res != nil {
		return res, nil
	}

	kind := obj.Kind
	if kind == "" {
		kind = "*"
	}
	qres, err := r.q.GetResources(ctx, datagen.GetResourcesParams{
		ResourceGroup: obj.Group,
		Column2:       kind,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		r.l.Error("failed to list resources", zap.Error(err))
		return nil, err
	}

	res = make([]Resourcev1, len(qres))
	for i, obj := range qres {
		res[i] = Resourcev1{
			Object: corev1.Object{
				Metadata: corev1.Metadata{
					Group: obj.ResourceGroup,
					Kind:  obj.ResourceKind,
				},
				Name: obj.ResourceName,
			},
			Description: obj.Description,
			CreatedAt:   obj.CreatedAt.Time,
			UpdatedAt:   obj.UpdatedAt.Time,
		}
	}

	r.setCache(ctx, cacheKey, res...)
	return res, nil
}

// func (r *SQLReader) QueryResources(ctx context.Context, q Query) ([]records.Resource, error) {
// 	rows, err := r.q.QueryResources(ctx, gen.QueryResourcesParams{
// 		Column1: q.Name,
// 		Column2: q.Group,
// 		Column3: q.Kind,
// 		Limit:   q.PageSize,
// 		Offset:  q.Page * q.PageSize,
// 	})
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return []records.Resource{}, nil
// 		}

// 		return nil, fmt.Errorf("failed to query resources: %v", err)
// 	}

// 	return rows, nil
// }

func (r *SQLReader) ClearCache(ctx context.Context) error {
	err := r.c.Clear(ctx)
	if err != nil {
		r.l.Error("failed to clear cache", zap.Error(err))
		return err
	}
	return nil
}

func (r *SQLReader) getCache(ctx context.Context, key string) []Resourcev1 {
	if r.c == nil {
		return nil
	}

	var resources []Resourcev1
	err := r.c.Get(ctx, key, &resources)
	if err != nil {
		if !cache.IsCacheMissErr(err) {
			r.l.Error("failed to get cache", zap.Error(err))
		}
		return nil
	}

	return resources

}

func (r *SQLReader) setCache(ctx context.Context, key string, resources ...Resourcev1) {
	if r.c == nil {
		return
	}

	err := r.c.Set(ctx, key, resources)
	if err != nil {
		r.l.Error("failed to set cache", zap.Error(err))
	}
}
