package permissionbindingsapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/v1/permissionbindings"
	"github.com/ooqls/go-auth/v1/permissionbindings/permissionbindingsapi/gen_permissionbindings"
	"go.uber.org/zap"
)

var _ gen_permissionbindings.ServerInterface = (*Server)(nil)

type Server struct {
	pb permissionbindings.Service
	l  *zap.Logger
}

func NewServer(pb permissionbindings.Service, l *zap.Logger) *Server {
	return &Server{pb: pb, l: l}
}

func (s *Server) AssignPermission(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var req gen_permissionbindings.AssignPermissionJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	roleIds := make([]uuid.UUID, len(req.RoleIds))
	for i, id := range req.RoleIds {
		roleIds[i] = uuid.UUID(id)
	}

	permissionIds := make([]uuid.UUID, len(req.PermissionIds))
	for i, id := range req.PermissionIds {
		permissionIds[i] = uuid.UUID(id)
	}

	if err := s.pb.AssignPermission(authCtx, roleIds, permissionIds); err != nil {
		s.l.Error("failed to assign permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *Server) UnassignPermission(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var req gen_permissionbindings.UnassignPermissionJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	roleIds := make([]uuid.UUID, len(req.RoleIds))
	for i, id := range req.RoleIds {
		roleIds[i] = uuid.UUID(id)
	}

	permissionIds := make([]uuid.UUID, len(req.PermissionIds))
	for i, id := range req.PermissionIds {
		permissionIds[i] = uuid.UUID(id)
	}

	if err := s.pb.UnassignPermission(authCtx, roleIds, permissionIds); err != nil {
		s.l.Error("failed to unassign permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *Server) ListPermissionBindings(c *gin.Context, params gen_permissionbindings.ListPermissionBindingsParams) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	page := 1
	pageSize := 20
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	bindings, err := s.pb.GetPermissionsBindings(authCtx, page, pageSize)
	if err != nil {
		s.l.Error("failed to list permission bindings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	result := make([]gen_permissionbindings.PermissionBinding, len(bindings))
	for i, b := range bindings {
		result[i] = gen_permissionbindings.PermissionBinding{
			RoleId:       openapi_types.UUID(b.RoleID),
			PermissionId: openapi_types.UUID(b.PermissionID),
		}
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) GetPermissionBindingsForRole(c *gin.Context, roleId openapi_types.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	bindings, err := s.pb.GetPermissionBindingsForRole(authCtx, uuid.UUID(roleId))
	if err != nil {
		s.l.Error("failed to get permission bindings for role", zap.Error(err), zap.String("roleId", roleId.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	result := make([]gen_permissionbindings.PermissionBinding, len(bindings))
	for i, b := range bindings {
		result[i] = gen_permissionbindings.PermissionBinding{
			RoleId:       openapi_types.UUID(b.RoleID),
			PermissionId: openapi_types.UUID(b.PermissionID),
		}
	}

	c.JSON(http.StatusOK, result)
}
