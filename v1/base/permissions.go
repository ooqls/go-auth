package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/permissions/permissionsapi"
	"github.com/ooqls/go-auth/v1/permissions/permissionsapi/gen_permissions"
)

func NewPermissionsServer(ctx *app.AppContext) (gen_permissions.ServerInterface, error) {
	cacheFactory, _ := ctx.CacheFactory()
	factory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)
	permissionService := permissions.NewServiceImpl(factory)
	return permissionsapi.NewPermissionsServer(permissionService, ctx.L()), nil
}

func RegisterPermissionsHandlers(e *gin.Engine, server gen_permissions.ServerInterface) {
	g := e.Group("api/v1/")
	gen_permissions.RegisterHandlers(g, server)
}

func RegisterPermissionsDocsHandler(e *gin.Engine, _ gen_permissions.ServerInterface) {
	g := e.Group("api/v1/")
	permissionsapi.RegisterDocsHandler(g)
}

func PermissionsHandlers() []RegisterFunc[gen_permissions.ServerInterface] {
	return []RegisterFunc[gen_permissions.ServerInterface]{
		RegisterPermissionsHandlers,
		RegisterPermissionsDocsHandler,
	}
}
