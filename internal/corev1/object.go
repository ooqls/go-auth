package corev1

import (
	"time"

	"github.com/google/uuid"
)

func FromID(id uuid.UUID, metadata Metadata) Object {
	return Object{
		Metadata: metadata,
		Id:       id,
	}
}

type Object struct {
	Metadata
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
