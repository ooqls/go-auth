package authorizationv1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/usersv1"
	"go.uber.org/zap"
)

const (
	InternalOperationUserID = "internal"
)

type Context struct {
	context.Context
	User              usersv1.User
	Domain            string
	internalOperation bool
	l                 *zap.Logger
}

func FromGinContext(ctx *gin.Context) (*Context, bool) {
	authObj, ok := ctx.Get("authorization_context")
	if !ok {
		return nil, false
	}

	authCtx, ok := authObj.(Context)
	if !ok {
		return nil, false
	}

	return &authCtx, true
}

func NewAuthorizationContext(user usersv1.User) Context {
	return Context{
		l:    zap.L().With(zap.String("user_id", user.Id.String())),
		User: user,
	}
}

func NewInternalOperationContext(ctx context.Context) Context {
	return Context{
		Context:           ctx,
		internalOperation: true,
		l:                 log.NewLogger("internal_operation").With(zap.String("user_id", InternalOperationUserID)),
	}
}

func (a *Context) IsInternalOperation() bool {
	return a.internalOperation
}

func (a *Context) GetAuthedUser() usersv1.User {
	return a.User
}

func (a *Context) L() *zap.Logger {
	if a.l != nil {
		return a.l
	}

	return zap.L()
}
