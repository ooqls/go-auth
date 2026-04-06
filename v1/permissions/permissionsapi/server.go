package permissionsapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/permissions/permissionsapi/gen_permissions"
	"go.uber.org/zap"
)

var _ gen_permissions.ServerInterface = &PermissionsServer{}

type PermissionsServer struct {
	service permissions.Service
	l       *zap.Logger
}

func NewPermissionsServer(service permissions.Service, l *zap.Logger) *PermissionsServer {
	return &PermissionsServer{
		service: service,
		l:       l,
	}
}

func (p *PermissionsServer) ListPermissions(ctx *gin.Context, params gen_permissions.ListPermissionsParams) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	page := 0
	pageSize := 100
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	perms, err := p.service.GetPermissions(authCtx, page, pageSize)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	genPerms := make([]gen.Permission, 0, len(perms))
	for _, perm := range perms {
		actions := strings.Split(perm.Actions, ",")
		genPerms = append(genPerms, gen.Permission{
			Id:            perm.Id,
			ResourceName:  perm.Name,
			ResourceGroup: perm.Group,
			ResourceKind:  perm.Kind,
			Actions:       actions,
			CreatedAt:     perm.CreatedAt,
			UpdatedAt:     perm.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, genPerms)
}

func (p *PermissionsServer) CreatePermission(ctx *gin.Context) {
	var req gen_permissions.CreatePermissionJSONRequestBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	perm := req.Permission
	if err := p.service.AddPermission(authCtx, perm.ResourceName, perm.ResourceGroup, perm.ResourceKind, perm.Actions); err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Permission created successfully"})
}

func (p *PermissionsServer) DeletePermission(ctx *gin.Context) {
	// The delete endpoint only provides an ID, but the service requires (name, group, kind).
	// A GetPermissionById reader method is needed to bridge this gap.
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (p *PermissionsServer) UpdatePermission(ctx *gin.Context) {
	// The service does not expose an UpdatePermission method.
	// Extend permissions.Service with UpdatePermission to implement this endpoint.
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
