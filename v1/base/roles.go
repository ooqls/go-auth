package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/aggs"
	"github.com/ooqls/go-auth/v1/permissions"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/roles/rolesapi"
	"github.com/ooqls/go-auth/v1/roles/rolesapi/gen_roles"
)

func NewRolesServer(ctx *app.AppContext) (gen_roles.ServerInterface, error) {
	cacheFactory, _ := ctx.CacheFactory()
	factory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)
	rolesService := roles.NewServiceImpl(factory)
	permissionService := permissions.NewServiceImpl(factory)
	aggsService := aggs.NewServiceImpl(factory)
	return rolesapi.NewRolesServer(rolesService, aggsService, permissionService, ctx.L()), nil
}

func RegisterRolesHandlers(e *gin.Engine, server gen_roles.ServerInterface) {
	g := e.Group("api/v1/")
	gen_roles.RegisterHandlers(g, server)
}

func RegisterRolesDocsHandler(e *gin.Engine, _ gen_roles.ServerInterface) {
	g := e.Group("api/v1/")
	rolesapi.RegisterDocsHandler(g)
}

func RolesHandlers() []RegisterFunc[gen_roles.ServerInterface] {
	return []RegisterFunc[gen_roles.ServerInterface]{
		RegisterRolesHandlers,
		RegisterRolesDocsHandler,
	}
}
