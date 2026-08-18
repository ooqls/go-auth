package authenticationapi

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/v1/authentication"
	"github.com/ooqls/go-auth/v1/authentication/api/gen_authentication"
	"go.uber.org/zap"
)

var _ gen_authentication.ServerInterface = &AuthenticationServerImpl{}

func getCookieDomain(ctx *gin.Context) string {
	host := ctx.Request.Host
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return ""
	}
	if strings.Contains(host, ":") {
		return host[:strings.Index(host, ":")]
	}
	return host
}

func NewAuthenticationServer(
	l *zap.Logger,
	svc authentication.Service,
	isSecure bool,
) *AuthenticationServerImpl {
	return &AuthenticationServerImpl{
		l:        l,
		service:  svc,
		IsSecure: isSecure,
	}
}

type AuthenticationServerImpl struct {
	l        *zap.Logger
	service  authentication.Service
	IsSecure bool
}

func (a *AuthenticationServerImpl) LoginChallenge(ctx *gin.Context) {
	var request gen_authentication.LoginChallengeRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	lc := contexts.NewLoggingContext(ctx, a.l)

	challenge, err := a.service.IssueChallenge(lc, request.Username)
	if err != nil {
		if errors.Is(err, authentication.ErrInvalidUsername) {
			ctx.JSON(403, gin.H{"error": "Authentication failed"})
			return
		}

		a.l.Error("failed to issue challenge", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "failed to issue challenge"})
		return
	}

	a.l.Sugar().Infof("Issued challenge for user %s", request.Username)
	serverResponse := gen_authentication.ChallengeServerResponse{
		Id:        challenge.ID,
		Challenge: challenge.Challenge,
		Salt:      challenge.User.Salt,
	}

	ctx.JSON(200, serverResponse)
}

func (a *AuthenticationServerImpl) LoginChallengeResponse(ctx *gin.Context) {
	var request gen_authentication.ChallengeClientResponse
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	lc := contexts.NewLoggingContext(ctx, a.l)
	result, err := a.service.ValidateChallenge(lc, request.Id, request.Challenge)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}
	host := getCookieDomain(ctx)
	// host := ""
	a.l.Info("setting cookie",
		zap.String("host", host),
		zap.Bool("is_secure", a.IsSecure),
		zap.Int("okey length", len(result.Okey)))
	// a.l.Info("okey", zap.String("okey", result.Okey))

	ctx.SetCookie("OKEY", result.Okey, 0, "/", host, a.IsSecure, true)
	ctx.SetCookie("RKEY", result.Rkey, 0, "/", host, a.IsSecure, true)
	ctx.SetCookie("UID", result.UserID, 0, "/", host, a.IsSecure, true)

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) IsAuthed(ctx *gin.Context) {
	lc := contexts.NewLoggingContext(ctx, a.l)
	token, err := ctx.Cookie("OKEY")
	if err != nil {
		a.l.Error("failed to get okey cookie", zap.Error(err))
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}
	a.l.Info("okey cookie", zap.String("okey", token))

	if token == "" {
		a.l.Error("okey cookie is empty")
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	claims, err := a.service.IsAuthenticated(lc, token)
	if err != nil {
		a.l.Error("failed to authenticate token", zap.Error(err))
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}
	a.l.Info("claims", zap.Any("claims", claims))
	ctx.Header("X-User-Id", claims.UserID.String())
	ctx.JSON(200, gen_authentication.AuthenticateResponse{
		UserId: claims.UserID,
	})
}

func (a *AuthenticationServerImpl) RefreshToken(ctx *gin.Context) {
	var request gen_authentication.RefreshRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	lc := contexts.NewLoggingContext(ctx, a.l)
	resp, err := a.service.AuthenticateWithToken(lc, request.RefreshToken)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	a.l.Sugar().Infow("User %s refreshed token successfully", request.RefreshToken)

	host := getCookieDomain(ctx)
	a.l.Info("setting cookie", zap.String("host", host))
	ctx.SetCookie("OKEY", resp.AuthToken, 0, "/", host, true, true)

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) VerifyKey(ctx *gin.Context) {
	var request gen_authentication.VerifyKeyRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	err := crypto.VerifyGCMAESKey(request.Key)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid key"})
		return
	}

	ctx.JSON(200, gin.H{"success": true})
}

func (a AuthenticationServerImpl) Register(ctx *gin.Context) {
	var req gen_authentication.RegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		a.l.Error("failed to bind register request", zap.Error(err))
		ctx.JSON(400, gin.H{"error": "bad register request"})
		return
	}
	lc := contexts.NewLoggingContext(ctx, a.l)
	auth, err := a.service.Register(lc, authenticationv1.RegistrationParam{
		Username: req.Username,
		Key:      req.Key,
		Email:    req.Email,
		Secret:   req.Secret,
	})
	if err != nil {
		a.l.Error("failed to register user", zap.Error(err))
		ctx.JSON(500, gin.H{"error": "failed to register user"})
		return
	}
	host := getCookieDomain(ctx)
	a.l.Info("setting cookie", zap.String("host", host))
	ctx.SetCookie("OKEY", auth.Okey, 0, "/", host, a.IsSecure, true)
	ctx.SetCookie("RKEY", auth.Rkey, 0, "/", host, a.IsSecure, true)
	ctx.SetCookie("UID", auth.UserID, 0, "/", host, a.IsSecure, true)

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) AuthenticateToken(ctx *gin.Context) {
	lc := contexts.NewLoggingContext(ctx, a.l)
	_, err := a.service.AuthenticateWithToken(lc, ctx.GetHeader("OKEY"))
	if err != nil {
		ctx.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	ctx.JSON(200, gin.H{})
}

func (a *AuthenticationServerImpl) Logout(ctx *gin.Context) {
	ctx.SetCookie("OKEY", "", -1, "/", "", true, true)
	ctx.SetCookie("RKEY", "", -1, "/", "", true, true)
	ctx.SetCookie("UID", "", -1, "/", "", true, true)
	ctx.JSON(200, gin.H{})
}
