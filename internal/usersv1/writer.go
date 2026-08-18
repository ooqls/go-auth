package usersv1

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
)

var _ Writer = &SQLWriter{}

//go:generate go run go.uber.org/mock/mockgen -source=writer.go -destination=mocks/mock_user_writer.go -package=mocks -mock_names=UserWriter=MockUserWriter
type Writer interface {
	CreateUser(ctx contexts.LContext, user datagen.CreateUserParams) (*datagen.Usersv1, error)
	DeleteUser(ctx contexts.LContext, id uuid.UUID) error
	UpdateUser(ctx contexts.LContext, id uuid.UUID, params datagen.UpdateUserParams) error
}

func NewSQLWriter(db *datagen.Queries) *SQLWriter {
	return &SQLWriter{
		query: db,
	}
}

type SQLWriter struct {
	query *datagen.Queries
}

func (w *SQLWriter) CreateUser(ctx contexts.LContext, params datagen.CreateUserParams) (*datagen.Usersv1, error) {
	user, err := w.query.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (w *SQLWriter) DeleteUser(ctx contexts.LContext, id uuid.UUID) error {
	err := w.query.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return v1.ErrNotFound(err, v1.M{
				"user_id": id,
			})
		}
		return v1.ErrInternal(err, v1.M{
			"user_id": id,
		})
	}
	return nil
}

func (w *SQLWriter) UpdateUser(ctx contexts.LContext, id uuid.UUID, params datagen.UpdateUserParams) error {
	err := w.query.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return v1.ErrNotFound(err, v1.M{
				"user_id": id,
			})
		}
		return v1.ErrInternal(err, v1.M{
			"user_id": id,
		})
	}
	return nil
}
