package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/getset/log/api/v1/gen"
	"go.uber.org/zap"
)

type Api string

const (
	ApiAuthentication Api = "authentication"
	ApiRoles          Api = "roles"
	ApiUsers          Api = "users"
	ApiPermissions    Api = "permissions"
	ApiResources      Api = "resources"
	ApiSeed           Api = "seed"
)

var AllApis = []string{string(ApiAuthentication), string(ApiRoles), string(ApiUsers), string(ApiPermissions), string(ApiResources)}

var apisFlag string

func init() {
	flag.StringVar(&apisFlag, "api", string(ApiAuthentication), "api to build")
}
func main() {
	flag.Parse()

	log.SetLogLevel(gen.DEBUG)
	l := log.NewLogger(apisFlag)

	var apis []string
	if apisFlag == "all" {
		apis = AllApis
	} else {
		apis = strings.Split(apisFlag, ",")
	}
	l.Info("starting app with apis", zap.String("apis", apisFlag))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var baseApp *app.App
	var err error

	registers := []RegisterRoutesFunc{}

	for _, api := range apis {
		switch api {
		case string(ApiAuthentication):
			registers = append(registers, func(ctx *app.AppContext, e *gin.Engine) error {
				srv, err := NewAuthenticationServer(ctx)
				if err != nil {
					return err
				}

				for _, reg := range AuthenticationHandlers() {
					reg(e, srv)
				}

				return nil
			})
		case string(ApiRoles):
			registers = append(registers, func(ctx *app.AppContext, e *gin.Engine) error {
				srv, err := NewRolesServer(ctx)
				if err != nil {
					return err
				}
				for _, reg := range RolesHandlers() {
					reg(e, srv)
				}
				return nil
			})
		case string(ApiUsers):
			registers = append(registers, func(ctx *app.AppContext, e *gin.Engine) error {
				srv, err := NewUsersServer(ctx)
				if err != nil {
					return err
				}
				for _, reg := range UserHandlers() {
					reg(e, srv)
				}
				return nil
			})
		case string(ApiPermissions):
			registers = append(registers, func(ctx *app.AppContext, e *gin.Engine) error {
				srv, err := NewPermissionsServer(ctx)
				if err != nil {
					return err
				}
				for _, reg := range PermissionsHandlers() {
					reg(e, srv)
				}
				return nil
			})
		case string(ApiResources):
			registers = append(registers, func(ctx *app.AppContext, e *gin.Engine) error {
				srv, err := NewResourcesServer(ctx)
				if err != nil {
					return err
				}
				for _, reg := range ResourceHandlers() {
					reg(e, srv)
				}
				return nil
			})
		case string(ApiSeed):
			baseApp = Seed(cancel)
			err := baseApp.Run(ctx)
			if err != nil {
				l.Fatal("failed to run seed app", zap.Error(err))
				os.Exit(1)
			}
			return
		default:
			l.Fatal("invalid api specified", zap.String("api", api))
			os.Exit(1)
		}
		baseApp, err = BuildBaseGinApp(registers)
		if err != nil {
			l.Fatal("failed to build base app", zap.Error(err))
			os.Exit(1)
		}
	}

	baseApp.Run(ctx)
}
