package main

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	gen "github.com/ooqls/go-auth/api/v1/gen/gen_authentication"
	"github.com/ooqls/go-auth/domain/authentication"
	"github.com/ooqls/go-auth/domain/authorization"
	"github.com/ooqls/go-auth/domain/serivce/users"
	"go.uber.org/zap"
)

var _ gen.ServerInterface = &AuthenticationServerImpl{}

func NewAuthenticationServer(l *zap.Logger, authenticator authentication.Authenticator, userService users.UserService) *AuthenticationServerImpl {
	return &AuthenticationServerImpl{
		l:             l,
		Authenticator: authenticator,
		userService:   userService,
	}
}

type AuthenticationServerImpl struct {
	l             *zap.Logger
	Authenticator authentication.Authenticator
	userService   users.UserService
}

func (a *AuthenticationServerImpl) LoginChallenge(ctx *gin.Context) {
	var request gen.LoginChallengeJSONRequestBody
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	authCtx := authorization.NewInternalOperationContext(ctx)

	user, err := a.userService.GetUserByUsername(authCtx, request.Username)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	challenge, err := a.Authenticator.ChallengeRequest(authCtx, user)
	if err != nil {
		a.l.Error("failed to issue challenge", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "failed to issue challenge"})
		return
	}
	a.l.Sugar().Infof("Issued challenge for user %s, salt: %s", user.Username, user.Salt)
	serverResponse := gen.ChallengeServerResponse{
		Id:              challenge.ID,
		Base64Challenge: base64.StdEncoding.EncodeToString(challenge.Challenge),
		Base64Salt:      base64.StdEncoding.EncodeToString(challenge.User.Salt),
	}

	ctx.JSON(200, serverResponse)
}

func (a *AuthenticationServerImpl) LoginChallengeResponse(ctx *gin.Context) {
	var request gen.ChallengeClientResponse
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	challengeStr, err := base64.StdEncoding.DecodeString(request.Base64Challenge)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	okey, rkey, userID, err := a.Authenticator.ChallengeResponse(ctx, request.Id, challengeStr)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	ctx.SetCookie("OKEY", okey, 0, "/", "", true, true)
	ctx.SetCookie("RKEY", rkey, 0, "/", "", true, true)
	ctx.SetCookie("UID", userID, 0, "/", "", true, true)

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) RefreshToken(ctx *gin.Context) {
	var request gen.RefreshTokenJSONRequestBody
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := a.Authenticator.AuthenticateWithToken(ctx, request.RefreshToken)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	a.l.Sugar().Infow("User %s refreshed token successfully", request.RefreshToken)

	ctx.SetCookie("OKEY", resp.AuthToken, 0, "/", "", true, true)

	ctx.JSON(200, gin.H{})
}

func (a AuthenticationServerImpl) Register(ctx *gin.Context) {
	var req gen.RegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "bad register request"})
		return
	}
	authCtx := authorization.NewInternalOperationContext(ctx)
	user, err := a.userService.GetUserByUsername(authCtx, req.Username)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "failed to get user"})
		return
	}

	if user != nil {
		ctx.JSON(400, gin.H{"error": "user already exists"})
		return
	}

	userKey, err := base64.StdEncoding.DecodeString(req.Base64Key)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid key"})
		return
	}

	secret, err := base64.StdEncoding.DecodeString(req.EncryptedSecret)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid secret"})
		return
	}

	salt, err := a.Authenticator.ValidateRegistration(ctx, authentication.Registration{
		Username: req.Username,
		Key:      userKey,
		Email:    req.Email,
		Secret:   secret,
	})
	if err != nil {
		ctx.JSON(400, gin.H{"error": "failed to validate registration"})
		return
	}

	user, err = a.userService.CreateUser(authCtx, req.Email, string(userKey), string(salt[:]), req.Username)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "failed to create user"})
		return
	}

	authed, err := a.Authenticator.AuthenticateNewUser(authCtx, user)
	if err != nil {
		a.l.Error("failed to get a token with newly created user", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "failed to get token"})
		return
	}

	ctx.SetCookie("OKEY", authed.AuthToken, 0, "/", "", true, true)
	ctx.SetCookie("RKEY", authed.RefreshToken, 0, "/", "", true, true)
	ctx.SetCookie("UID", authed.UserId.String(), 0, "/", "", true, true)

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) AuthenticateToken(ctx *gin.Context) {
	_, err := a.Authenticator.AuthenticateWithToken(ctx, ctx.GetHeader("OKEY"))
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	ctx.JSON(200, gin.H{})
}
