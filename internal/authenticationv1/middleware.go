package authenticationv1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/datav1"
	"go.uber.org/zap"
)

func GetFirstHeader(c *gin.Context, headerName ...string) string {
	for _, name := range headerName {
		if value := c.GetHeader(name); value != "" {
			return value
		}
	}
	return ""
}

func NoProxyMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		lctx := contexts.NewLoggingContext(c, log.NewLogger("no-proxy-middleware"))
		okey, err := c.Cookie("OKEY")
		if err != nil {
			lctx.L().Info("OKEY cookie not found, skipping authentication")
			c.Next()
			return
		}
		authenticator := GetAuthenticatorV1()
		claims, err := authenticator.AuthenticateWithToken(lctx, okey)
		if err != nil {
			lctx.L().Error("failed to authenticate with token", zap.Error(err))
			c.Next()
			return
		}

		c.Set("X-User-Id", claims.UserId.String())
		c.Next()
	}
}

func TrustedMiddleware(factory datav1.Factory) func(c *gin.Context) {
	userReader := factory.NewUserReader()
	l := log.NewLogger("trusted-middleware")
	return func(c *gin.Context) {
		var err error
		userID := GetFirstHeader(c, "X-User-Id")
		l.Info("userID header", zap.String("userID", userID))

		if userID == "" {
			userID = c.GetString("X-User-Id")
		}

		if userID == "" {
			c.Set("authorization_context", authorizationv1.NewUnauthenticatedContext(c))
			c.Next()
			return
		}

		l.Info("parsing user id")
		userIDUUID, err := uuid.Parse(userID)
		if err != nil {
			l.Error("failed to parse user ID", zap.Error(err))
			c.JSON(401, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		l.Info("retrieving user from database", zap.String("user_id", userIDUUID.String()))
		lc := contexts.NewLoggingContext(c, l)
		userObj, err := userReader.GetUser(lc, userIDUUID)
		if err != nil {
			l.Error("failed to get user", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		if userObj == nil {
			l.Info("user does not exist", zap.String("user_id", userID))
			c.JSON(401, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		l.Info("user found, setting authorization context", zap.String("user_id", userIDUUID.String()))
		authCtx := authorizationv1.NewAuthorizationContext(c, *userObj)
		c.Set("authorization_context", authCtx)
		c.Next()
	}
}
