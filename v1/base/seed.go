package main

import (
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/schema"
	"github.com/ooqls/go-auth/v1/config"
	"github.com/ooqls/go-auth/v1/seed"
	"go.uber.org/zap"
)

var defaultSeed = seed.Seed{
	Users: []seed.UserSeed{
		{
			Username: "admin",
			Email:    "admin@admin.com",
			Password: "admin",
			Roles: []seed.RoleSeed{
				{
					Name:        "admin",
					Description: "Full access to all resources",
					Hierarchy:   0,
					Permissions: []string{
						"resource:*:*:delete",
						"resource:*:*:update",
						"resource:*:*:create",
						"resource:*:*:read",
						"core:auth:user:create",
						"core:auth:user:read",
						"core:auth:user:update",
						"core:auth:user:delete",
						"core:auth:user:assign",
						"core:auth:user:unassign",
						"core:auth:user:update-key",
						"core:auth:role:create",
						"core:auth:role:read",
						"core:auth:role:update",
						"core:auth:role:delete",
						"core:auth:role:assign",
						"core:auth:role:unassign",
						"core:auth:role_binding:create",
						"core:auth:role_binding:read",
						"core:auth:role_binding:update",
						"core:auth:role_binding:delete",
						"core:auth:role_binding:assign",
						"core:auth:role_binding:unassign",
						"core:auth:resource:create",
						"core:auth:resource:read",
						"core:auth:resource:update",
						"core:auth:resource:delete",
						"core:auth:permission:create",
						"core:auth:permission:read",
						"core:auth:permission:update",
						"core:auth:permission:delete",
						"core:auth:permission:assign",
						"core:auth:permission:unassign",
						"core:auth:login_challenge:create",
						"core:auth:login_challenge:read",
						"core:auth:login_challenge:update",
						"core:auth:login_challenge:delete",
						"core:auth:challenge_attempt:create",
						"core:auth:challenge_attempt:read",
						"core:auth:challenge_attempt:update",
						"core:auth:challenge_attempt:delete",
						"core:auth:audit:create",
						"core:auth:audit:read",
						"core:auth:audit:update",
						"core:auth:audit:delete",
					},
				},
			},
		},
	},
}

func Seed(done func()) *app.App {
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
	features := app.WithConfig(appConfig)
	features.SQL = app.PGX(
		app.WithCreateTableStatements(schema.GetSchemaMigrations()),
		app.WithCreateIndexStatements(schema.GetViewMigrations()),
	)
	features.HTTP.Enabled = false
	features.Grpc.Enabled = false
	features.Docs.Enabled = false
	features.Gin.Enabled = false
	features.JWT.Enabled = false
	features.Cache.Enabled = false
	features.LoggingAPI.Enabled = false

	seedApp := app.New("seed", features)
	seedApp.OnRunning(func(ctx *app.AppContext) error {
		defer done()
		data := datav1.NewFactory(pgx.GetPGX(), ctx.CacheFactory())
		svc := seed.NewServiceImpl(data)
		authCtx := authorizationv1.NewInternalOperationContext(ctx)

		var seedObj seed.Seed
		if standalone {
			seedObj = defaultSeed
		} else {
			loaded, err := seed.LoadSeedFromFile("/seed/seed.yaml")
			if err != nil {
				ctx.L().Error("failed to load seed file", zap.Error(err))
				return err
			}
			seedObj = *loaded
		}

		err := svc.Seed(&authCtx, seedObj)
		if err != nil {
			ctx.L().Error("failed to seed database", zap.Error(err))
			return err
		}

		ctx.L().Info("successfully seeded database")
		return nil
	})
	seedApp.OnStopped(func(ctx *app.AppContext) error {
		ctx.L().Info("seed app stopping")
		return nil
	})
	return seedApp

}
