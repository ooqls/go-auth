package permissionsv1

import (
	"github.com/ooqls/go-auth/internal/corev1"
)

func NewPermission(actions string) *Permission {
	return &Permission{
		Permission: actions,
	}
}

type Permission struct {
	corev1.Metadata
	Permission string `json:"permission"`
}

func fromDatagenPermission(p string) *Permission {
	return &Permission{
		Metadata:   corev1.PermissionsV1,
		Permission: p,
	}
}
