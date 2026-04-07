package main

import (
	"context"
	"flag"
	"os"

	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

type Api string

const (
	ApiAuthentication Api = "authentication"
	ApiRoles          Api = "roles"
	ApiUsers          Api = "users"
	ApiPermissions    Api = "permissions"
	ApiResources      Api = "resources"
)

var api string

func init() {
	flag.StringVar(&api, "api", string(ApiAuthentication), "api to build")
}
func main() {
	flag.Parse()
	l := log.NewLogger(api)
	var app *app.App
	var err error

	switch api {
	case string(ApiAuthentication):
		app, err = BuildBaseGinApp(
			NewAuthenticationServer,
			AuthenticationHandlers(),
		)
	case string(ApiRoles):
		app, err = BuildBaseGinApp(
			NewRolesServer,
			RolesHandlers(),
		)
	case string(ApiUsers):
		app, err = BuildBaseGinApp(
			NewUsersServer,
			UserHandlers(),
		)
	case string(ApiPermissions):
		app, err = BuildBaseGinApp(
			NewPermissionsServer,
			PermissionsHandlers(),
		)
	case string(ApiResources):
		app, err = BuildBaseGinApp(
			NewResourcesServer,
			ResourceHandlers(),
		)
	}
	if err != nil {
		l.Fatal("failed to build base app", zap.Error(err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Run(ctx)
}
