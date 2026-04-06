package authorizationv1

import (
	"context"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/datav1"
)

type ContextFactory interface {
	New(ctx context.Context, userID uuid.UUID) *Context
}

type ContextFactoryImpl struct {
	factory datav1.Factory
}
