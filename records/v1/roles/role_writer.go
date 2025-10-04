package roles

import (
	"context"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/v1/gen"
)

var _ Writer = &SQLWriter{}

//go:generate go run github.com/golang/mock/mockgen -source=role_writer.go -destination=mocks/mock_role_writer.go -package=mocks -mock_names=RoleWriter=MockRoleWriter
type Writer interface {
	CreateRole(ctx context.Context, r Role) (*Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, r Role) (*Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	AddRoleToUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
}

type SQLWriter struct {
	q gen.Queries
}

func (r *SQLWriter) CreateRole(ctx context.Context, role Role) (*Role, error) {
	role, err := r.q.CreateRole(ctx, gen.CreateRoleParams{
		RoleName:    role.RoleName,
		Description: role.Description,
	})
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SQLWriter) UpdateRole(ctx context.Context, id uuid.UUID, role Role) (*Role, error) {
	role, err := r.q.UpdateRole(ctx, gen.UpdateRoleParams{
		ID:          id,
		RoleName:    role.RoleName,
		Description: role.Description,
	})
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SQLWriter) DeleteRole(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.DeleteRole(ctx, id)
	return err
}

func (r *SQLWriter) AddRoleToUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
	err := r.q.AddRoleToUser(ctx, gen.AddRoleToUserParams{
		UserID: userId,
		RoleID: roleId,
	})
	return err
}

func (r *SQLWriter) RemoveRoleFromUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
	err := r.q.RemoveRoleFromUser(ctx, gen.RemoveRoleFromUserParams{
		UserID: userId,
		RoleID: roleId,
	})
	return err
}
