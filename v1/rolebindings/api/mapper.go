package rolebindingsapi

import (
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/ooqls/go-auth/v1/rolebindings/api/gen_rolebindings"
)

// toGenRoleBinding translates a domain Rolebinding into the generated OpenAPI
// model. All rolebindings handlers must serialize through this function so the
// domain -> API mapping lives in exactly one place. The domain type's JSON
// tags (role_id/user_id) differ from the API contract (roleID/userID), so
// serializing the domain object directly would produce the wrong field names.
func toGenRoleBinding(rb rolebindingsv1.Rolebinding) gen_rolebindings.RoleBinding {
	roleID := rb.RoleID
	userID := rb.UserID
	return gen_rolebindings.RoleBinding{
		RoleID: &roleID,
		UserID: &userID,
	}
}

// toGenRoleBindingList translates a slice of domain role bindings into
// generated OpenAPI models.
func toGenRoleBindingList(bindings []rolebindingsv1.Rolebinding) []gen_rolebindings.RoleBinding {
	items := make([]gen_rolebindings.RoleBinding, 0, len(bindings))
	for _, rb := range bindings {
		items = append(items, toGenRoleBinding(rb))
	}
	return items
}
