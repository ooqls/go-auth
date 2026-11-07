package resources

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	authmocks "github.com/ooqls/go-auth/internal/authorizationv1/mocks"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	resmocks "github.com/ooqls/go-auth/internal/resourcesv1/mocks"
	"github.com/ooqls/go-auth/internal/usersv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestService(
	rr resourcesv1.Reader,
	rw resourcesv1.Writer,
	ra authorizationv1.Authorizer,
) *ServiceImpl {
	return &ServiceImpl{rr: rr, rw: rw, ra: ra}
}

func newAuthCtx(user usersv1.User) *authorizationv1.Context {
	ctx := authorizationv1.NewAuthorizationContext(context.Background(), user)
	return &ctx
}

func authAllowed(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformResourceAction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
}

func authDenied(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformResourceAction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(authorizationv1.ErrPermissionDenied).
		AnyTimes()
}

func sampleUser(name string) usersv1.User {
	return usersv1.NewUser(uuid.New(), name, "test@test.com", "", "")
}

func sampleResource(name, group, kind string) *resourcesv1.Resourcev1 {
	return &resourcesv1.Resourcev1{
		Metadata: corev1.Metadata{Group: group, Kind: kind},
		Id:       uuid.New(),
		Name:     name,
	}
}

func assertForbidden(t *testing.T, err error) {
	t.Helper()
	var target *v1.ForbiddenError
	assert.ErrorAs(t, err, &target)
}

func strPtr(s string) *string { return &s }

