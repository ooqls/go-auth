package permissionsapi

import (
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/v1/gen"
)

// toGenPermission translates a domain Permission into the generated OpenAPI
// model. All permissions handlers must serialize through this function so the
// domain -> API mapping lives in exactly one place.
//
// Note: the domain Permission only carries the permission string, so the
// generated Id/CreatedAt/UpdatedAt fields are intentionally left at their zero
// values until the domain model surfaces them.
func toGenPermission(p permissionsv1.Permission) gen.Permission {
	return gen.Permission{
		Permission: p.Permission,
	}
}

// toGenPermissionList translates a slice of domain permissions into generated
// OpenAPI models.
func toGenPermissionList(perms []permissionsv1.Permission) []gen.Permission {
	items := make([]gen.Permission, 0, len(perms))
	for _, p := range perms {
		items = append(items, toGenPermission(p))
	}
	return items
}
