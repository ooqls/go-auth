package permissionsv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionsv1/datagen"
)

var Metadata corev1.Metadata = corev1.Metadata{
	Group: "Authv1",
	Kind:  "Permissionv1",
}

type Permission struct {
	corev1.Object
	Resource corev1.Object `json:"resource"`
	Actions  string        `json:"actions"`
}

func fromUserPermRow(id uuid.UUID, name, group, kind, actions string) Permission {
	return Permission{
		Object: corev1.Object{
			Metadata: Metadata,
			Id:       id,
			Name:     name,
		},
		Resource: corev1.Object{
			Metadata: corev1.Metadata{
				Group: group,
				Kind:  kind,
			},
			Name: name,
		},
		Actions: actions,
	}
}

func fromDatagenPermission(p datagen.Permissionsv1) *Permission {
	return &Permission{
		Object: corev1.Object{
			Metadata:  Metadata,
			Name:      p.Name,
			Id:        p.ID,
			CreatedAt: p.CreatedAt.Time,
			UpdatedAt: p.UpdatedAt.Time,
		},
		Resource: corev1.Object{
			Metadata: corev1.Metadata{
				Group: p.Group,
				Kind:  p.Kind,
			},
			Name: p.Name,
		},
		Actions: p.Actions,
	}
}
