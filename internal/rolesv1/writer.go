package rolesv1

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/go-auth/internal/rolesv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
)

var _ Writer = &SQLWriter{}

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_role_writer.go -package=mocks -mock_names=RoleWriter=MockRoleWriter
type Writer interface {
	CreateRole(ctx context.Context, name, description string, hierarchy int32) (*Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, name, description *string, hierarchy *int32) (*Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	// AssignRole(ctx context.Context, id records.RoleId, userID records.UserId) error
	// GetRole(ctx context.Context, id uuid.UUID) (*datagen.Rolesv1, error)
	// GetRoles(ctx context.Context, limit, offset int32) ([]datagen.Rolesv1, error)
	// GetRoleByName(ctx context.Context, name string) (*datagen.Rolesv1, error)
	// // AddRoleToUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
	// RemoveRoleFromUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error
}

func NewSQLWriter(db *datagen.Queries) *SQLWriter {
	return &SQLWriter{
		q: db,
	}
}

type SQLWriter struct {
	q *datagen.Queries
}

func (r *SQLWriter) CreateRole(ctx context.Context, name, description string, hierarchy int32) (*Role, error) {
	role, err := r.q.CreateRole(ctx, datagen.CreateRoleParams{
		Name:        name,
		Description: description,
		Hierarchy:   hierarchy,
	})
	if err != nil {
		return nil, err
	}
	rolev1 := FromDatagenRole(role)
	return &rolev1, nil
}

func (r *SQLWriter) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description *string, hierarchy *int32) (*Role, error) {
	params := datagen.UpdateRoleParams{
		ID: roleID,
	}

	if name != nil {
		params.Name = pgtype.Text{
			String: *name,
			Valid:  true,
		}
	}

	if description != nil {
		params.Description = pgtype.Text{
			String: *description,
		}
	}

	if hierarchy != nil {
		params.Hierarchy = pgtype.Int4{
			Int32: *hierarchy,
			Valid: true,
		}
	}

	role, err := r.q.UpdateRole(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, v1.ErrNotFound(err, v1.M{
				"role_id": params.ID,
			})
		}
		return nil, v1.ErrInternal(err, v1.M{
			"role_id": params.ID,
		})
	}
	rolev1 := FromDatagenRole(role)
	return &rolev1, nil
}

// func (r *SQLWriter) GetRole(ctx context.Context, id uuid.UUID) (*datagen.Rolesv1, error) {
// 	role, err := r.q.GetRole(ctx, id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &role, nil
// }

// func (r *SQLWriter) GetRoles(ctx context.Context, limit, offset int32) ([]datagen.Rolesv1, error) {
// 	roles, err := r.q.GetRoles(ctx, datagen.GetRolesParams{
// 		Limit:  limit,
// 		Offset: offset,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return roles, nil
// }

// func (r *SQLWriter) GetRoleByName(ctx context.Context, name string) (*datagen.Rolesv1, error) {
// 	role, err := r.q.GetRolesByName(ctx, name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &role, nil
// }

func (r *SQLWriter) DeleteRole(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteRole(ctx, id)
	return err
}

// func (r *SQLWriter) AddRoleToUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
// 	err := r.q.AddRoleToUser(ctx, datagen.AddRoleToUserParams{
// 		UserID: userId,
// 		RoleID: roleId,
// 	})
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return v1.ErrNotFound(err, v1.M{
// 				"user_id": userId,
// 				"role_id": roleId,
// 			})
// 		}
// 		return v1.ErrInternal(err, v1.M{
// 			"user_id": userId,
// 			"role_id": roleId,
// 		})
// 	}
// 	return err
// }

// func (r *SQLWriter) RemoveRoleFromUser(ctx context.Context, userId uuid.UUID, roleId uuid.UUID) error {
// 	err := r.q.RemoveRoleFromUser(ctx, gen.RemoveRoleFromUserParams{
// 		UserID: userId,
// 		RoleID: roleId,
// 	})
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return v1.ErrNotFound(err, v1.M{
// 				"user_id": userId,
// 				"role_id": roleId,
// 			})
// 		}
// 		return v1.ErrInternal(err, v1.M{
// 			"user_id": userId,
// 			"role_id": roleId,
// 		})
// 	}
// 	return nil
// }
