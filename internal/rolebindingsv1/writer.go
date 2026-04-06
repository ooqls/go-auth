package rolebindingsv1

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/rolebindingsv1/datagen"
)

type Writer interface {
	AddRoleToUser(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) error
	RemoveRoleFromUser(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) error
}

type SQLWriter struct {
	q *datagen.Queries
}

func NewSQLWriter(db *datagen.Queries) *SQLWriter {
	return &SQLWriter{q: db}
}

func (w *SQLWriter) AddRoleToUser(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) error {
	return w.q.AddRoleToUser(ctx, datagen.AddRoleToUserParams{
		UserID: userId,
		RoleID: roleId,
	})
}

func (w *SQLWriter) RemoveRoleFromUser(ctx contexts.LContext, userId uuid.UUID, roleId uuid.UUID) error {
	err := w.q.RemoveRoleFromUser(ctx, datagen.RemoveRoleFromUserParams{
		UserID: userId,
		RoleID: roleId,
	})
	if err != nil {
		return err
	}
	return nil
}
