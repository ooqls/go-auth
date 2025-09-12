package users

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ooqls/go-auth/records/gen"
)

var _ Writer = &SQLWriter{}

//go:generate go run github.com/golang/mock/mockgen -source=user_writer.go -destination=mocks/mock_user_writer.go -package=mocks -mock_names=UserWriter=MockUserWriter
type Writer interface {
	CreateUser(ctx context.Context, user gen.Authv1User) error
	DeleteUser(ctx context.Context, id UserId) error
	UpdateUser(ctx context.Context, user gen.Authv1User) error
}

func NewSQLWriter(db *sqlx.DB) *SQLWriter {
	return &SQLWriter{
		query: gen.New(db),
	}
}

type SQLWriter struct {
	query *gen.Queries
}

func (w *SQLWriter) CreateUser(ctx context.Context, user gen.Authv1User) error {
	_, err := w.query.CreateUser(ctx, gen.CreateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Salt:     user.Salt,
		Key:      user.Key,
	})
	return err
}

func (w *SQLWriter) DeleteUser(ctx context.Context, id UserId) error {
	return w.query.DeleteUser(ctx, id)
}

func (w *SQLWriter) UpdateUser(ctx context.Context, user gen.Authv1User) error {
	err := w.query.UpdateUser(ctx, gen.UpdateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Key:      user.Key,
	})
	return err
}
