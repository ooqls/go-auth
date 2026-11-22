package resourcesapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/resources"
	"github.com/ooqls/go-auth/v1/resources/api/gen_resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type resourceReadFailure struct {
	resources.Service
	err       error
	operation string
}

func (s *resourceReadFailure) GetResources(*authorizationv1.Context, *string, *string, int32, int32) (*corev1.Result[[]resources.Resourcev1], error) {
	s.operation = "list"
	return nil, s.err
}

func (s *resourceReadFailure) SearchResources(*authorizationv1.Context, *string, *string, *string, *string) (*corev1.Result[[]resources.Resourcev1], error) {
	s.operation = "search"
	return nil, s.err
}

func TestResourcesServer_ReadErrors(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"list", "search"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			for _, tt := range []struct {
				name   string
				err    error
				status int
			}{
				{
					name:   "permission denied",
					err:    v1.ErrPermissionDenied(errors.New("permission denied"), nil),
					status: http.StatusForbidden,
				},
				{
					name:   "wrapped permission denied",
					err:    fmt.Errorf("read resources: %w", v1.ErrPermissionDenied(errors.New("permission denied"), nil)),
					status: http.StatusForbidden,
				},
				{
					name:   "dependency failure",
					err:    v1.ErrInternal(errors.New("dependency unavailable"), nil),
					status: http.StatusInternalServerError,
				},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					svc := &resourceReadFailure{err: tt.err}
					server := &ResourcesServer{service: svc, l: zap.NewNop()}
					recorder := httptest.NewRecorder()
					ctx, _ := gin.CreateTestContext(recorder)
					ctx.Set("authorization_context", authorizationv1.NewAuthorizationContext(context.Background(), usersv1.User{}))
					if operation == "search" {
						ctx.Request = httptest.NewRequest(http.MethodPost, "/resources/search", strings.NewReader("{}"))
						ctx.Request.Header.Set("Content-Type", "application/json")
						server.PostResourceSearch(ctx)
					} else {
						ctx.Request = httptest.NewRequest(http.MethodGet, "/resources", nil)
						server.ListResources(ctx, gen_resources.ListResourcesParams{})
					}
					assert.Equal(t, operation, svc.operation)
					require.Equal(t, tt.status, recorder.Code)
					assert.Contains(t, recorder.Body.String(), `"error":`)
					assert.NotContains(t, recorder.Body.String(), `"items":`)
				})
			}
		})
	}
}
