package authenticationv1

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/datav1"
	userdata "github.com/ooqls/go-auth/internal/usersv1"
	"go.uber.org/zap"
)

func AuthMiddleware(
	authenticator Authenticator,
	userReader userdata.Reader,
	l *zap.Logger) func(c *gin.Context) {

	return func(c *gin.Context) {
		lc := contexts.NewLoggingContext(c, l)
		token, err := c.Cookie("OKEY")
		if err != nil {
			c.JSON(401, gin.H{"error": "Authentication failed"})
			return
		}

		if token == "" {
			c.JSON(401, gin.H{"error": "Authentication failed"})
			return
		}

		claims, err := authenticator.IsAuthenticated(lc, token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Authentication failed"})
			return
		}

		cl := *l.With(zap.String("token", token), zap.String("user_id", claims.UserID.String()))

		userObj, err := userReader.GetUser(lc, uuid.UUID(claims.UserID))
		if err != nil {
			cl.Error("failed to get user", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if userObj == nil {
			cl.Info("user does not exist")
			c.JSON(500, gin.H{"error": "user not found"})
			return
		}

		authCtx := authorizationv1.NewAuthorizationContext(*userObj)

		c.Set("authorization_context", authCtx)
		c.Next()
	}

}

func TrustedMiddleware(
	factory datav1.Factory,
	l *zap.Logger) func(c *gin.Context) {
	userReader := factory.NewUserReader()
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-Id")
		l.Info("userID", zap.String("userID", userID))
		log.Printf("headers: %v", c.Request.Header)
		if userID == "" {
			c.Set("authorization_context", authorizationv1.NewUnauthenticatedContext(c))
			c.Next()
			return
		}

		userIDUUID, err := uuid.Parse(userID)
		if err != nil {
			l.Error("failed to parse user ID", zap.Error(err))
			c.JSON(401, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		lc := contexts.NewLoggingContext(c, l)
		userObj, err := userReader.GetUser(lc, userIDUUID)
		if err != nil {
			l.Error("failed to get user", zap.Error(err))
			c.JSON(500, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}
		l.Info("setting authorization context", zap.String("user_id", userIDUUID.String()))
		authCtx := authorizationv1.NewAuthorizationContext(*userObj)
		c.Set("authorization_context", authCtx)
		c.Next()
	}
}
