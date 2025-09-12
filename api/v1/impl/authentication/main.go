package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ooqls/go-app/app"
	"github.com/ooqls/go-auth/api/v1/gen/gen_authentication"
	"github.com/ooqls/go-auth/domain/authentication"
	"github.com/ooqls/go-auth/domain/authorization"
	usersvc "github.com/ooqls/go-auth/domain/serivce/users"
	"github.com/ooqls/go-auth/records/users"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-cache/factory"
	"github.com/ooqls/go-cache/store"
	"github.com/ooqls/go-crypto/jwt"
	"github.com/ooqls/go-crypto/keys"
	"github.com/ooqls/go-db/redis"
	"github.com/ooqls/go-db/sqlx"
	"github.com/ooqls/go-log"
	"github.com/ooqls/go-log/api/v1/gen"
)

var (
	appConfigPath string
)

func init() {
	flag.StringVar(&appConfigPath, "app-config", "", "path to app config")
}

func main() {
	flag.Parse()
	log.SetLogLevel(gen.DEBUG)
	appConfig := app.AppConfig{
		ServerConfig: app.ServerConfig{
			Port: 8080,
		},
		SQLFiles: app.SQLFilesConfig{
			Enabled:      true,
			SQLPackage:   app.SQLXPackage,
			SQLFilesDirs: []string{"./migrations/"},
		},
		Gin: app.GinConfig{
			Enabled: true,
			Port:    8080,
		},
		LoggingAPI: app.LoggingAPIConfig{
			Enabled: false,
		},
		JWT: app.JWTConfig{
			Enabled: true,
			TokenConfigurations: []jwt.TokenConfiguration{
				{
					Audience:                []string{"auth"},
					Issuer:                  app.AuthIssuer,
					ValidityDurationSeconds: 3600,
				},
				{
					Audience:                []string{"refresh"},
					Issuer:                  app.RefreshIssuer,
					ValidityDurationSeconds: 3600,
				},
			},
		},
		DocsConfig: app.DocsConfig{
			Enabled:     true,
			DocsDir:     "./docs/",
			DocsApiPath: "/api/v1/docs",
		},
	}

	authApp := app.New("authentication", app.WithConfig(&appConfig))

	authApp.WithTestEnvironment(app.TestEnvironment{
		Redis:    true,
		Postgres: true,
	})
	authApp.OnStartup(func(ctx *app.AppContext) error {
		db := sqlx.GetSQLX()
		store.Register(authentication.Challenge{})

		authCfg, ok := ctx.AuthIssuerConfig()
		if !ok {
			return fmt.Errorf("auth issuer config not found")
		}

		refreshCfg, ok := ctx.RefreshIssuerConfig()
		if !ok {
			return fmt.Errorf("refresh issuer config not found")
		}
		userR := users.NewSQLUserReader(cache.New[[]users.User](cache.NewMemCache()), db)
		userW := users.NewSQLWriter(db)

		ua := authorization.NewUserAuthorizerImpl(userR)
		chalStore := store.NewRedisStore("challenges", *redis.GetConnection(), time.Minute*15)
		authIssuer := jwt.NewJwtTokenIssuer[authentication.UserClaims](authCfg, keys.JWT())
		refreshIssuer := jwt.NewJwtTokenIssuer[authentication.UserClaims](refreshCfg, keys.JWT())

		challenger := authentication.NewChallengerV1(chalStore)
		cacheFactory := factory.NewRedisCacheFactory(*redis.GetConnection())
		authenticator := authentication.NewAuthenticatorV1(authIssuer, refreshIssuer, cacheFactory, challenger, []string{"auth"})
		userService := usersvc.NewUserServiceImpl(ua, userR, userW)
		server := NewAuthenticationServer(ctx.L(), authenticator, userService)

		e := authApp.Features().Gin.Engine
		gen_authentication.RegisterHandlers(e, server)

		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	authApp.Run(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	cancel()
}
