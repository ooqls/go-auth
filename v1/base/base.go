package main

import (
	"flag"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/config"
	"github.com/ooqls/go-auth/v1/seed"
	"go.uber.org/zap"
)

var (
	appConfigPath   string
	standalone      bool
	testEnvironment bool
)

var l *zap.Logger = log.NewLogger("base-startup")

type RegisterFunc[T any] func(e *gin.Engine, server T)

func init() {
	flag.StringVar(&appConfigPath, "app-config", "", "path to app config")
	flag.BoolVar(&standalone, "standalone", false, "run in standalone mode")
	flag.BoolVar(&testEnvironment, "test-environment", false, "run in test environment")
}

func BuildBaseGinApp(
	registers []RegisterRoutesFunc,
	middlewares ...func(c *gin.Context)) (*app.App, error) {
	var appConfig *app.AppConfig
	var err error
	if standalone {
		appConfig = config.GetStandaloneAppConfig(8080)
	} else {
		appConfig, err = app.LoadConfig(appConfigPath)
		if err != nil {
			l.Fatal("failed to load app config", zap.Error(err))
		}
	}

	application := app.New("users", app.WithConfig(appConfig))

	if testEnvironment || standalone {
		application.WithTestEnvironment(app.TestEnvironment{
			Redis:         false,
			Postgres:      true,
			Elasticsearch: true,
			Valkey:        true,
		})

	}

	application.OnStartup(func(ctx *app.AppContext) error {
		cacheFactory := ctx.CacheFactory()

		factory := datav1.NewFactory(pgx.GetPGX(), cacheFactory)

		seedService := seed.NewServiceImpl(factory)
		if standalone {
			internalCtx := authorizationv1.NewInternalOperationContext(ctx)
			err = seedService.Seed(&internalCtx, defaultSeed)
			if err != nil {
				l.Error("failed to seed standalone database", zap.Error(err))
				return err
			}
		}

		e := application.Features().Gin.Engine

		// In standalone mode authentication is cookie-based: NoProxyMiddleware
		// validates the OKEY cookie and sets the X-User-Id header that
		// TrustedMiddleware then trusts. It must therefore run before
		// TrustedMiddleware, and both must be registered before any routes are
		// registered below (gin only applies middleware to routes added after
		// the e.Use call).
		if standalone {
			e.Use(authenticationv1.NoProxyMiddleware())
		}

		e.Use(authenticationv1.TrustedMiddleware(factory))
		for _, middleware := range middlewares {
			e.Use(middleware)
		}

		for _, register := range registers {
			err := register(ctx, e)
			if err != nil {
				l.Fatal("failed to build and register server", zap.Error(err))
				os.Exit(1)
			}
		}

		return nil
	})

	application.OnRunning(func(ctx *app.AppContext) error {
		cfg, ok := ctx.AuthIssuerConfig()
		if ok {
			l.Info("issuer config", zap.Float64("valid seconds", cfg.ValidityDurationSeconds))
		}
		
		return nil
	})

	return application, nil
}
