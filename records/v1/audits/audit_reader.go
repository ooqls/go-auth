package audits

// import (
// 	"context"
// 	"errors"

// 	"github.com/eko/gocache/lib/v4/cache"
// 	"github.com/elastic/go-elasticsearch/v9"
// 	"github.com/ooqls/go-auth/records/gen"
// 	"github.com/ooqls/go-log"
// 	"github.com/redis/go-redis/v9"
// 	"go.uber.org/zap"
// )

// type AuditReader interface {
// 	GetAuditByUserID(ctx context.Context, userID string) ([]Audit, error)
// 	GetAuditByDomain(ctx context.Context, domain string) ([]Audit, error)
// 	GetAuditByResourceOrn(ctx context.Context, resourceOrn string) ([]Audit, error)
// }

// type CachedAuditReader struct {
// 	r     AuditReader
// 	cache *cache.Cache[[]Audit]
// 	l     *zap.Logger
// }

// func NewCachedAuditReader(r AuditReader, cache *cache.Cache[[]Audit]) *CachedAuditReader {
// 	return &CachedAuditReader{
// 		r:     r,
// 		cache: cache,
// 		l:     log.NewLogger("CachedSQLAuditReader"),
// 	}
// }

// func (r *CachedAuditReader) GetAuditByUserID(ctx context.Context, userID string) ([]Audit, error) {
// 	cachedAudit, err := r.cache.Get(ctx, userID)
// 	if err == nil && len(cachedAudit) > 0 {
// 		return cachedAudit, nil
// 	}
// 	if err != nil && !errors.Is(err, redis.Nil) {
// 		r.l.Warn("something went wrong when accessing cache", zap.Error(err))
// 	}
// 	// Cache miss, fetch from the database
// 	audit, err := r.r.GetAuditByUserID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = r.cache.Set(ctx, userID, audit)
// 	if err != nil {
// 		r.l.Warn("failed to set audit in cache", zap.Error(err))
// 	}

// 	return audit, nil
// }

// func (r *CachedAuditReader) GetAuditByDomain(ctx context.Context, domain string) ([]Audit, error) {
// 	cachedAudit, err := r.cache.Get(ctx, domain)
// 	if err == nil && len(cachedAudit) > 0 {
// 		return cachedAudit, nil
// 	}
// 	if err != nil && !errors.Is(err, redis.Nil) {
// 		r.l.Warn("something went wrong when accessing cache", zap.Error(err))
// 	}
// 	// Cache miss, fetch from the database
// 	audit, err := r.r.GetAuditByDomain(ctx, domain)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = r.cache.Set(ctx, domain, audit)
// 	if err != nil {
// 		r.l.Warn("failed to set audit in cache", zap.Error(err))
// 	}

// 	return audit, nil
// }

// func (r *CachedAuditReader) GetAuditByResourceOrn(ctx context.Context, resourceOrn string) ([]Audit, error) {
// 	cachedAudit, err := r.cache.Get(ctx, resourceOrn)
// 	if err == nil && len(cachedAudit) > 0 {
// 		return cachedAudit, nil
// 	}
// 	if err != nil && !errors.Is(err, redis.Nil) {
// 		r.l.Warn("something went wrong when accessing cache", zap.Error(err))
// 	}
// 	// Cache miss, fetch from the database
// 	audit, err := r.r.GetAuditByResourceOrn(ctx, resourceOrn)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = r.cache.Set(ctx, resourceOrn, audit)
// 	if err != nil {
// 		r.l.Warn("failed to set audit in cache", zap.Error(err))
// 	}

// 	return audit, nil
// }

// type DocumentAuditReader struct {
// 	cli elasticsearch.TypedClient
// 	l   *zap.Logger
// }

// func NewDocumentAuditReader(cli elasticsearch.TypedClient) *DocumentAuditReader {
// 	return &DocumentAuditReader{
// 		cli: cli,
// 		l:   log.NewLogger("DocumentAuditReader"),
// 	}
// }

// func (r *DocumentAuditReader) GetAuditByUserID(ctx context.Context, userID string) ([]Audit, error) {
// 	// Implement the logic to fetch audit logs from Elasticsearch
// 	// using the userID as a filter.
// 	// This is a placeholder implementation.
// 	s := r.cli.Search()
// 	s.Index()

// }

// type SQLAuditReader struct {
// 	q gen.Queries
// 	l *zap.Logger
// }

// func NewSQLAuditReader(q gen.Queries) *SQLAuditReader {
// 	return &SQLAuditReader{
// 		q: q,
// 		l: log.NewLogger("SQLAuditReader"),
// 	}
// }

// func (r *SQLAuditReader) GetAuditByUserID(ctx context.Context, userID string) ([]Audit, error) {
// 	audits, err := r.q.GetAuditByUserID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return audits, nil
// }

// func (r *SQLAuditReader) GetAuditByDomain(ctx context.Context, domain string) ([]Audit, error) {
// 	audits, err := r.q.GetAuditByDomain(ctx, domain)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return audits, nil
// }

// func (r *SQLAuditReader) GetAuditByResourceOrn(ctx context.Context, resourceOrn string) ([]Audit, error) {
// 	audits, err := r.q.GetAuditByResourceOrn(ctx, resourceOrn)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return audits, nil
// }
