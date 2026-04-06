package rolebindingsapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"github.com/ooqls/go-auth/v1/rolebindings/rolebindingsapi/gen_rolebindings"
	"go.uber.org/zap"
)

var _ gen_rolebindings.ServerInterface = (*Server)(nil)

type Server struct {
	rb rolebindings.Service
	l  *zap.Logger
}

func NewServer(rb rolebindings.Service, l *zap.Logger) *Server {
	return &Server{rb: rb, l: l}
}

func (s *Server) AssignRole(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var req gen_rolebindings.AssignRoleJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	for _, userID := range req.UserIDs {
		for _, roleID := range req.RoleIDs {
			if err := s.rb.AssignRoleToUser(authCtx, uuid.UUID(userID), uuid.UUID(roleID)); err != nil {
				s.l.Error("failed to assign role to user", zap.Error(err), zap.String("userID", userID.String()), zap.String("roleID", roleID.String()))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *Server) UnassignRole(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var req gen_rolebindings.UnassignRoleJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	for _, userID := range req.UserIDs {
		for _, roleID := range req.RoleIDs {
			if err := s.rb.UnassignRoleFromUser(authCtx, uuid.UUID(userID), uuid.UUID(roleID)); err != nil {
				s.l.Error("failed to unassign role from user", zap.Error(err), zap.String("userID", userID.String()), zap.String("roleID", roleID.String()))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s *Server) GetAllRoleBindings(c *gin.Context, params gen_rolebindings.GetAllRoleBindingsParams) {
	_, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	// TODO: Implement list all role bindings with pagination
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (s *Server) GetRolebindingsForUser(c *gin.Context, userID openapi_types.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	bindings, err := s.rb.GetRoleBindingsForUser(authCtx, uuid.UUID(userID))
	if err != nil {
		s.l.Error("failed to get role bindings for user", zap.Error(err), zap.String("userID", userID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, bindings)
}
