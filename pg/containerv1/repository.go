package containerv1

import (
	"context"
	"fmt"

	"github.com/braumsmilk/go-auth/pg"
	"github.com/braumsmilk/go-auth/pg/query/insert"
	"github.com/braumsmilk/go-auth/pg/userv1"
	"github.com/braumsmilk/go-auth/pg/tables"
	"go.uber.org/zap"
)

var _ ContainerRepository = &PostgresRepository{}

var (
	ErrUserNotFound error = fmt.Errorf("user not found")
)

type ContainerRepository interface {
	GetAllContainers(ctx context.Context) ([]Container, error)
	GetContainer(ctx context.Context, id Id) (*Container, error)
	QueryContainers(ctx context.Context, query string) ([]Container, error)
	GetJoinedContainers(ctx context.Context, userid userv1.Id) ([]Container, error)
	GetUsersInContainer(ctx context.Context, id Id, page, pagesize int) ([]userv1.User, error)
	AddUserToContainer(ctx context.Context, userid userv1.Id, conainerid Id) error
	RemoveUserFromContainer(ctx context.Context, userid userv1.Id, containerid Id) error
	CreateContainer(ctx context.Context, name string, ct ContainerType) (Id, error)
	DeleteContainer(ctx context.Context, id Id) error
}

func NewDefaultRepository() ContainerRepository {
	return &PostgresRepository{}
}

type PostgresRepository struct{}

func (*PostgresRepository) GetJoinedContainers(ctx context.Context, userid userv1.Id) ([]Container, error) {
	dbConn := pg.Get()
	query := "SELECT " + tables.Containers + ".* FROM " + tables.Containers + " RIGHT JOIN " +
		tables.ContainerUserMapping + " c ON container_id = c.container_id WHERE c.user_id = $1"
	rows, err := dbConn.QueryxContext(ctx, query, userid)
	if err != nil {
		return nil, fmt.Errorf("failed to query containers that user has joined: %v", err)
	}

	containers, err := pg.ParseRows[Container](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rows into container objects: %v", err)
	}

	return containers, nil
}

func (*PostgresRepository) QueryContainers(ctx context.Context, query string) ([]Container, error) {
	dbConn := pg.Get()
	var containers []Container
	queryStr := "SELECT * FROM " + tables.Containers + " WHERE name LIKE ?"
	err := dbConn.SelectContext(ctx, containers, queryStr, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query containers: %v", err)
	}

	// containers, err := pg.ParseRows[Container](rows)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse container rows into container objects: %v", err)
	// }

	return containers, nil
}

func (*PostgresRepository) GetUsersInContainer(ctx context.Context, containerid Id, page, pagesize int) ([]userv1.User, error) {
	dbConn := pg.Get()
	query := "SELECT id, name FROM " + tables.ContainerUserMapping +
		" mapping RIGHT JOIN " + tables.Users +
		" u ON u.id = mapping.user_id WHERE container_id = $1 LIMIT $2 OFFSET $3;"

	rows, err := dbConn.QueryxContext(ctx, query, containerid, pagesize, page*pagesize)
	if err != nil {
		return nil, fmt.Errorf("failed to query for users in container: %v", err)
	}

	users, err := pg.ParseRows[userv1.User](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rows into user object: %v", err)
	}

	return users, nil
}

func (r *PostgresRepository) GetAllContainers(ctx context.Context) ([]Container, error) {
	dbConn := pg.Get()
	query := fmt.Sprintf("SELECT * FROM %s;", tables.Containers)
	rows, err := dbConn.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query container database: %v", err)
	}

	containers, err := pg.ParseRows[Container](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rows into containers: %v", err)
	}

	return containers, nil
}

func (PostgresRepository) GetContainer(ctx context.Context, id Id) (*Container, error) {
	dbConn := pg.Get()
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1;", tables.Containers)
	rows, err := dbConn.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query for container: %v", err)
	}

	c, err := pg.ParseRows[Container](rows)
	if len(c) > 1 {
		return nil, pg.ErrTooManyRows
	}

	if len(c) == 1 {
		return &c[0], nil
	}

	return nil, nil
}

func (PostgresRepository) CreateContainer(ctx context.Context, name string, ct ContainerType) (Id, error) {
	q := insert.BuildInsertQuery(tables.Containers, "name", "container_type")
	zap.L().Info("executing query", zap.String("query", q))

	rows, err := pg.Get().NamedQueryContext(ctx, q, Container{Name: name, ContainerType: ct})
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert: %v", err)
	}

	ids, err := pg.ParseRows[Container](rows)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rows into id: %v", err)
	}

	if len(ids) == 0 {
		return 0, ErrUserNotFound
	}

	if len(ids) > 1 {
		return 0, pg.ErrTooManyRows
	}

	return ids[0].Id, nil
}

func (PostgresRepository) DeleteContainer(ctx context.Context, id Id) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tables.Containers)
	_, err := pg.Get().ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %v", err)
	}

	return nil
}

func (PostgresRepository) AddUserToContainer(ctx context.Context, userid userv1.Id, containerId Id) error {
	q := fmt.Sprintf("INSERT INTO %s (\"container_id\", \"user_id\") VALUES ($1, $2);", tables.ContainerUserMapping)
	_, err := pg.Get().ExecContext(ctx, q, containerId, userid)
	if err != nil {
		return fmt.Errorf("failed to execute insertion : %v", err)
	}

	return nil
}

func (PostgresRepository) RemoveUserFromContainer(ctx context.Context, userid userv1.Id, containerId Id) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE container_id = $1 AND user_id = $2;", tables.ContainerUserMapping)
	_, err := pg.Get().ExecContext(ctx, q, containerId, userid)
	if err != nil {
		return fmt.Errorf("failed to execute deletion: %v", err)
	}

	return nil
}
