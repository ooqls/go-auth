package resourcesv1

//go:generate go run go.uber.org/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
)

type Writer interface {
	CreateResource(ctx context.Context, group, kind, name string) (*Resourcev1, error)
	UpdateResource(ctx context.Context, group, kind, name string, newName string) (*Resourcev1, error)
	DeleteResource(ctx context.Context, group, kind, name string) error
	DeleteResourceById(ctx context.Context, id uuid.UUID) error
}

type SQLWriter struct {
	query *datagen.Queries
}

func NewSQLWriter(db *datagen.Queries) Writer {
	return &SQLWriter{
		query: db,
	}
}

func (w *SQLWriter) CreateResource(ctx context.Context, group, kind, name string) (*Resourcev1, error) {
	res, err := w.query.CreateResource(ctx, datagen.CreateResourceParams{
		ID:        uuid.New(),
		Name:      name,
		Rgroup:    group,
		Kind:      kind,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return FromDatagenResource(res), nil
}

func (w *SQLWriter) UpdateResource(ctx context.Context, group, kind, name string, newName string) (*Resourcev1, error) {
	res, err := w.query.UpdateResource(ctx, datagen.UpdateResourceParams{
		Rgroup: group,
		Kind:   kind,
		Name:   name,
		Name_2: newName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return FromDatagenResource(res), nil
}

func (w *SQLWriter) DeleteResource(ctx context.Context, group, kind, name string) error {
	err := w.query.DeleteResource(ctx, datagen.DeleteResourceParams{
		Name:   name,
		Rgroup: group,
		Kind:   kind,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (w *SQLWriter) DeleteResourceById(ctx context.Context, id uuid.UUID) error {
	err := w.query.DeleteResourceById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
