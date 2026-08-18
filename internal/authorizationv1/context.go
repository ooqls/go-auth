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
	authenticated     bool
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

func NewAuthorizationContext(ctx context.Context, user usersv1.User) Context {
	return Context{
		Context:       ctx,
		l:             log.NewLogger("authorization_context").With(zap.String("user_id", user.Id.String())),
		authenticated: true,
		User:          user,
	}
}

func NewInternalOperationContext(ctx context.Context) Context {
	return Context{
		Context:           ctx,
		internalOperation: true,
		authenticated:     false,
		l:                 log.NewLogger("internal_operation").With(zap.String("user_id", InternalOperationUserID)),
	}
}

func NewUnauthenticatedContext(ctx context.Context) Context {
	return Context{
		Context:       ctx,
		authenticated: true,
		l:             log.NewLogger("unauthenticated_operation").With(zap.String("user_id", "unauthenticated")),
	}
}

func (a *Context) IsInternalOperation() bool {
	return a.internalOperation
}

func (a *Context) IsAuthenticated() bool {
	return a.authenticated
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
