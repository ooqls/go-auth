package rolebindingsv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1/datagen"
)

type Rolebinding struct {
	corev1.Metadata
	RoleID uuid.UUID `json:"role_id"`
	UserID uuid.UUID `json:"user_id"`
}

func fromDatagenRoleBinding(rb datagen.Rolebindingsv1) Rolebinding {
	return Rolebinding{
		Metadata: corev1.RoleBindingsV1,
		RoleID:   rb.RoleID,
		UserID:   rb.UserID,
	}
}
