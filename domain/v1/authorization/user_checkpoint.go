package authorization

import (
	"github.com/ooqls/go-auth/records"
	authv1 "github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/roles"
)

type UserCheckpoint struct {
	name     string
	rr       roles.Reader
	isAuthed func(ctx *Context, action Action, role authv1.User) (bool, error)
}

func (c *UserCheckpoint) IsAuthorized(ctx *Context, action Action, user authv1.User) (bool, error) {
	return c.isAuthed(ctx, action, user)
}

func (c *UserCheckpoint) GetName() string {
	return c.name
}

func HasHigherRole(targetUserRoles []records.Role) *UserCheckpoint {
	return &UserCheckpoint{
		name: "has_higher_role",
		isAuthed: func(ctx *Context, action Action, targetUser records.User) (bool, error) {
			highestRoleHierarchy := int32(0)
			for _, r := range ctx.GetRoles() {
				if r.RoleHierarchy > highestRoleHierarchy {
					highestRoleHierarchy = r.RoleHierarchy
				}
			}

			for _, r := range targetUserRoles {
				if r.RoleHierarchy > highestRoleHierarchy {
					return false, nil
				}
			}

			return true, nil
		},
	}
}
