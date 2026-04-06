package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/config"
	"go.uber.org/zap"
)

var (
	appConfigPath   string
	standalone      bool
	testEnvironment bool
)

var l *zap.Logger = log.NewLogger("base-startup")

func init() {
	flag.StringVar(&appConfigPath, "app-config", "", "path to app config")
	flag.BoolVar(&standalone, "standalone", false, "run in standalone mode")
	flag.BoolVar(&testEnvironment, "test-environment", false, "run in test environment")
}

func BuildBaseGinApp[T any](
	newServer func(ctx *app.AppContext) (T, error),
	registerFunc func(e *gin.Engine, server T),
	middlewares ...func(c *gin.Context)) (*app.App, error) {
	var appConfig *app.AppConfig
	var err error
	if standalone {
		appConfig = config.GetStandaloneAppConfig(8082)
	} else {
		appConfig, err = app.LoadConfig(appConfigPath)
		if err != nil {
			l.Fatal("failed to load app config", zap.Error(err))
		}
	}

	application := app.New("users", app.WithConfig(appConfig))

	if testEnvironment {
		application.WithTestEnvironment(app.TestEnvironment{
			Redis:    true,
			Postgres: true,
		})
	}

	application.OnStartup(func(ctx *app.AppContext) error {

		server, err := newServer(ctx)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		cacheFactory, _ := ctx.CacheFactory()
		factory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)

		e := application.Features().Gin.Engine
		e.Use(authenticationv1.TrustedMiddleware(factory, l))
		for _, middleware := range middlewares {
			e.Use(middleware)
		}
		registerFunc(e, server)
		return nil
	})

	return application, nil
}
