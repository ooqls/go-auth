package permissions

import (
	"context"

	"github.com/ooqls/go-auth/records/gen"
)

var _ PermissionWriter = &SQLPermissionWriter{}

//go:generate go run github.com/golang/mock/mockgen -source=permission_writer.go -destination=mocks/mock_permission_writer.go -package=mocks -mock_names=PermissionWriter=MockPermissionWriter
type PermissionWriter interface {
	CreatePermission(ctx context.Context, p Permission) (*Permission, error)
	UpdatePermission(ctx context.Context, id PermissionId, p Permission) (*Permission, error)
	DeletePermission(ctx context.Context, id PermissionId) error
	AddPermissionToRole(ctx context.Context, roleId RoleId, permissionId PermissionId) error
	RemovePermissionFromRole(ctx context.Context, roleId RoleId, permissionId PermissionId) error
}

type SQLPermissionWriter struct {
	q gen.Queries
}

func (w *SQLPermissionWriter) CreatePermission(ctx context.Context, p Permission) (*Permission, error) {
	_, err := w.q.CreatePermission(ctx, gen.CreatePermissionParams{
		ResourceName:  p.ResourceName,
		ResourceGroup: p.ResourceGroup,
		ResourceKind:  p.ResourceKind,
		Actions:       p.Actions,
	})
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (w *SQLPermissionWriter) UpdatePermission(ctx context.Context, id PermissionId, p Permission) (*Permission, error) {
	permission, err := w.q.UpdatePermission(ctx, gen.UpdatePermissionParams{
		ID:            id,
		ResourceName:  p.ResourceName,
		ResourceGroup: p.ResourceGroup,
		ResourceKind:  p.ResourceKind,
		Actions:       p.Actions,
	})
	if err != nil {
		return nil, err
	}

	return &Permission{
		ID:            permission.ID,
		ResourceName:  permission.ResourceName,
		ResourceGroup: permission.ResourceGroup,
		ResourceKind:  permission.ResourceKind,
		Actions:       permission.Actions,
	}, nil
}

func (w *SQLPermissionWriter) DeletePermission(ctx context.Context, id PermissionId) error {
	return w.q.DeletePermission(ctx, id)
}

func (w *SQLPermissionWriter) AddPermissionToRole(ctx context.Context, roleId RoleId, permissionId PermissionId) error {
	return w.q.AddPermissionToRole(ctx, gen.AddPermissionToRoleParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	})
}

func (w *SQLPermissionWriter) RemovePermissionFromRole(ctx context.Context, roleId RoleId, permissionId PermissionId) error {
	return w.q.RemovePermissionFromRole(ctx, gen.RemovePermissionFromRoleParams{
		RoleID:       roleId,
		PermissionID: permissionId,
	})
}
