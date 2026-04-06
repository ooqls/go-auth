package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/users"
	"github.com/ooqls/go-auth/v1/users/usersapi"
	"github.com/ooqls/go-auth/v1/users/usersapi/gen_users"
)

func NewUsersServer(ctx *app.AppContext) (gen_users.ServerInterface, error) {
	emailCli, _ := ctx.EmailClient()
	factory := datav1.NewFactory(*pgx.GetPGX(), nil)
	service := users.NewServiceImpl(factory)
	return usersapi.NewUsersServer(service, emailCli, ctx.L()), nil
}

func RegisterUsersHandlers(e *gin.Engine, server gen_users.ServerInterface) {
	gen_users.RegisterHandlers(e, server)
}
