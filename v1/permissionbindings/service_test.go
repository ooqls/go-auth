package permissionbindings

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMutator struct {
	mutate func(*authorizationv1.Context, []uuid.UUID, []string) error
}

func (m testMutator) AssignPermissions(ctx *authorizationv1.Context, roles []uuid.UUID, permissions []string) error {
	return m.mutate(ctx, roles, permissions)
}

func (m testMutator) UnassignPermissions(ctx *authorizationv1.Context, roles []uuid.UUID, permissions []string) error {
	return m.mutate(ctx, roles, permissions)
}

func TestBindingMutations(t *testing.T) {
	t.Parallel()
	for _, operation := range []struct {
		name string
		call func(*ServiceImpl, *authorizationv1.Context, []uuid.UUID, []string) error
	}{
		{name: "assign", call: (*ServiceImpl).AssignPermission},
		{name: "unassign", call: (*ServiceImpl).UnassignPermission},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			for _, tc := range []struct {
				name string
				err  error
				code int
			}{
				{name: "success"},
				{name: "forbidden", err: v1.ErrPermissionDenied(errors.New("denied"), nil), code: 403},
				{name: "missing role", err: v1.ErrNotFound(errors.New("missing role"), nil), code: 404},
				{name: "storage failure", err: v1.ErrInternal(errors.New("database failure"), nil), code: 500},
				{name: "existing internal error", err: v1.ErrInternal(errors.New("transaction failure"), nil), code: 500},
			} {
				t.Run(tc.name, func(t *testing.T) {
					ctx := authorizationv1.NewInternalOperationContext(context.Background())
					roles := []uuid.UUID{uuid.New(), uuid.New()}
					permissions := []string{"first", "second"}
					calls := 0
					s := &ServiceImpl{m: testMutator{mutate: func(gotCtx *authorizationv1.Context, gotRoles []uuid.UUID, gotPermissions []string) error {
						calls++
						assert.Same(t, &ctx, gotCtx)
						assert.Equal(t, roles, gotRoles)
						assert.Equal(t, permissions, gotPermissions)
						return tc.err
					}}}
					err := operation.call(s, &ctx, roles, permissions)
					require.Equal(t, 1, calls, "the entire batch must cross the mutation seam once")
					if tc.err == nil {
						require.NoError(t, err)
						return
					}
					require.ErrorIs(t, err, tc.err)
					assert.Same(t, tc.err, err, "the mutator's typed error must be returned unchanged")
					switch tc.code {
					case 403:
						var target *v1.ForbiddenError
						require.ErrorAs(t, err, &target)
					case 404:
						var target *v1.NotFoundError
						require.ErrorAs(t, err, &target)
					case 500:
						var target *v1.InternalError
						require.ErrorAs(t, err, &target)
					}
				})
			}
		})
	}
}
