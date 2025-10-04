package roles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/roles"
)

type RoleRepository interface {
	CreateRole(ctx context.Context, role *records.Role) error
	DeleteRole(ctx context.Context, id string) error
	GetRole(ctx context.Context, id string) (*records.Role, error)
	ListRoles(ctx context.Context) ([]records.Role, error)
	UpdateRole(ctx context.Context, role *records.Role) error
}

type RolesRepositoryImpl struct {
	rr roles.Reader
	rw roles.Writer
}

func NewRolesRepositoryImpl(rr roles.Reader, rw roles.Writer) *RolesRepositoryImpl {
	return &RolesRepositoryImpl{rr: rr, rw: rw}
}

func (r *RolesRepositoryImpl) CreateRole(ctx context.Context, roleName, description string) error {
	role, err := r.rr.GetRoleByName(ctx, roleName)
	if err != nil {
		return err
	}

	if role != nil {
		return fmt.Errorf("role already exists")
	}

	_, err = r.rw.CreateRole(ctx, roles.Role{
		ID:          records.NewRoleID(),
		RoleName:    roleName,
		Description: description,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *RolesRepositoryImpl) DeleteRole(ctx context.Context, id records.RoleId) error {
	return r.rw.DeleteRole(ctx, id)
}

func (r *RolesRepositoryImpl) GetRole(ctx context.Context, id records.RoleId) (*records.Role, error) {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &roles.Role{
		ID:          role.ID,
		RoleName:    role.RoleName,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}
