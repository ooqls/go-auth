package rolesv1

import (
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
)

func FromDatagenRole(role datagen.Rolesv1) Role {
	return Role{
		Object: corev1.Object{
			Metadata: corev1.Metadata{
				Group: "roles",
				Kind:  Kind,
			},
			Name: role.Name,
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
