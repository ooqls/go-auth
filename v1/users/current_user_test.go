package users

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	authmocks "github.com/ooqls/go-auth/internal/authorizationv1/mocks"
	"github.com/ooqls/go-auth/internal/usersv1"
	usermocks "github.com/ooqls/go-auth/internal/usersv1/mocks"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserServiceImpl_GetCurrentUser(t *testing.T) {
	t.Parallel()

	backendErr := errors.New("authorization backend unavailable")
	tests := []struct {
		name        string
		allowed     [5]bool
		failAt      int
		wrappedDeny bool
	}{
		{name: "self-read with no grants", failAt: -1},
		{name: "all supported operations allowed", allowed: [5]bool{true, true, true, true, true}, failAt: -1},
		{name: "read does not imply create", allowed: [5]bool{false, false, true, false, false}, failAt: -1},
		{name: "resource scope denied independently", allowed: [5]bool{true, true, true, true, false}, failAt: -1},
		{name: "wrapped denials become false decisions", failAt: -1, wrappedDeny: true},
		{name: "core evaluation failure", failAt: 2},
		{name: "resource evaluation failure", failAt: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			reader := usermocks.NewMockReader(ctrl)
			authorizer := authmocks.NewMockAuthorizer(ctrl)
			user := usersv1.NewUser(uuid.New(), "alice", "alice@example.com", "", "")
			authCtx := authorizationv1.NewAuthorizationContext(context.Background(), user)
			reader.EXPECT().GetUser(&authCtx, user.Id).Return(&user, nil)
			for i, target := range []struct{ kind, action string }{
				{"user", "read"}, {"role", "read"}, {"permission", "read"}, {"permission", "create"}, {"*", "read"},
			} {
				var decisionErr error
				if i == tt.failAt {
					decisionErr = backendErr
				} else if !tt.allowed[i] {
					decisionErr = authorizationv1.ErrPermissionDenied
					if tt.wrappedDeny {
						decisionErr = fmt.Errorf("authorization: %w", decisionErr)
					}
				}
				if i == 4 {
					authorizer.EXPECT().IsAuthorizedToPerformResourceAction(&authCtx, target.action, "*", "*").Return(decisionErr)
				} else {
					authorizer.EXPECT().IsAuthorizedToPerformCoreAction(&authCtx, target.action, "auth", target.kind).Return(decisionErr)
				}
				if i == tt.failAt {
					break
				}
			}

			svc := &UserServiceImpl{ua: authorizer, userR: reader}
			current, err := svc.GetCurrentUser(&authCtx)
			if tt.failAt >= 0 {
				require.ErrorIs(t, err, backendErr)
				var internal *v1.InternalError
				assert.ErrorAs(t, err, &internal)
				assert.Nil(t, current, "do not return a partially evaluated snapshot")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, current)
			assert.Equal(t, user, current.User)
			assert.Equal(t, CurrentUserCapabilities{
				UsersRead: tt.allowed[0], RolesRead: tt.allowed[1], PermissionsRead: tt.allowed[2],
				PermissionsCreate: tt.allowed[3], ResourcesRead: tt.allowed[4],
			}, current.Capabilities)
		})
	}
}

func TestUserServiceImpl_GetCurrentUserRequiresIdentity(t *testing.T) {
	t.Parallel()
	unauthenticated := authorizationv1.NewUnauthenticatedContext(context.Background())
	internal := authorizationv1.NewInternalOperationContext(context.Background())
	for _, tt := range []struct {
		name string
		ctx  *authorizationv1.Context
	}{
		{name: "missing context"},
		{name: "missing user identity", ctx: &unauthenticated},
		{name: "internal operation has no current user", ctx: &internal},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &UserServiceImpl{}
			current, err := svc.GetCurrentUser(tt.ctx)
			var forbidden *v1.ForbiddenError
			require.ErrorAs(t, err, &forbidden)
			assert.Nil(t, current)
		})
	}
}

func TestUserServiceImpl_GetCurrentUserReaderFailure(t *testing.T) {
	t.Parallel()
	readerErr := errors.New("user reader unavailable")
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "missing current user"},
		{name: "reader unavailable", err: readerErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			reader := usermocks.NewMockReader(ctrl)
			user := usersv1.NewUser(uuid.New(), "alice", "alice@example.com", "", "")
			authCtx := authorizationv1.NewAuthorizationContext(context.Background(), user)
			reader.EXPECT().GetUser(&authCtx, user.Id).Return(nil, tt.err)
			svc := &UserServiceImpl{userR: reader}
			current, err := svc.GetCurrentUser(&authCtx)
			assert.Nil(t, current)
			if tt.err != nil {
				require.ErrorIs(t, err, readerErr)
				var internal *v1.InternalError
				assert.ErrorAs(t, err, &internal)
			} else {
				var notFound *v1.NotFoundError
				assert.ErrorAs(t, err, &notFound)
			}
		})
	}
}
