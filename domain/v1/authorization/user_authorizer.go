package authorization

import (
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/users"
)

type UserAuthorizer interface {
	IsAuthorizedToPerformUserAction(ctx *Context, action Action, user records.User) error
}

type UserAuthorizerImpl struct{}

func NewUserAuthorizerImpl(ur users.Reader) UserAuthorizer {
	return &UserAuthorizerImpl{}
}

func (ua *UserAuthorizerImpl) IsAuthorizedToPerformUserAction(ctx *Context, action Action, targetUser records.User) error {
	checkpoints := []*ResourceCheckpoint{}

	if ctx.GetUserID() != targetUser.ID && !ctx.IsInternalOperation() {
		checkpoints = append(checkpoints, UserHasResourcePermission())
	}

	for _, checkpoint := range checkpoints {
		if !checkpoint.IsAuthorized(ctx, action, NewUserResource(targetUser)) {
			return ErrPermissionDenied
		}
	}

	return nil
}
