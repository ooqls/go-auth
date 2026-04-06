package permissionsv1

import (
	"context"
	"strings"

	"github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
)

var _ Writer = &SQLWriter{}

//go:generate go run github.com/golang/mock/mockgen -source=writer.go -destination=mocks/mock_writer.go -package=mocks
type Writer interface {
	CreatePermission(name, group, kind string, actions []string) error
	DeletePermission(name, group, kind string) error
}

func NewSQLWriter(q *datagen.Queries) *SQLWriter {
	return &SQLWriter{q: q}
}

type SQLWriter struct {
	q *datagen.Queries
}

func (w *SQLWriter) CreatePermission(name, group, kind string, actions []string) error {
	_, err := w.q.CreatePermission(context.Background(), datagen.CreatePermissionParams{
		Name:    name,
		Group:   group,
		Kind:    kind,
		Actions: strings.Join(actions, ","),
	})
	if err != nil {
		return v1.ErrInternal(err, v1.M{"name": name, "group": group, "kind": kind})
	}
	return nil
}

func (w *SQLWriter) DeletePermission(name, group, kind string) error {
	err := w.q.DeletePermission(context.Background(), datagen.DeletePermissionParams{
		Name:  name,
		Group: group,
		Kind:  kind,
	})
	if err != nil {
		return v1.ErrInternal(err, v1.M{"name": name, "group": group, "kind": kind})
	}
	return nil
}
