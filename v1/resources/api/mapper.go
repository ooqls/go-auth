package resourcesapi

import (
	"github.com/ooqls/go-auth/internal/resourcesv1"
	v1gen "github.com/ooqls/go-auth/v1/gen"
)

// toGenResource translates a domain Resourcev1 into the generated OpenAPI
// model. All resources handlers must serialize through this function so the
// domain -> API mapping lives in exactly one place.
func toGenResource(res resourcesv1.Resourcev1) v1gen.Resource {
	id := res.Id
	updatedAt := res.UpdatedAt
	return v1gen.Resource{
		Id:        &id,
		Name:      res.Name,
		Group:     res.Group,
		Kind:      res.Kind,
		CreatedAt: res.CreatedAt,
		UpdatedAt: &updatedAt,
	}
}

// toGenResourceList translates a slice of domain resources into generated
// OpenAPI models.
func toGenResourceList(resources []resourcesv1.Resourcev1) []v1gen.Resource {
	items := make([]v1gen.Resource, 0, len(resources))
	for _, res := range resources {
		items = append(items, toGenResource(res))
	}
	return items
}
