package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/valkey"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/authentication"
	authenticationapi "github.com/ooqls/go-auth/v1/authentication/api"
	"github.com/ooqls/go-auth/v1/authentication/api/gen_authentication"
	"github.com/ooqls/go-auth/v1/users"
	"go.uber.org/zap"
)

func NewAuthenticationServer(ctx *app.AppContext) (gen_authentication.ServerInterface, error) {
	refreshCfg, ok := ctx.RefreshIssuerConfig()
	if !ok {
		return nil, fmt.Errorf("refresh issuer config not found")
	}

	authCfg, ok := ctx.AuthIssuerConfig()
	if !ok {
		return nil, fmt.Errorf("auth issuer config not found")
	}
	ctx.L().Info("initializing token issuers",
		zap.Float64("refresh_expiry_seconds", refreshCfg.ValidityDurationSeconds))
	ctx.L().Info("initializing token issuers",
		zap.Float64("auth_expiry_seconds", authCfg.ValidityDurationSeconds))
	refreshIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](refreshCfg, keys.JWT())
	authenticationv1.InitRefreshIssuer(refreshIssuer)

	authenticationIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](authCfg, keys.JWT())
	authenticationv1.InitAuthenticationIssuer(authenticationIssuer)
	cacheFactory := ctx.CacheFactory()
	userIssuer := authenticationv1.NewUserTokenIssuer(authenticationIssuer, refreshIssuer)
	challenger := authenticationv1.NewChallengerV1(cacheFactory.NewStore("challenges", 15*time.Minute))
	authenticationv1.InitAuthenticator(userIssuer, ctx.CacheFactory(), challenger, []string{"auth"})
	authenticator := authenticationv1.GetAuthenticatorV1()
	dataFactory := datav1.NewFactory(pgx.GetPGX(), cacheFactory)
	userService := users.NewServiceImpl(dataFactory)
	authService := authentication.NewServiceImpl(authenticator, userService)

	c := valkey.GetConnection(ctx)
	resp := c.Do(ctx, c.B().Ping().Build())
	if resp.Error() != nil {
		l.Error("failed to ping valkey", zap.Error(resp.Error()))
		return nil, resp.Error()
	}
	l.Info("valkey ping", zap.String("status", resp.String()))
	return authenticationapi.NewAuthenticationServer(l, authService, false), nil
}

func RegisterV1AuthenticationHandlers(e *gin.Engine, server gen_authentication.ServerInterface) {
	g := e.Group("api/v1/")
	gen_authentication.RegisterHandlers(g, server)
}

func RegisterV1AuthenticationDocsHandler(e *gin.Engine, _ gen_authentication.ServerInterface) {
	g := e.Group("api/")
	authenticationapi.RegisterDocsHandler(g)
}

func AuthenticationHandlers() []RegisterFunc[gen_authentication.ServerInterface] {
	return []RegisterFunc[gen_authentication.ServerInterface]{
		RegisterV1AuthenticationHandlers,
		RegisterV1AuthenticationDocsHandler,
	}
}
