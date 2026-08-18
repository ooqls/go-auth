package resourcesv1

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/google/uuid"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/corev1"
	v1gen "github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/resources/api/gen_resources"
	"go.uber.org/zap"
)

const (
	defaultPage     int32 = 1
	defaultPageSize int32 = 100
)

type Param struct {
	Page     int32
	PageSize int32
	Text     string
}

type Query interface {
	SearchResources(ctx context.Context, q *Param) (*corev1.Result[gen_resources.ResourceList], error)
}

// resourceDocument mirrors the shape of a resource document indexed in
// Elasticsearch.
type resourceDocument struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Group     string    `json:"group"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ElasticsearchQuery struct {
	Client *elasticsearch.TypedClient
	index  string
	l      *zap.Logger
}

func NewElasticsearchQuery(client *elasticsearch.TypedClient, index string) *ElasticsearchQuery {
	return &ElasticsearchQuery{
		Client: client,
		index:  index,
		l:      log.NewLogger("ElasticsearchQuery"),
	}
}

func (q *ElasticsearchQuery) SearchResources(ctx context.Context, qry *Param) (*corev1.Result[gen_resources.ResourceList], error) {
	page := qry.Page
	if page < 1 {
		page = defaultPage
	}

	pageSize := qry.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	var esQuery types.Query
	if qry.Text == "" {
		esQuery = types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	} else {
		esQuery = types.Query{
			MultiMatch: &types.MultiMatchQuery{
				Query:  qry.Text,
				Fields: []string{"name", "group", "kind"},
			},
		}
	}

	s := q.Client.API.Search()
	s.Index(q.index)
	s.Query(&esQuery)
	s.From(int((page - 1) * pageSize))
	s.Size(int(pageSize))

	res, err := s.Do(ctx)
	if err != nil {
		q.l.Error("failed to search resources", zap.Error(err))
		return nil, err
	}

	items := make([]v1gen.Resource, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var doc resourceDocument
		if err := json.Unmarshal(hit.Source_, &doc); err != nil {
			q.l.Error("failed to unmarshal resource document", zap.Error(err))
			return nil, err
		}

		id := doc.ID
		updatedAt := doc.UpdatedAt
		items = append(items, v1gen.Resource{
			Id:        &id,
			Name:      doc.Name,
			Group:     doc.Group,
			Kind:      doc.Kind,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: &updatedAt,
		})
	}

	var total int64
	if res.Hits.Total != nil {
		total = res.Hits.Total.Value
	}

	return &corev1.Result[gen_resources.ResourceList]{
		Items: gen_resources.ResourceList{
			Items:      items,
			TotalCount: int(total),
		},
		TotalCount: total,
	}, nil
}
