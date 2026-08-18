package permissionbindingsv1

//go:generate go run go.uber.org/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
)

type Writer interface {
	AssignPermission(ctx contexts.LContext, roleID uuid.UUID, permission string) error
	UnassignPermission(ctx contexts.LContext, roleID uuid.UUID, permission string) error
	UnassignAllPermissions(ctx contexts.LContext, roleID uuid.UUID) error
	UnassignPermissionFromAllRoles(ctx contexts.LContext, permission string) error
}

func NewSQLWriter(q datagen.Queries) Writer {
	return &SQLWriter{
		q: q,
	}
}

type SQLWriter struct {
	q datagen.Queries
}

func (w *SQLWriter) AssignPermission(ctx contexts.LContext, roleID uuid.UUID, permission string) error {
	if err := w.q.AssignPermission(ctx, datagen.AssignPermissionParams{
		RoleID:       roleID,
		Permission: permission,
	}); err != nil {
		return err
	}

	return nil
}

func (w *SQLWriter) UnassignPermission(ctx contexts.LContext, roleID uuid.UUID, permission string) error {
	return w.q.UnassignPermission(ctx, datagen.UnassignPermissionParams{
		Permission: permission,
		RoleID:     roleID,
	})
}

func (w *SQLWriter) UnassignAllPermissions(ctx contexts.LContext, roleID uuid.UUID) error {
	return w.q.UnassignAllPermissions(ctx, roleID)
}

func (w *SQLWriter) UnassignPermissionFromAllRoles(ctx contexts.LContext, permission string) error {
	return w.q.UnassignFromAllRoles(ctx, permission)
}
