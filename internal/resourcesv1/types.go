package resourcesv1

import (
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
)

type Resourcev1 struct {
	corev1.Metadata
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func fromDatagenResource(res datagen.Resourcesv1) *Resourcev1 {
	return &Resourcev1{
		Metadata: corev1.Metadata{
			Group: res.Rgroup,
			Kind:  res.Kind,
		},
		Id:        res.ID,
		Name:      res.Name,
		CreatedAt: res.CreatedAt.Time,
		UpdatedAt: res.UpdatedAt.Time,
	}
}

func fromDatagenResourceList(resources []datagen.Resourcesv1) []Resourcev1 {
	resList := make([]Resourcev1, len(resources))
	for i, res := range resources {
		resList[i] = *fromDatagenResource(res)
	}
	return resList
}

