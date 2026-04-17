package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/permissionbindings"
	"github.com/ooqls/go-auth/v1/permissionbindings/permissionbindingsapi"
	"github.com/ooqls/go-auth/v1/permissionbindings/permissionbindingsapi/gen_permissionbindings"
)

func NewPermissionBindingsServer(ctx *app.AppContext) (gen_permissionbindings.ServerInterface, error) {
	cacheFactory, _ := ctx.CacheFactory()
	factory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)
	authorizer := authorizationv1.NewAuthorizerImpl(factory)
	pbService := permissionbindings.NewServiceImpl(authorizer, factory)
	return permissionbindingsapi.NewServer(pbService, ctx.L()), nil
}

func RegisterPermissionBindingsHandlers(e *gin.Engine, server gen_permissionbindings.ServerInterface) {
	g := e.Group("api/v1/")
	gen_permissionbindings.RegisterHandlers(g, server)
}

func RegisterPermissionBindingsDocsHandler(e *gin.Engine, _ gen_permissionbindings.ServerInterface) {
	g := e.Group("api/v1/")
	permissionbindingsapi.RegisterDocsHandler(g)
}

func PermissionBindingsHandlers() []RegisterFunc[gen_permissionbindings.ServerInterface] {
	return []RegisterFunc[gen_permissionbindings.ServerInterface]{
		RegisterPermissionBindingsHandlers,
		RegisterPermissionBindingsDocsHandler,
	}
}
