package resourcesv1

//go:generate go run go.uber.org/mock/mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	"go.uber.org/zap"
)

func getCacheKey(name string) string {
	return fmt.Sprintf("resource-%s", name)
}

func getPaginatedKey(limit, offset int32) string {
	return fmt.Sprintf("resource-page-%d-%d", limit, offset)
}

type Reader interface {
	GetResourceByID(ctx context.Context, id uuid.UUID) (*Resourcev1, error)
	GetResource(ctx context.Context, group, kind, name string) (*Resourcev1, error)
	GetResources(ctx context.Context, limit, offset int32) (*corev1.Result[[]Resourcev1], error)
	GetResourcesByGroup(ctx context.Context, group string, limit, offset int32) (*corev1.Result[[]Resourcev1], error)
	GetResourcesByGroupAndKind(ctx context.Context, group, kind string, limit, offset int32) (*corev1.Result[[]Resourcev1], error)
	SearchResources(ctx context.Context, group, kind, name, query *string, limit, offset int32) (*corev1.Result[[]Resourcev1], error)
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

func (r *SQLReader) GetResource(ctx context.Context, group, kind, name string) (*Resourcev1, error) {
	cacheKey := getCacheKey(name)
	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &cached[0], nil
	}

	res, err := r.q.GetResourceByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get resource from database: %v", err)
	}

	resObj := FromDatagenResource(res)

	r.setCache(ctx, cacheKey, *resObj)
	return resObj, nil
}

func (r *SQLReader) GetResourceByID(ctx context.Context, id uuid.UUID) (*Resourcev1, error) {
	cacheKey := getCacheKey(id.String())
	if cached := r.getCache(ctx, cacheKey); cached != nil {
		return &cached[0], nil
	}

	res, err := r.q.GetResourceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get resource from database: %v", err)
	}

	resObj := FromDatagenResource(res)

	r.setCache(ctx, cacheKey, *resObj)
	return resObj, nil
}

func (r *SQLReader) GetResources(ctx context.Context, limit, offset int32) (*corev1.Result[[]Resourcev1], error) {
	cacheKey := getPaginatedKey(limit, offset)
	if res := r.getCache(ctx, cacheKey); res != nil {
		total, err := r.q.CountAllResources(ctx)
		if err != nil {
			return nil, err
		}
		return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
	}

	qres, err := r.q.GetResources(ctx, datagen.GetResourcesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		r.l.Error("failed to list resources", zap.Error(err))
		return nil, err
	}

	total, err := r.q.CountAllResources(ctx)
	if err != nil {
		r.l.Error("failed to count resources", zap.Error(err))
		return nil, err
	}

	res := FromDatagenResourceList(qres)

	r.setCache(ctx, cacheKey, res...)
	return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
}

func (r *SQLReader) GetResourcesByGroup(ctx context.Context, group string, limit, offset int32) (*corev1.Result[[]Resourcev1], error) {
	cacheKey := fmt.Sprintf("group-%s:%s", group, getPaginatedKey(limit, offset))
	if res := r.getCache(ctx, cacheKey); res != nil {
		total, err := r.q.CountResourcesByGroup(ctx, group)
		if err != nil {
			return nil, err
		}
		return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
	}

	qres, err := r.q.GetResourcesByGroup(ctx, datagen.GetResourcesByGroupParams{
		Rgroup: group,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		r.l.Error("failed to list resources by group", zap.String("group", group), zap.Error(err))
		return nil, err
	}

	total, err := r.q.CountResourcesByGroup(ctx, group)
	if err != nil {
		r.l.Error("failed to count resources by group", zap.String("group", group), zap.Error(err))
		return nil, err
	}

	res := FromDatagenResourceList(qres)

	r.setCache(ctx, cacheKey, res...)
	return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
}

func (r *SQLReader) GetResourcesByGroupAndKind(ctx context.Context, group, kind string, limit, offset int32) (*corev1.Result[[]Resourcev1], error) {
	cacheKey := fmt.Sprintf("group-%s:kind-%s:%s", group, kind, getPaginatedKey(limit, offset))
	if res := r.getCache(ctx, cacheKey); res != nil {
		total, err := r.q.CountResources(ctx, datagen.CountResourcesParams{Rgroup: group, Kind: kind})
		if err != nil {
			return nil, err
		}
		return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
	}

	qres, err := r.q.GetResourcesByGroupAndKind(ctx, datagen.GetResourcesByGroupAndKindParams{
		Rgroup: group,
		Kind:   kind,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		r.l.Error("failed to list resources by group and kind", zap.String("group", group), zap.String("kind", kind), zap.Error(err))
		return nil, err
	}

	total, err := r.q.CountResources(ctx, datagen.CountResourcesParams{Rgroup: group, Kind: kind})
	if err != nil {
		r.l.Error("failed to count resources by group and kind", zap.String("group", group), zap.String("kind", kind), zap.Error(err))
		return nil, err
	}

	res := FromDatagenResourceList(qres)

	r.setCache(ctx, cacheKey, res...)
	return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
}

func (r *SQLReader) SearchResources(ctx context.Context,
	group,
	kind,
	name,
	query *string,
	limit, offset int32) (*corev1.Result[[]Resourcev1], error) {
	cacheKey := "search_resources"
	keys := []*string{group, kind, name, query}
	for _, key := range keys {
		if key != nil {
			cacheKey = fmt.Sprintf("%s:%s", cacheKey, *key)
		} else {
			cacheKey += ":*"
		}
	}

	cached := r.getCache(ctx, cacheKey)
	if cached != nil {
		total, err := r.q.CountSearchResources(ctx, datagen.CountSearchResourcesParams{
			Rgroup: toPgText(group),
			Kind:   toPgText(kind),
			Name:   toPgText(name),
			Query:  toPgText(query),
		})
		if err != nil {
			return nil, err
		}
		return &corev1.Result[[]Resourcev1]{Items: cached, TotalCount: total}, nil
	}

	qres, err := r.q.SearchResources(ctx, datagen.SearchResourcesParams{
		Rgroup: toPgText(group),
		Kind:   toPgText(kind),
		Name:   toPgText(name),
		Query:  toPgText(query),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		r.l.Error("failed to search resources", zap.Error(err))
		return nil, err
	}

	total, err := r.q.CountSearchResources(ctx, datagen.CountSearchResourcesParams{
		Rgroup: toPgText(group),
		Kind:   toPgText(kind),
		Name:   toPgText(name),
		Query:  toPgText(query),
	})
	if err != nil {
		r.l.Error("failed to count search resources", zap.Error(err))
		return nil, err
	}

	res := FromDatagenResourceList(qres)

	r.setCache(ctx, cacheKey, res...)
	return &corev1.Result[[]Resourcev1]{Items: res, TotalCount: total}, nil
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

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
