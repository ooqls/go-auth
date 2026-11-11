package users

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	authmocks "github.com/ooqls/go-auth/internal/authorizationv1/mocks"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1"
	usermocks "github.com/ooqls/go-auth/internal/usersv1/mocks"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestUserServiceImpl_GetUserStillRequiresUserRead(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	authorizer := authmocks.NewMockAuthorizer(ctrl)
	reader := usermocks.NewMockReader(ctrl)
	user := usersv1.NewUser(uuid.New(), "alice", "alice@example.com", "", "")
	authCtx := authorizationv1.NewAuthorizationContext(context.Background(), user)
	authorizer.EXPECT().IsAuthorizedToPerformCoreAction(&authCtx, authorizationv1.ReadAction, "auth", "user").
		Return(authorizationv1.ErrPermissionDenied)
	svc := &UserServiceImpl{ua: authorizer, userR: reader}
	got, err := svc.GetUser(&authCtx, uuid.New())
	require.ErrorIs(t, err, authorizationv1.ErrPermissionDenied)
	assert.Nil(t, got)
}

func TestUserServiceImpl_GetUsers(t *testing.T) {
	t.Parallel()

	authorizerErr := errors.New("authorization backend unavailable")
	readerErr := errors.New("user reader unavailable")
	tests := []struct {
		name          string
		authErr       error
		readErr       error
		wantErr       error
		wantForbidden bool
		wantInternal  bool
	}{
		{name: "authorized"},
		{
			name:          "permission denied without reading users",
			authErr:       authorizationv1.ErrPermissionDenied,
			wantErr:       authorizationv1.ErrPermissionDenied,
			wantForbidden: true,
		},
		{
			name:          "wrapped permission denial without reading users",
			authErr:       fmt.Errorf("authorization: %w", authorizationv1.ErrPermissionDenied),
			wantErr:       authorizationv1.ErrPermissionDenied,
			wantForbidden: true,
		},
		{
			name:         "authorization failure without reading users",
			authErr:      authorizerErr,
			wantErr:      authorizerErr,
			wantInternal: true,
		},
		{
			name:    "reader failure after authorization",
			readErr: readerErr,
			wantErr: readerErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			authorizer := authmocks.NewMockAuthorizer(ctrl)
			reader := usermocks.NewMockReader(ctrl)
			user := usersv1.NewUser(uuid.New(), "alice", "alice@example.com", "", "")
			authCtx := authorizationv1.NewAuthorizationContext(context.Background(), user)
			want := &corev1.Result[[]User]{Items: []User{user}, TotalCount: 1}

			authCall := authorizer.EXPECT().
				IsAuthorizedToPerformCoreAction(&authCtx, authorizationv1.ReadAction, "auth", "user").
				Return(tt.authErr)
			if tt.authErr == nil {
				result := want
				if tt.readErr != nil {
					result = nil
				}
				reader.EXPECT().GetUsers(&authCtx, int32(20), int32(10)).
					After(authCall).Return(result, tt.readErr)
			} else {
				reader.EXPECT().GetUsers(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			}

			svc := &UserServiceImpl{l: zap.NewNop(), ua: authorizer, userR: reader}
			got, err := svc.GetUsers(&authCtx, 2, 10)
			if tt.wantErr == nil {
				require.NoError(t, err)
				assert.Equal(t, want, got)
				return
			}

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got)
			if tt.wantForbidden {
				var forbidden *v1.ForbiddenError
				assert.ErrorAs(t, err, &forbidden)
			}
			if tt.wantInternal {
				var internal *v1.InternalError
				assert.ErrorAs(t, err, &internal)
			}
		})
	}
}
