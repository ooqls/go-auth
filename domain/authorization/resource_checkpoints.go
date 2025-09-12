package authorization

import (
	"strings"

	"github.com/ooqls/go-auth/records"
)

var CoreResourceGroup = "core"

type Resource struct {
	ResourceGroup string
	ResourceKind  string
	ResourceName  string
}

func NewUserResource(user records.User) Resource {
	return Resource{
		ResourceGroup: CoreResourceGroup,
		ResourceKind:  "user",
		ResourceName:  user.ID.String(),
	}
}

func NewRoleResource(role records.Role) Resource {
	return Resource{
		ResourceGroup: CoreResourceGroup,
		ResourceKind:  "role",
		ResourceName:  role.ID.String(),
	}
}

type ResourceCheckpoint struct {
	name     string
	isAuthed func(ctx *Context, action Action, resource Resource) bool
}

func (c *ResourceCheckpoint) IsAuthorized(ctx *Context, action Action, resource Resource) bool {
	return c.isAuthed(ctx, action, resource)
}

func (c *ResourceCheckpoint) GetName() string {
	return c.name
}

func UserHasResourcePermission() *ResourceCheckpoint {
	return &ResourceCheckpoint{
		name: "is_resource_hierarchy_greater_than",
		isAuthed: func(ctx *Context, action Action, resource Resource) bool {
			for _, r := range ctx.GetRoles() {
				for _, p := range r.Permissions {
					if (p.ResourceGroup == resource.ResourceGroup || p.ResourceGroup == "*") &&
						(p.ResourceKind == resource.ResourceKind || p.ResourceKind == "*") &&
						(p.ResourceName == resource.ResourceName || p.ResourceName == "*") &&
						(strings.Contains(p.Actions, string(action)) || p.Actions == "*") {
						return true
					}
				}
			}

			return false
		},
	}
}
