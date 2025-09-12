package authorization

import (
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/roles"
	"go.uber.org/zap"
)

type RoleAuthorizer struct {
	l *zap.Logger
}

func NewRoleAuthorizer(rr roles.Reader) *RoleAuthorizer {
	return &RoleAuthorizer{}
}

func (ra *RoleAuthorizer) IsAuthorizedToAssignRole(ctx *Context, action Action, role records.Role, targetUser records.User, targetUserRoles []records.Role) error {
	roleCheckpoints := []*RoleCheckpoint{
		UserHasRolePermission(),
		IsRoleHierarchyGreaterThan(),
		UserCanModifyUser(),
	}

	UserCheckpoints := []*UserCheckpoint{
		HasHigherRole(targetUserRoles),
	}

	for _, checkpoint := range roleCheckpoints {
		if !checkpoint.IsAuthorized(ctx, action, role) {
			ra.l.Error("user is not authorized to assign role", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	for _, checkpoint := range UserCheckpoints {
		isAuthed, err := checkpoint.IsAuthorized(ctx, action, targetUser)
		if err != nil {
			ra.l.Error("error checking user authorization", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()), zap.Error(err))
			return err
		}
		if !isAuthed {
			ra.l.Error("user is not authorized to assign role", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	return nil
}

func (ra *RoleAuthorizer) IsAuthorizedToUnassignRole(ctx *Context, action Action, role records.Role, targetUser records.User, targetUserRoles []records.Role) error {
	roleCheckpoints := []*RoleCheckpoint{
		UserHasRolePermission(),
		UserCanModifyUser(),
		IsRoleHierarchyGreaterThan(),
	}

	UserCheckpoints := []*UserCheckpoint{}

	if ctx.GetUserID() != targetUser.ID {
		UserCheckpoints = append(UserCheckpoints, HasHigherRole(targetUserRoles))
	}

	for _, checkpoint := range roleCheckpoints {
		if !checkpoint.IsAuthorized(ctx, action, role) {
			ra.l.Error("user is not authorized to unassign role", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	for _, checkpoint := range UserCheckpoints {
		isAuthed, err := checkpoint.IsAuthorized(ctx, action, targetUser)
		if err != nil {
			ra.l.Error("error checking user authorization", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()), zap.Error(err))
			return err
		}
		if !isAuthed {
			ra.l.Error("user is not authorized to unassign role", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	return nil
}

func (ra *RoleAuthorizer) IsAuthorizedToPerformRoleAction(ctx *Context, action Action, role records.Role) error {
	roleCheckpoints := []*RoleCheckpoint{
		UserHasRolePermission(),
		IsRoleHierarchyGreaterThan(),
	}

	for _, checkpoint := range roleCheckpoints {
		if !checkpoint.IsAuthorized(ctx, action, role) {
			ra.l.Error("user is not authorized to perform role action", zap.String("action", string(action)), zap.String("role", role.RoleName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	return nil
}
