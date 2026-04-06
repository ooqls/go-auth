package permissionbindingsv1

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
)

type Writer interface {
	AssignPermission(ctx contexts.LContext, roleID uuid.UUID, permissionID uuid.UUID) error
	UnassignPermission(ctx contexts.LContext, roleID uuid.UUID, permissionID uuid.UUID) error
	UnassignAllPermissions(ctx contexts.LContext, roleID uuid.UUID) error
	UnassignPermissionFromAllRoles(ctx contexts.LContext, permissionID uuid.UUID) error
}

func NewSQLWriter(q datagen.Queries) Writer {
	return &SQLWriter{
		q: q,
	}
}

type SQLWriter struct {
	q datagen.Queries
}

func (w *SQLWriter) AssignPermission(ctx contexts.LContext, roleID uuid.UUID, permissionID uuid.UUID) error {
	if err := w.q.AssignPermission(ctx, datagen.AssignPermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	}); err != nil {
		return err
	}

	return nil
}

func (w *SQLWriter) UnassignPermission(ctx contexts.LContext, roleID uuid.UUID, permissionID uuid.UUID) error {
	return w.q.UnassignPermission(ctx, datagen.UnassignPermissionParams{
		PermissionID: permissionID,
		RoleID:       roleID,
	})
}

func (w *SQLWriter) UnassignAllPermissions(ctx contexts.LContext, roleID uuid.UUID) error {
	return w.q.UnassignAllPermissions(ctx, roleID)
}

func (w *SQLWriter) UnassignPermissionFromAllRoles(ctx contexts.LContext, permissionID uuid.UUID) error {
	return w.q.UnassignFromAllRoles(ctx, permissionID)
}
