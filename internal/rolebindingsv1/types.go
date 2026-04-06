package rolebindingsv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1/datagen"
)

var Metadata corev1.Metadata = corev1.Metadata{
	Group: "authv1",
	Kind:  "Rolebindingsv1",
}

type Rolebinding struct {
	corev1.Metadata
	RoleID uuid.UUID `json:"role_id"`
	UserID uuid.UUID `json:"user_id"`
}

func fromDatagenRoleBinding(rb datagen.Rolebindingsv1) Rolebinding {
	return Rolebinding{
		Metadata: Metadata,
		RoleID:   rb.RoleID,
		UserID:   rb.UserID,
	}
}
