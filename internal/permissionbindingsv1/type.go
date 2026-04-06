package permissionbindingsv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
)

var Metadata corev1.Metadata = corev1.Metadata{
	Group: "authv1",
	Kind:  "Permission",
}

type Permissionbindingv1 struct {
	corev1.Metadata
	RoleID       uuid.UUID
	PermissionID uuid.UUID
}
