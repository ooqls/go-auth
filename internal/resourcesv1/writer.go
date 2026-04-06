package resourcesv1

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1/datagen"
)

type Writer interface {
	CreateResource(ctx context.Context, group, kind, name, description string) (*Resourcev1, error)
	UpdateResource(ctx context.Context, id uuid.UUID, name, description *string) (*Resourcev1, error)
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

func (w *SQLWriter) CreateResource(ctx context.Context, group, kind, name, description string) (*Resourcev1, error) {
	res, err := w.query.CreateResource(ctx, datagen.CreateResourceParams{
		ResourceKind:  kind,
		ResourceGroup: group,
		ResourceName:  name,
		Description:   description,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &Resourcev1{
		Object: corev1.Object{
			Metadata: corev1.Metadata{
				Group: res.ResourceGroup,
				Kind:  res.ResourceKind,
			},
			Name: res.ResourceName,
			Id:   res.ID,
		},
		Description: res.Description,
		CreatedAt:   res.CreatedAt.Time,
		UpdatedAt:   res.UpdatedAt.Time,
	}, err
}

func (w *SQLWriter) UpdateResource(ctx context.Context,
	id uuid.UUID, name, description *string) (*Resourcev1, error) {
	params := datagen.UpdateResourceParams{
		ID: id,
	}
	if name != nil {
		params.Name = *name
	}
	if description != nil {
		params.Description = *description
	}

	res, err := w.query.UpdateResource(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &Resourcev1{
		Object: corev1.Object{
			Metadata: corev1.Metadata{
				Group: res.ResourceGroup,
				Kind:  res.ResourceKind,
			},
			Id:   res.ID,
			Name: res.ResourceName,
		},
		Description: res.Description,
		CreatedAt:   res.CreatedAt.Time,
		UpdatedAt:   res.UpdatedAt.Time,
	}, nil
}
func (w *SQLWriter) DeleteResource(ctx context.Context, group, kind, name string) error {
	err := w.query.DeleteResource(ctx, datagen.DeleteResourceParams{
		ResourceName:  name,
		ResourceGroup: group,
		ResourceKind:  kind,
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
