package permissionbindingsv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
)

type Permissionbindingv1 struct {
	corev1.Metadata
	RoleID       uuid.UUID
	Permission string
}
