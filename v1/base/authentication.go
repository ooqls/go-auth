package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/app/app"
	"github.com/ooqls/getset/crypto/jwt"
	"github.com/ooqls/getset/crypto/keys"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/authentication"
	"github.com/ooqls/go-auth/v1/authentication/authenticationapi"
	"github.com/ooqls/go-auth/v1/authentication/authenticationapi/gen_authentication"
	"github.com/ooqls/go-auth/v1/users"
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

	refreshIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](refreshCfg, keys.JWT())
	authenticationv1.InitRefreshIssuer(refreshIssuer)

	authenticationIssuer := jwt.NewJwtTokenIssuer[claims.UserClaims](authCfg, keys.JWT())
	authenticationv1.InitAuthenticationIssuer(authenticationIssuer)
	cacheFactory, ok := ctx.CacheFactory()
	if !ok {
		return nil, fmt.Errorf("cache factory not found")
	}
	userIssuer := authenticationv1.NewUserTokenIssuer(authenticationIssuer, refreshIssuer)
	challenger := authenticationv1.NewChallengerV1(cacheFactory.NewStore("challenges", 15*time.Minute))
	authenticator := authenticationv1.NewAuthenticatorV1(
		userIssuer,
		cacheFactory,
		challenger,
		[]string{"auth"},
	)

	dataFactory := datav1.NewFactory(*pgx.GetPGX(), cacheFactory)
	userService := users.NewServiceImpl(dataFactory)
	authService := authentication.NewServiceImpl(authenticator, userService)
	return authenticationapi.NewAuthenticationServer(l, authService, false), nil
}

func RegisterAuthenticationHandlers(e *gin.Engine, server gen_authentication.ServerInterface) {
	gen_authentication.RegisterHandlers(e, server)
}
