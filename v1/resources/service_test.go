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

func newInternalCtx() *authorizationv1.Context {
	ctx := authorizationv1.NewInternalOperationContext(nil)
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
		Object: corev1.Object{
			Metadata: corev1.Metadata{Group: group, Kind: kind},
			Id:       uuid.New(),
			Name:     name,
		},
		Description: "test description",
	}
}

func assertForbidden(t *testing.T, err error) {
	t.Helper()
	var target *v1.ForbiddenError
	assert.ErrorAs(t, err, &target)
}

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var target *v1.NotFoundError
	assert.ErrorAs(t, err, &target)
}

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
		rw.EXPECT().CreateResource(ctx, "mygroup", "mykind", "myresource", "desc").Return(res, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource", "desc")
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
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource", "desc")
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		writerErr := errors.New("db error")
		rw.EXPECT().CreateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, writerErr)

		svc := newTestService(rr, rw, ra)
		got, err := svc.CreateResource(ctx, "mygroup", "mykind", "myresource", "desc")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, writerErr))
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
		rr.EXPECT().GetResource(ctx, "myresource", corev1.Metadata{Group: "mygroup", Kind: "mykind"}).Return(res, nil)

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
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		readerErr := errors.New("db error")
		rr.EXPECT().GetResource(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, readerErr)

		svc := newTestService(rr, rw, ra)
		got, err := svc.GetResource(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, readerErr))
	})
}

func TestUpdateResourceName(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)
	id := uuid.New()
	newName := "newname"
	newDesc := "new desc"

	t.Run("authorized — updates resource", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := sampleResource("oldname", "mygroup", "mykind")
		existing.Id = id
		updated := sampleResource(newName, "mygroup", "mykind")
		updated.Id = id

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResourceByID(ctx, id).Return(existing, nil)
		authAllowed(ra)
		rw.EXPECT().UpdateResource(ctx, id, &newName, &newDesc).Return(updated, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, id, newName, &newDesc)
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("resource not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResourceByID(ctx, id).Return(nil, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, id, newName, &newDesc)
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := sampleResource("oldname", "mygroup", "mykind")
		existing.Id = id

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResourceByID(ctx, id).Return(existing, nil)
		authDenied(ra)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, id, newName, &newDesc)
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := sampleResource("oldname", "mygroup", "mykind")
		existing.Id = id
		writerErr := errors.New("db error")

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rr.EXPECT().GetResourceByID(ctx, id).Return(existing, nil)
		authAllowed(ra)
		rw.EXPECT().UpdateResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, writerErr)

		svc := newTestService(rr, rw, ra)
		got, err := svc.UpdateResourceName(ctx, id, newName, &newDesc)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, writerErr))
	})
}

func TestDeleteResource(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)
	id := uuid.New()

	t.Run("deletes by id", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResourceById(ctx, id).Return(nil)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResource(ctx, id)
		require.NoError(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		writerErr := errors.New("db error")
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResourceById(ctx, id).Return(writerErr)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResource(ctx, id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, writerErr))
	})
}

func TestDeleteResourceByName(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("deletes by name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResource(ctx, "mygroup", "mykind", "myresource").Return(nil)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResourceByName(ctx, "mygroup", "mykind", "myresource")
		require.NoError(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		writerErr := errors.New("db error")
		rr.EXPECT().ClearCache(gomock.Any()).Return(nil)
		rw.EXPECT().DeleteResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(writerErr)

		svc := newTestService(rr, rw, ra)
		err := svc.DeleteResourceByName(ctx, "mygroup", "mykind", "myresource")
		require.Error(t, err)
		assert.True(t, errors.Is(err, writerErr))
	})
}

func TestListResources(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("returns resources", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		resources := []resourcesv1.Resourcev1{
			*sampleResource("res1", "mygroup", "mykind"),
			*sampleResource("res2", "mygroup", "mykind"),
		}
		rr.EXPECT().GetResources(ctx, corev1.Metadata{Group: "mygroup", Kind: "mykind"}, int32(10), int32(0)).Return(resources, nil)

		svc := newTestService(rr, rw, ra)
		got, err := svc.ListResources(ctx, "mygroup", "mykind", 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		rr := resmocks.NewMockReader(ctrl)
		rw := resmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		readerErr := errors.New("db error")
		rr.EXPECT().GetResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, readerErr)

		svc := newTestService(rr, rw, ra)
		got, err := svc.ListResources(ctx, "mygroup", "mykind", 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, readerErr))
	})
}
