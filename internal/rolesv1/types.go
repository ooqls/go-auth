package rolesv1

import (
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
)

func FromDatagenRole(role datagen.Rolesv1) Role {
	return Role{
		Object: corev1.Object{
			Metadata:  corev1.RolesV1,
			Name:      role.Name,
			Id:        role.ID,
			CreatedAt: role.CreatedAt.Time,
			UpdatedAt: role.UpdatedAt.Time,
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
