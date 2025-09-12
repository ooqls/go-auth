package authv1

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ooqls/go-log"
	"go.uber.org/zap"
)

type AuditWriter interface {
	CreateAudit(ctx context.Context, audit Audit) error
}

type ElasticsearchAuditWriter struct {
	c     *elasticsearch.TypedClient
	index string
	l     *zap.Logger
}

func NewElasticsearchAuditWriter(c *elasticsearch.TypedClient, index string) *ElasticsearchAuditWriter {
	return &ElasticsearchAuditWriter{
		c:     c,
		index: index,
		l:     log.NewLogger("ElasticsearchAuditWriter"),
	}
}

func (w *ElasticsearchAuditWriter) CreateAudit(ctx context.Context, audit Audit) error {
	// b, err := json.Marshal(audit)
	// if err != nil {
	// 	w.l.Error("failed to marshal audit", zap.Error(err))
	// 	return err
	// }

	_, err := w.c.Index(w.index).Document(audit).Do(ctx)
	if err != nil {
		w.l.Error("failed to create audit", zap.Error(err))
		return err
	}
	return nil
}
