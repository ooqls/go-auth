package roles

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
	"github.com/ooqls/go-auth/v1/gen"
)

type CreateRoleParams struct {
	Name        string
	Description string
	Hierarchy   int32
	Permissions datagen.RolePermissions
}

func (p CreateRoleParams) ToDatagenCreateRoleParams() (datagen.CreateRoleParams, error) {
	return datagen.CreateRoleParams{
		Name:        p.Name,
		Description: p.Description,
		Hierarchy:   p.Hierarchy,
	}, nil
}

func (p CreateRoleParams) ToRole() Role {
	return Role{
		Object: corev1.Object{
			Metadata: corev1.RolesV1,
			Id:       uuid.New(),
			Name:     p.Name,
		},
		Description: p.Description,
		Hierarchy:   p.Hierarchy,
	}
}

type UpdateRoleParams struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Hierarchy   *int32
}

func (p UpdateRoleParams) ToDatagenUpdateRoleParams() (datagen.UpdateRoleParams, error) {
	param := datagen.UpdateRoleParams{
		ID: p.ID,
	}
	if p.Name != nil {
		param.Name = pgtype.Text{String: *p.Name, Valid: true}
	}

	if p.Description != nil {
		param.Description = pgtype.Text{String: *p.Description, Valid: true}
	}

	if p.Hierarchy != nil {
		param.Hierarchy = pgtype.Int4{Int32: *p.Hierarchy, Valid: true}
	}

	return param, nil

}

func ToDatagenPermissions(permissions []gen.Permission) []datagen.Permission {
	datagenPermissions := make([]datagen.Permission, len(permissions))
	for i, perm := range permissions {
		datagenPermissions[i] = datagen.Permission{
			Actions: perm.Actions,
			Name:    perm.ResourceName,
			Group:   perm.ResourceGroup,
			Kind:    perm.ResourceKind,
		}
	}

	return datagenPermissions
}

func FromDatagenPermissions(permissions []datagen.Permission) *[]gen.Permission {
	genPermissions := make([]gen.Permission, len(permissions))
	for i, perm := range permissions {
		genPermissions[i] = gen.Permission{
			Actions:       perm.Actions,
			ResourceName:  perm.Name,
			ResourceGroup: perm.Group,
			ResourceKind:  perm.Kind,
		}
	}

	return &genPermissions
}
