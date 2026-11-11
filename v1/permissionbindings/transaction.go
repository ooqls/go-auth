package permissionbindings

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ooqls/getset/cache/cache"
	aggsdata "github.com/ooqls/go-auth/internal/aggsv1/datagen"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
	permissionsdata "github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	rolesdata "github.com/ooqls/go-auth/internal/rolesv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
	"go.uber.org/zap"
)

// transactionData provides uncached readers and writers on the same snapshot.
// Callers decide which operations are authorized before invoking the writers.
type transactionData struct {
	Bindings    *datagen.Queries
	Roles       *rolesdata.Queries
	Permissions *permissionsdata.Queries
	Aggregates  *aggsdata.Queries
}

type transaction interface {
	Run(ctx contexts.LContext, operation func(transactionData) error) error
}

type transactionBeginner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type sqlTransaction struct {
	db     transactionBeginner
	caches []*cache.GenericCache
}

func newSQLTransaction(db transactionBeginner, caches ...*cache.GenericCache) *sqlTransaction {
	return &sqlTransaction{db: db, caches: caches}
}

// Run commits the entire operation or rolls it back. Serialization conflicts
// are returned so the caller can retry with a fresh snapshot. No authorization
// policy or requester identity is interpreted here.
func (t *sqlTransaction) Run(ctx contexts.LContext, operation func(transactionData) error) error {
	tx, err := t.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return v1.ErrInternal(err, nil)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()

	if err := operation(transactionData{
		Bindings:    datagen.New(tx),
		Roles:       rolesdata.New(tx),
		Permissions: permissionsdata.New(tx),
		Aggregates:  aggsdata.New(tx),
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return v1.ErrInternal(err, nil)
	}

	// Invalidate derived data only after commit. Cache failure cannot undo a
	// committed write and must not make a successful operation look rolled back.
	cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	for _, c := range t.caches {
		if c != nil {
			if err := c.Clear(cacheCtx); err != nil {
				ctx.L().Warn("failed to clear permission-binding derived cache", zap.Error(err))
			}
		}
	}
	return nil
}
