package authorization

import (
	"strings"

	authv1 "github.com/ooqls/go-auth/records"
)

type RoleCheckpoint struct {
	name     string
	isAuthed func(ctx *Context, action Action, role authv1.Role) bool
}

func (c *RoleCheckpoint) IsAuthorized(ctx *Context, action Action, role authv1.Role) bool {
	return c.isAuthed(ctx, action, role)
}

func (c *RoleCheckpoint) GetName() string {
	return c.name
}

func IsRoleHierarchyGreaterThan() *RoleCheckpoint {
	return &RoleCheckpoint{
		name: "is_role_hierarchy_greater_than",
		isAuthed: func(ctx *Context, action Action, role authv1.Role) bool {
			isAuthed := false
			for _, r := range ctx.GetRoles() {
				if r.RoleHierarchy > role.RoleHierarchy {
					isAuthed = true
					break
				}
			}

			return isAuthed
		},
	}
}

func UserHasRolePermission() *RoleCheckpoint {
	return &RoleCheckpoint{
		name: "user_has_permission",
		isAuthed: func(ctx *Context, action Action, role authv1.Role) bool {
			for _, r := range ctx.GetRoles() {
				for _, p := range r.Permissions {
					if (p.ResourceGroup == authv1.GroupAuth || p.ResourceGroup == "*") &&
						(p.ResourceKind == authv1.KindRole || p.ResourceKind == "*") &&
						(p.ResourceName == role.RoleName || p.ResourceName == "*") &&
						(strings.Contains(p.Actions, string(action)) || p.Actions == "*") {
						return true
					}
				}
			}

			return false
		},
	}
}

func UserCanModifyUser() *RoleCheckpoint {
	return &RoleCheckpoint{
		name: "user_can_modify_user",
		isAuthed: func(ctx *Context, action Action, role authv1.Role) bool {
			return false
		},
	}
}
