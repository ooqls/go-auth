package permissions

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	authmocks "github.com/ooqls/go-auth/internal/authorizationv1/mocks"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	permmocks "github.com/ooqls/go-auth/internal/permissionsv1/mocks"
	"github.com/ooqls/go-auth/internal/usersv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(
	pr permissionsv1.Reader,
	pw permissionsv1.Writer,
	ra authorizationv1.Authorizer,
) *ServiceImpl {
	return &ServiceImpl{pr: pr, pw: pw, ra: ra}
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

func samplePermission(name string) permissionsv1.Permission {
	return permissionsv1.Permission{
		Object: corev1.Object{
			Metadata: permissionsv1.Metadata,
			Id:       uuid.New(),
			Name:     name,
		},
		Actions: "read",
	}
}

func sampleUser(name string) usersv1.User {
	return usersv1.NewUser(uuid.New(), name, "test@test.com", "", "")
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

func TestAddPermission(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — creates permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		pw.EXPECT().CreatePermission("myperm", "mygroup", "mykind", []string{"read"}).Return(nil)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm", "mygroup", "mykind", []string{"read"})
		require.NoError(t, err)
	})

	t.Run("unauthorized — permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm", "mygroup", "mykind", []string{"read"})
		require.Error(t, err)
		assertForbidden(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		writerErr := errors.New("db error")
		pw.EXPECT().CreatePermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(writerErr)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm", "mygroup", "mykind", []string{"read"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, writerErr))
	})
}

func TestDeletePermission(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — deletes permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		pw.EXPECT().DeletePermission("myperm", "mygroup", "mykind").Return(nil)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, "myperm", "mygroup", "mykind")
		require.NoError(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, "myperm", "mygroup", "mykind")
		require.Error(t, err)
		assertForbidden(t, err)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		writerErr := errors.New("db error")
		pw.EXPECT().DeletePermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(writerErr)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, "myperm", "mygroup", "mykind")
		require.Error(t, err)
		assert.True(t, errors.Is(err, writerErr))
	})
}

func TestGetPermission(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)
	perm := samplePermission("myperm")

	t.Run("authorized — returns permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		pr.EXPECT().GetPermission(ctx, "myperm", "mygroup", "mykind").Return(&perm, nil)
		authAllowed(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm", "mygroup", "mykind")
		require.NoError(t, err)
		assert.Equal(t, &perm, got)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		pr.EXPECT().GetPermission(ctx, "myperm", "mygroup", "mykind").Return(&perm, nil)
		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm", "mygroup", "mykind")
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		pr.EXPECT().GetPermission(ctx, "myperm", "mygroup", "mykind").Return(nil, nil)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm", "mygroup", "mykind")
		require.Error(t, err)
		assert.Nil(t, got)
		assertNotFound(t, err)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		readerErr := errors.New("db error")
		pr.EXPECT().GetPermission(ctx, "myperm", "mygroup", "mykind").Return(nil, readerErr)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm", "mygroup", "mykind")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, readerErr))
	})
}

func TestGetPermissions(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("all authorized — returns all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		perms := []permissionsv1.Permission{
			samplePermission("perm1"),
			samplePermission("perm2"),
		}
		pr.EXPECT().GetPermissions(ctx, 1, 10).Return(perms, nil)
		authAllowed(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("none authorized — returns empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		perms := []permissionsv1.Permission{
			samplePermission("perm1"),
			samplePermission("perm2"),
		}
		pr.EXPECT().GetPermissions(ctx, 1, 10).Return(perms, nil)
		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("partial — filters unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		allowed1 := samplePermission("allowed1")
		denied := samplePermission("denied")
		allowed2 := samplePermission("allowed2")
		perms := []permissionsv1.Permission{allowed1, denied, allowed2}

		pr.EXPECT().GetPermissions(ctx, 1, 10).Return(perms, nil)
		ra.EXPECT().
			IsAuthorizedToPerformAction(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(c *authorizationv1.Context, action authorizationv1.Action, target corev1.Object) error {
				if target.Name == "denied" {
					return authorizationv1.ErrPermissionDenied
				}
				return nil
			}).
			AnyTimes()

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got, 2)
		for _, p := range got {
			assert.NotEqual(t, "denied", p.Name)
		}
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		readerErr := errors.New("db error")
		pr.EXPECT().GetPermissions(ctx, 1, 10).Return(nil, readerErr)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, readerErr))
	})
}
