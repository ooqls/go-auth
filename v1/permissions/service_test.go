package permissions

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
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
	ctx := authorizationv1.NewAuthorizationContext(context.Background(), user)
	return &ctx
}

func authAllowed(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformCoreAction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
}

func authDenied(mock *authmocks.MockAuthorizer) {
	mock.EXPECT().
		IsAuthorizedToPerformCoreAction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(authorizationv1.ErrPermissionDenied).
		AnyTimes()
}

func samplePermission(name string) permissionsv1.Permission {
	return permissionsv1.Permission{
		Metadata:   corev1.PermissionsV1,
		Permission: name,
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

		perm := samplePermission("myperm")
		authAllowed(ra)
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(nil, nil)
		pw.EXPECT().CreatePermission(gomock.Any(), "myperm").Return(&perm, nil)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm")
		require.NoError(t, err)
	})

	t.Run("unauthorized — permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm")
		require.Error(t, err)
		assertForbidden(t, err)
	})

	t.Run("already exists — returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		existing := samplePermission("myperm")
		authAllowed(ra)
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(&existing, nil)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm")
		require.Error(t, err)
		var target *v1.AlreadyExistsError
		assert.ErrorAs(t, err, &target)
	})

	t.Run("writer error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(nil, nil)
		writerErr := errors.New("db error")
		pw.EXPECT().CreatePermission(gomock.Any(), "myperm").Return(nil, writerErr)

		svc := newTestService(pr, pw, ra)
		err := svc.AddPermission(ctx, "myperm")
		require.Error(t, err)
	})
}

func TestDeletePermission(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)
	permission := "myperm"

	t.Run("authorized — deletes permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		pw.EXPECT().DeletePermission(gomock.Any(), permission).Return(nil)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, permission)
		require.NoError(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, permission)
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
		pw.EXPECT().DeletePermission(gomock.Any(), permission).Return(writerErr)

		svc := newTestService(pr, pw, ra)
		err := svc.DeletePermission(ctx, permission)
		require.Error(t, err)
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

		authAllowed(ra)
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(&perm, nil)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm")
		require.NoError(t, err)
		assert.Equal(t, &perm, got)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm")
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(nil, nil)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm")
		require.Error(t, err)
		assert.Nil(t, got)
		assertNotFound(t, err)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authAllowed(ra)
		readerErr := errors.New("db error")
		pr.EXPECT().GetPermission(gomock.Any(), "myperm").Return(nil, readerErr)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermission(ctx, "myperm")
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestGetPermissions(t *testing.T) {
	user := sampleUser("testuser")
	ctx := newAuthCtx(user)

	t.Run("authorized — returns all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		perms := []permissionsv1.Permission{
			samplePermission("perm1"),
			samplePermission("perm2"),
		}
		authAllowed(ra)
		pr.EXPECT().GetPermissions(gomock.Any(), 1, 10).Return(
			&corev1.Result[[]permissionsv1.Permission]{Items: perms, TotalCount: 2}, nil,
		)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.NoError(t, err)
		assert.Len(t, got.Items, 2)
		assert.Equal(t, int64(2), got.TotalCount)
	})

	t.Run("unauthorized — returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		authDenied(ra)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
		assertForbidden(t, err)
	})

	t.Run("reader error propagates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		pr := permmocks.NewMockReader(ctrl)
		pw := permmocks.NewMockWriter(ctrl)
		ra := authmocks.NewMockAuthorizer(ctrl)

		readerErr := errors.New("db error")
		authAllowed(ra)
		pr.EXPECT().GetPermissions(gomock.Any(), 1, 10).Return(nil, readerErr)

		svc := newTestService(pr, pw, ra)
		got, err := svc.GetPermissions(ctx, 1, 10)
		require.Error(t, err)
		assert.Nil(t, got)
	})
}
