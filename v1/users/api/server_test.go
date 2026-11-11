package usersapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/users"
	"github.com/ooqls/go-auth/v1/users/api/gen_users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type userServiceStub struct {
	users.Service
	getUsers       func(*authorizationv1.Context, int, int) (*corev1.Result[[]users.User], error)
	getCurrentUser func(*authorizationv1.Context) (*users.CurrentUser, error)
}

func (s *userServiceStub) GetUsers(ctx *authorizationv1.Context, page, pageSize int) (*corev1.Result[[]users.User], error) {
	return s.getUsers(ctx, page, pageSize)
}

func (s *userServiceStub) GetCurrentUser(ctx *authorizationv1.Context) (*users.CurrentUser, error) {
	return s.getCurrentUser(ctx)
}

func TestUsersServer_GetCurrentUser(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		withAuth   bool
		serviceErr error
		wantStatus int
	}{
		{name: "missing identity", wantStatus: http.StatusForbidden},
		{name: "authenticated self with evaluated capabilities", withAuth: true, wantStatus: http.StatusOK},
		{name: "identity rejected", withAuth: true, serviceErr: v1.ErrPermissionDenied(errors.New("authenticated user required"), nil), wantStatus: http.StatusForbidden},
		{name: "current user missing", withAuth: true, serviceErr: v1.ErrNotFound(errors.New("current user not found"), nil), wantStatus: http.StatusNotFound},
		{name: "evaluation unavailable", withAuth: true, serviceErr: v1.ErrInternal(errors.New("authorization unavailable"), nil), wantStatus: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			user := sampleUser()
			called := false
			svc := &userServiceStub{getCurrentUser: func(ctx *authorizationv1.Context) (*users.CurrentUser, error) {
				called = true
				assert.Equal(t, user.Id, ctx.GetAuthedUser().Id)
				if tt.serviceErr != nil {
					return nil, tt.serviceErr
				}
				return &users.CurrentUser{User: user, Capabilities: users.CurrentUserCapabilities{PermissionsRead: true}}, nil
			}}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/user", nil)
			if tt.withAuth {
				ctx.Set("authorization_context", authorizationv1.NewAuthorizationContext(context.Background(), user))
			}
			NewUsersServer(svc, nil, zap.NewNop()).GetCurrentUser(ctx)
			assert.Equal(t, tt.withAuth, called)
			require.Equal(t, tt.wantStatus, recorder.Code)
			assert.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
			if tt.wantStatus != http.StatusOK {
				assert.NotContains(t, recorder.Body.String(), "capabilities")
				return
			}
			var body gen_users.CurrentUserResponse
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
			assert.Equal(t, user.Id, *body.User.Id)
			assert.Equal(t, gen_users.CurrentUserCapabilities{PermissionsRead: true}, body.Capabilities)
			assert.Contains(t, recorder.Body.String(), `"permissions_create":false`)
			assert.NotContains(t, recorder.Body.String(), "super-secret")
			assert.NotContains(t, recorder.Body.String(), `"key"`)
			assert.NotContains(t, recorder.Body.String(), `"salt"`)
		})
	}
}

func TestUsersServer_GetUsers(t *testing.T) {
	t.Parallel()

	page, pageSize := 2, 5
	tests := []struct {
		name         string
		withAuth     bool
		params       gen_users.GetUsersParams
		serviceErr   error
		wantStatus   int
		wantPage     int
		wantPageSize int
	}{
		{
			name:       "missing authorization context",
			wantStatus: http.StatusForbidden,
		},
		{
			name:         "authorized with default pagination",
			withAuth:     true,
			wantStatus:   http.StatusOK,
			wantPageSize: 10,
		},
		{
			name:         "authorized with explicit pagination",
			withAuth:     true,
			params:       gen_users.GetUsersParams{Page: &page, PageSize: &pageSize},
			wantStatus:   http.StatusOK,
			wantPage:     page,
			wantPageSize: pageSize,
		},
		{
			name:         "permission denied",
			withAuth:     true,
			serviceErr:   v1.ErrPermissionDenied(authorizationv1.ErrPermissionDenied, nil),
			wantStatus:   http.StatusForbidden,
			wantPageSize: 10,
		},
		{
			name:         "wrapped permission denial",
			withAuth:     true,
			serviceErr:   fmt.Errorf("list users: %w", v1.ErrPermissionDenied(authorizationv1.ErrPermissionDenied, nil)),
			wantStatus:   http.StatusForbidden,
			wantPageSize: 10,
		},
		{
			name:         "authorization backend failure",
			withAuth:     true,
			serviceErr:   v1.ErrInternal(errors.New("authorization backend unavailable"), nil),
			wantStatus:   http.StatusInternalServerError,
			wantPageSize: 10,
		},
		{
			name:         "reader failure",
			withAuth:     true,
			serviceErr:   errors.New("user reader unavailable"),
			wantStatus:   http.StatusInternalServerError,
			wantPageSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user := sampleUser()
			called := false
			svc := &userServiceStub{
				getUsers: func(ctx *authorizationv1.Context, page, pageSize int) (*corev1.Result[[]users.User], error) {
					called = true
					assert.Equal(t, user.Id, ctx.GetAuthedUser().Id)
					assert.Equal(t, tt.wantPage, page)
					assert.Equal(t, tt.wantPageSize, pageSize)
					if tt.serviceErr != nil {
						return nil, tt.serviceErr
					}
					return &corev1.Result[[]users.User]{Items: []users.User{user}, TotalCount: 1}, nil
				},
			}
			server := NewUsersServer(svc, nil, zap.NewNop())
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
			if tt.withAuth {
				ctx.Set("authorization_context", authorizationv1.NewAuthorizationContext(context.Background(), user))
			}

			server.GetUsers(ctx, tt.params)

			assert.Equal(t, tt.withAuth, called)
			require.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantStatus != http.StatusOK {
				var body map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
				assert.Contains(t, body, "error")
				assert.NotContains(t, body, "items")
				return
			}

			var body gen_users.UserList
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
			assert.Equal(t, 1, body.TotalCount)
			require.Len(t, body.Items, 1)
			assert.Equal(t, user.Username, body.Items[0].Username)
			assert.Equal(t, user.Email, body.Items[0].Email)
		})
	}
}
