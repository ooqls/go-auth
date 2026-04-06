package resourcesv1

import (
	"time"

	"github.com/ooqls/go-auth/internal/corev1"
)

var Metadata = corev1.Metadata{
	Group: "authv1",
	Kind:  "resources",
}

type Resourcev1 struct {
	corev1.Object
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
