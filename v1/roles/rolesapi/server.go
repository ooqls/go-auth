package rolesapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/aggs"
	"github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/permissionbindings"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/roles/rolesapi/gen_roles"

	"go.uber.org/zap"
)

var _ gen_roles.ServerInterface = &RolesServer{}

type RolesServer struct {
	service           roles.Service
	aggService        aggs.Service
	permissionService permissions.Service
	permissionBinding permissionbindings.Service
	l                 *zap.Logger
}

func NewRolesServer(service roles.Service, aggsService aggs.Service, permissionService permissions.Service, l *zap.Logger) *RolesServer {
	return &RolesServer{
		service:           service,
		aggService:        aggsService,
		permissionService: permissionService,
		l:                 l,
	}
}

func (r *RolesServer) CreateAuthRole(ctx *gin.Context) {
	var createReq gen_roles.CreateAuthRoleJSONRequestBody
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}
	roleId, err := r.service.CreateRole(authCtx, roles.CreateRoleParams{
		Name:        createReq.Name,
		Description: createReq.Description,
		Hierarchy:   int32(createReq.Hierarchy),
	})

	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Role created successfully", "id": roleId})
}

func (r *RolesServer) DeleteAuthRole(ctx *gin.Context) {
	var deleteReq gen_roles.DeleteAuthRoleJSONRequestBody
	if err := ctx.ShouldBindJSON(&deleteReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	err := r.service.DeleteRole(authCtx, deleteReq.Id)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusNoContent, gin.H{"message": "Role deleted successfully"})
}

func (r *RolesServer) GetAuthRole(ctx *gin.Context, id uuid.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	role, err := r.service.GetRole(authCtx, id)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	if role == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	roleAgg := gen.Role{
		CreatedAt:   role.CreatedAt,
		Description: role.Description,
		Hierarchy:   int(role.Hierarchy),
		Id:          role.Id,
		Name:        role.Name,
		UpdatedAt:   role.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gen_roles.GetAuthRoleResponse{
		JSON200: &gen.Role{
			CreatedAt:   roleAgg.CreatedAt,
			Description: roleAgg.Description,
			Hierarchy:   roleAgg.Hierarchy,
			Id:          roleAgg.Id,
			Name:        roleAgg.Name,
			UpdatedAt:   roleAgg.UpdatedAt,
		},
	})
}

func (r *RolesServer) GetAuthRoles(ctx *gin.Context, params gen_roles.GetAuthRolesParams) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	page := int32(0)
	pageSize := int32(100)
	if params.Page != nil {
		page = int32(*params.Page)
	}
	if params.PageSize != nil {
		pageSize = int32(*params.PageSize)
	}

	roleList, err := r.service.ListRoles(authCtx, pageSize, page*pageSize)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	genRoles := []gen.Role{}
	for _, role := range roleList {
		genRoles = append(genRoles, gen.Role{
			CreatedAt:   role.CreatedAt,
			Description: role.Description,
			Hierarchy:   int(role.Hierarchy),
			Id:          role.Id,
			Name:        role.Name,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, gen_roles.GetAuthRolesResponse{
		JSON200: &genRoles,
	})
}

func (r *RolesServer) UpdateAuthRole(ctx *gin.Context) {
	var updateReq gen_roles.UpdateAuthRoleJSONRequestBody
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	id, err := uuid.Parse(updateReq.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := r.service.GetRole(authCtx, id)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	if role == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	updateParams := roles.UpdateRoleParams{
		ID: id,
	}

	for _, ev := range updateReq.Events {
		switch ev.Event {
		case "update_name":
			if name, ok := ev.Data["name"].(string); ok {
				updateParams.Name = &name
			}
		case "update_description":
			if desc, ok := ev.Data["description"].(string); ok {
				updateParams.Description = &desc
			}
		case "update_hierarchy":
			if h, ok := ev.Data["hierarchy"].(float64); ok {
				hierarchy := int32(h)
				updateParams.Hierarchy = &hierarchy
			}
		}
	}

	err = r.service.UpdateRole(authCtx, updateParams)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusNoContent, gin.H{"message": "Role updated successfully"})
}

func (r *RolesServer) GetAuthPermissions(ctx *gin.Context, params gen_roles.GetAuthPermissionsParams) {
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

	perms, err := r.permissionService.GetPermissions(authCtx, page, pageSize)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	genPerms := make([]gen.Permission, 0, len(perms))
	for _, p := range perms {
		genPerms = append(genPerms, gen.Permission{
			Id:            p.Id,
			ResourceName:  p.Name,
			ResourceGroup: p.Group,
			ResourceKind:  p.Kind,
			Actions:       strings.Split(p.Actions, ","),
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		})
	}

	ctx.JSON(http.StatusOK, genPerms)
}

func (r *RolesServer) CreateAuthPermission(ctx *gin.Context) {
	var createReq gen_roles.CreateAuthPermissionJSONRequestBody
	if err := ctx.ShouldBindJSON(&createReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	if createReq.Permissions == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "permissions are required"})
		return
	}

	for _, p := range *createReq.Permissions {
		if err := r.permissionService.AddPermission(authCtx, p.ResourceName, p.ResourceGroup, p.ResourceKind, p.Actions); err != nil {
			v1.GinHandleError(ctx, err)
			return
		}
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Permissions created successfully"})
}

func (r *RolesServer) AssignRolePermissions(ctx *gin.Context) {
	var addReq gen_roles.AssignRolePermissionsJSONRequestBody
	if err := ctx.ShouldBindJSON(&addReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	roleIds := *addReq.RoleIds
	permissionIds := *addReq.PermissionIds
	err := r.permissionBinding.AssignPermission(authCtx, roleIds, permissionIds)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Permissions added successfully"})
}

func (r *RolesServer) UnassignRolePermissions(ctx *gin.Context) {
	var unassignReq gen_roles.UnassignPermissionRequest
	if err := ctx.ShouldBindJSON(&unassignReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}
	roleIds := *unassignReq.RoleIds
	permissionIds := *unassignReq.PermissionIds
	err := r.permissionBinding.UnassignPermission(authCtx, roleIds, permissionIds)
	if err != nil {
		v1.GinHandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Permissions deleted successfully"})
}
