package main

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/ooqls/go-auth/api/v1/gen/gen_authorization_roles"
// 	"github.com/ooqls/go-auth/domain/authorization"
// 	"github.com/ooqls/go-auth/records/roles"
// 	"go.uber.org/zap"
// )

// var _ gen_authorization_roles.ServerInterface = &RolesServer{}

// type RolesServer struct {
// 	roleAuthorizer authorization.RoleAuthorizer
// 	roleService    roles.RolesService
// 	l              zap.Logger
// }

// func NewRolesServer(l *zap.Logger, ra *authorization.RoleAuthorizer, rs *roles.RolesService) *RolesServer {
// 	return &RolesServer{l: l, roleAuthorizer: ra, roleService: rs}
// }

// func (r *RolesServer) CreateAuthRole(ctx *gin.Context) {
// 	var createReq gen_authorization_roles.CreateAuthRoleJSONRequestBody
// 	if err := ctx.ShouldBindJSON(&createReq); err != nil {
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// }

// func (r *RolesServer) DeleteAuthRole(ctx *gin.Context) {

// }

// func (r *RolesServer) GetAuthRole(ctx *gin.Context) {

// }

// func (r *RolesServer) ListAuthRoles(ctx *gin.Context) {

// }

// func (r *RolesServer) UpdateAuthRole(ctx *gin.Context) {

// }
