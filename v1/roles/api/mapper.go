package rolesapi

import (
	"github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/roles"
)

// toGenRole translates a domain Role into the generated OpenAPI model. All
// roles handlers must serialize through this function so the domain -> API
// mapping lives in exactly one place.
func toGenRole(role roles.Role) gen.Role {
	id := role.Id
	return gen.Role{
		Id:          &id,
		Name:        role.Name,
		Description: role.Description,
		Hierarchy:   int(role.Hierarchy),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// toGenRoleList translates a slice of domain roles into generated OpenAPI
// models.
func toGenRoleList(list []roles.Role) []gen.Role {
	items := make([]gen.Role, 0, len(list))
	for _, role := range list {
		items = append(items, toGenRole(role))
	}
	return items
}
