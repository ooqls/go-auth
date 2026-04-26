package resources

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
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
)

func newTestService(
	rr resourcesv1.Reader,
	rw resourcesv1.Writer,
	ra authorizationv1.Authorizer,
) *ServiceImpl {
	return &ServiceImpl{rr: rr, rw: rw, ra: ra}
}

func newAuthCtx(user usersv1.User) *authorizationv1.Context {
	ctx := authorizationv1.NewAuthorizationContext(user)
	return &ctx
}

func authAllowed(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformAction(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
}

func authDenied(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformAction(gomock.Any(), gomock.Any(), gomock.Any()).
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

	t.Run("group and kind — dispatches to GetResourcesByGroupAndKind", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		resources := []resourcesv1.Resourcev1{
			*sampleResource("res1", "mygroup", "mykind"),
			*sampleResource("res2", "mygroup", "mykind"),
		}
		authAllowed(ra)
		rr.EXPECT().GetResourcesByGroupAndKind(gomock.Any(), "mygroup", "mykind", int32(10), int32(0)).Return(resources, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), strPtr("mykind"), 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("group only — dispatches to GetResourcesByGroup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		resources := []resourcesv1.Resourcev1{
			*sampleResource("res1", "mygroup", "kind1"),
		}
		authAllowed(ra)
		rr.EXPECT().GetResourcesByGroup(gomock.Any(), "mygroup", int32(10), int32(0)).Return(resources, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("no filter — dispatches to GetResources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		resources := []resourcesv1.Resourcev1{
			*sampleResource("res1", "g1", "k1"),
		}
		authAllowed(ra)
		rr.EXPECT().GetResources(gomock.Any(), int32(10), int32(0)).Return(resources, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 1)
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
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		rr.EXPECT().GetResourcesByGroupAndKind(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResources(ctx, strPtr("mygroup"), strPtr("mykind"), 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}
