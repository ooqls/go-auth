package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/resources/resourcesapi"
	"github.com/ooqls/go-auth/v1/resources/resourcesapi/gen_resources"
)

func NewResourcesServer(ctx *app.AppContext) (gen_resources.ServerInterface, error) {
	factory := datav1.NewFactory(*pgx.GetPGX(), nil)
	return resourcesapi.NewResourcesServer(factory, ctx.L()), nil
}

func RegisterResourcesHandlers(e *gin.Engine, server gen_resources.ServerInterface) {
	gen_resources.RegisterHandlers(e, server)
}
