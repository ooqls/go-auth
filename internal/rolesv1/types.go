package rolesv1

import (
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
)

var Metadata = corev1.Metadata{
	Group: "roles",
	Kind:  Kind,
}

func FromDatagenRole(role datagen.Rolesv1) Role {
	return Role{
		Object: corev1.Object{
			Metadata: Metadata,
			Name:     role.Name,
		},
		Description: role.Description,
		Hierarchy:   role.Hierarchy,
	}
}

type Role struct {
	corev1.Object
	Description string `json:"description"`
	Hierarchy   int32  `json:"hierarchy"`
}
