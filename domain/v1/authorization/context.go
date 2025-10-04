package authorization

import (
	"context"

	"github.com/ooqls/go-auth/records"
	authv1 "github.com/ooqls/go-auth/records"
)

const (
	InternalOperationUserID = "internal"
)

type Context struct {
	context.Context
	User                authv1.UserAgg
	Roles               []authv1.RoleAgg
	Domain              string
	internalOperation bool
}

func NewAuthorizationContext(user authv1.UserAgg) Context {
	return Context{
		User: user,
	}
}

func NewInternalOperationContext(ctx context.Context) Context {
	return Context{
		Context:             ctx,
		internalOperation: true,
		User: authv1.UserAgg{
			UserId: records.NewUserID(),
			Roles:  []authv1.RoleAgg{},
		},
	}
}

func (a *Context) IsInternalOperation() bool {
	return a.internalOperation
}

func (a *Context) GetUserID() authv1.UserId {
	return a.User.UserId
}

func (a *Context) GetRoles() []authv1.RoleAgg {
	return a.Roles
}