func TestCreateResource(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — creates resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		res := sampleResource("myresource", "mygroup", "mykind")
		authAllowed(ra)
		rw.EXPECT().CreateResource(gomock.Any(), "mygroup", "mykind", "myresource").Return(res, nil)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource")
		require.NoError(t, err)
		assert.Equal(t, res, got)
	})

	t.Run("unauthorized — permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rw.EXPECT().CreateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetResource(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — returns resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		res := sampleResource("myresource", "mygroup", "mykind")
		authAllowed(ra)
		rr.EXPECT().GetResource(gomock.Any(), "mygroup", "mykind", "myresource").Return(res, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResource(ctx, "mygroup", "mykind", "myresource")
		require.NoError(t, err)
		assert.Equal(t, res, got)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().GetResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestUpdateResourceName(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — updates resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := sampleResource("oldname", "mygroup", "mykind")
		updated := sampleResource("newname", "mygroup", "mykind")

		authAllowed(ra)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResource(gomock.Any(), "mygroup", "mykind", "oldname").Return(existing, nil)
		rw.EXPECT().UpdateResource(gomock.Any(), "mygroup", "mykind", "oldname", "newname").Return(updated, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, "mygroup", "mykind", "oldname", "newname")
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, "mygroup", "mykind", "oldname", "newname")
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("resource not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResource(gomock.Any(), "mygroup", "mykind", "oldname").Return(nil, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, "mygroup", "mykind", "oldname", "newname")
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := sampleResource("oldname", "mygroup", "mykind")

		authAllowed(ra)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(existing, nil)
		rw.EXPECT().UpdateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, "mygroup", "mykind", "oldname", "newname")
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestDeleteResource(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — deletes resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResource(gomock.Any(), "mygroup", "mykind", "myresource").Return(nil)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResource(ctx, "mygroup", "mykind", "myresource")
		require.NoError(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assertForbidden(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
	})
}

func TestGetResources(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	makeResult := func(items ...resourcesv1.Resourcev1) *corev1.Result[[]resourcesv1.Resourcev1] {
		return &corev1.Result[[]resourcesv1.Resourcev1]{Items: items, TotalCount: int64(len(items))}
	}

	t.Run("group and kind — preserves both filters", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "mygroup", "mykind").Return(nil)
		rr.EXPECT().SearchResources(gomock.Any(), strPtr("mygroup"), strPtr("mykind"), nil, nil, int32(10), int32(0)).
			Return(makeResult(*sampleResource("res1", "mygroup", "mykind"), *sampleResource("res2", "mygroup", "mykind")), nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), strPtr("mykind"), 1, 10)
		require.NoError(t, err)
		assert.Len(t, got.Items, 2)
		assert.Equal(t, int64(2), got.TotalCount)
	})

	t.Run("group only — preserves group filter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "mygroup", "*").Return(nil)
		rr.EXPECT().SearchResources(gomock.Any(), strPtr("mygroup"), nil, nil, nil, int32(10), int32(0)).
			Return(makeResult(*sampleResource("res1", "mygroup", "kind1")), nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got.Items, 1)
	})

	t.Run("kind only — preserves authorized scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)
		kind := "mykind"
		result := makeResult(*sampleResource("res1", "group1", kind))
		gomock.InOrder(
			ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "*", kind).Return(nil),
			rr.EXPECT().SearchResources(ctx, nil, &kind, nil, nil, int32(10), int32(10)).Return(result, nil),
		)

		got, err := newTestService(rr, nil, ra).GetResources(ctx, nil, &kind, 2, 10)
		require.NoError(t, err)
		assert.Equal(t, result, got)
	})

	t.Run("kind only — denied before reading", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)
		kind := "mykind"
		ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "*", kind).
			Return(authorizationv1.ErrPermissionDenied)

		got, err := newTestService(rr, nil, ra).GetResources(ctx, nil, &kind, 1, 10)
		require.ErrorIs(t, err, authorizationv1.ErrPermissionDenied)
		assert.Nil(t, got)
	})

	t.Run("no filter — reads all resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "*", "*").Return(nil)
		rr.EXPECT().SearchResources(gomock.Any(), nil, nil, nil, nil, int32(10), int32(0)).
			Return(makeResult(*sampleResource("res1", "g1", "k1")), nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got.Items, 1)
	})

	t.Run("unauthorized — returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), strPtr("mykind"), 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().SearchResources(gomock.Any(), gomock.Any(), gomock.Any(), nil, nil, gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), strPtr("mykind"), 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestSearchResources(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                    string
		group, kind             *string
		targetGroup, targetKind string
	}{
		{name: "group and kind", group: strPtr("group"), kind: strPtr("kind"), targetGroup: "group", targetKind: "kind"},
		{name: "group only", group: strPtr("group"), targetGroup: "group", targetKind: "*"},
		{name: "kind only", kind: strPtr("kind"), targetGroup: "*", targetKind: "kind"},
		{name: "no filters", targetGroup: "*", targetKind: "*"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, allowed := range []bool{true, false} {
				name := "denied"
				if allowed {
					name = "allowed"
				}
				t.Run(name, func(t *testing.T) {
					ctrl := gomock.NewController(t)
					rr := resmocks.NewMockReader(ctrl)
					ra := authmocks.NewMockAuthorizer(ctrl)
					ctx := newAuthCtx(sampleUser("searcher"))
					resourceName, query := "resource", "query"
					auth := ra.EXPECT().IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, tc.targetGroup, tc.targetKind)
					result := &corev1.Result[[]resourcesv1.Resourcev1]{}
					if allowed {
						auth.Return(nil)
						rr.EXPECT().SearchResources(ctx, tc.group, tc.kind, &resourceName, &query, int32(defaultSearchPageSize), int32(0)).
							After(auth).Return(result, nil)
					} else {
						auth.Return(authorizationv1.ErrPermissionDenied)
					}
					got, err := newTestService(rr, nil, ra).SearchResources(ctx, tc.group, tc.kind, &resourceName, &query)
					if allowed {
						require.NoError(t, err)
						assert.Equal(t, result, got)
					} else {
						require.ErrorIs(t, err, authorizationv1.ErrPermissionDenied)
						assertForbidden(t, err)
						assert.Nil(t, got)
					}
				})
			}
		})
	}
}
