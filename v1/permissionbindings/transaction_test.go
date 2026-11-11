package permissionbindings

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTransaction struct {
	pgx.Tx
	commitErr error
	commits   int
	rollbacks int
}

func (tx *testTransaction) Commit(context.Context) error {
	tx.commits++
	return tx.commitErr
}

func (tx *testTransaction) Rollback(context.Context) error {
	tx.rollbacks++
	return nil
}

type testTransactionBeginner struct {
	tx  *testTransaction
	err error
}

func (b testTransactionBeginner) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return b.tx, b.err
}

func TestSQLTransaction_Run(t *testing.T) {
	t.Parallel()
	failure := errors.New("database failure")
	denied := v1.ErrPermissionDenied(errors.New("denied"), nil)
	for _, tc := range []struct {
		name                       string
		beginErr, operationErr     error
		commitErr                  error
		wantOperations, wantCommit int
	}{
		{name: "success", wantOperations: 1, wantCommit: 1},
		{name: "begin failure", beginErr: failure},
		{name: "typed operation failure", operationErr: denied, wantOperations: 1},
		{name: "commit failure", commitErr: failure, wantOperations: 1, wantCommit: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tx := &testTransaction{commitErr: tc.commitErr}
			ctx := authorizationv1.NewInternalOperationContext(context.Background())
			runner := newSQLTransaction(testTransactionBeginner{tx: tx, err: tc.beginErr})
			operations := 0
			err := runner.Run(&ctx, func(transactionData) error {
				operations++
				return tc.operationErr
			})
			assert.Equal(t, tc.wantOperations, operations)
			assert.Equal(t, tc.wantCommit, tx.commits)
			if tc.beginErr == nil {
				assert.Equal(t, 1, tx.rollbacks, "every opened transaction must be closed")
			} else {
				assert.Zero(t, tx.rollbacks)
			}
			switch {
			case tc.operationErr != nil:
				assert.Same(t, tc.operationErr, err)
			case tc.beginErr != nil || tc.commitErr != nil:
				var internal *v1.InternalError
				require.ErrorAs(t, err, &internal)
				require.ErrorIs(t, err, failure)
			default:
				require.NoError(t, err)
			}
		})
	}
}
