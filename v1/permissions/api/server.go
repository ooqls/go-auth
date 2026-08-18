package permissionsapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/permissions/api/gen_permissions"
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

	ctx.JSON(http.StatusOK, gen_permissions.PermissionList{
		Items:      toGenPermissionList(perms.Items),
		TotalCount: int(perms.TotalCount),
	})
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

	if err := p.service.AddPermission(authCtx, req.Permission.Permission); err != nil {
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

func (p *PermissionsServer) SearchPermissions(ctx *gin.Context) {
	var req gen_permissions.SearchPermissionsJSONRequestBody
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	results, err := p.service.SearchPermissions(authCtx, req.Permission)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (p *PermissionsServer) ListMyPermissions(ctx *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	perms, err := p.service.ListPermissionsForUser(authCtx)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}
	permArr := []string{}
	for _, p := range perms {
		permArr = append(permArr, p.Permission)
	}

	ctx.JSON(http.StatusOK, permArr)
}
