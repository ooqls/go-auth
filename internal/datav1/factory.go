package datav1

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/go-auth/internal/aggsv1"
	aggsdatagen "github.com/ooqls/go-auth/internal/aggsv1/datagen"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	permissionbindingsdatagen "github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	permissionsdatagen "github.com/ooqls/go-auth/internal/permissionsv1/datagen"
	resources "github.com/ooqls/go-auth/internal/resourcesv1"
	resourcesdatagen "github.com/ooqls/go-auth/internal/resourcesv1/datagen"
	rolebindings "github.com/ooqls/go-auth/internal/rolebindingsv1"
	rolebindingsdatagen "github.com/ooqls/go-auth/internal/rolebindingsv1/datagen"
	"github.com/ooqls/go-auth/internal/rolesv1"
	roles "github.com/ooqls/go-auth/internal/rolesv1"
	rolesdatagen "github.com/ooqls/go-auth/internal/rolesv1/datagen"
	users "github.com/ooqls/go-auth/internal/usersv1"
	usersdatagen "github.com/ooqls/go-auth/internal/usersv1/datagen"
	// permdata "github.com/ooqls/go-auth/v1/permissions/data"
)

var _ Factory = (*SQLFactory)(nil)

func NewFactory(db *pgxpool.Pool, cacheFactory factory.CacheFactory) *SQLFactory {
	return &SQLFactory{
		db:           db,
		cacheFactory: cacheFactory,
	}
}

//go:generate go run go.uber.org/mock/mockgen -source=factory.go -destination=mocks/mock_factory.go -package=mocks
type Factory interface {
	NewRoleReader() roles.Reader
	NewRoleWriter() roles.Writer
	NewRoleBindingsReader() rolebindings.Reader
	NewRoleBindingsWriter() rolebindings.Writer
	NewUserReader() users.Reader
	NewUserWriter() users.Writer
	NewResourceReader() resources.Reader
	NewResourceWriter() resources.Writer
	NewPermissionReader() permissionsv1.Reader
	NewPermissionWriter() permissionsv1.Writer
	NewPermissionBindingReader() permissionbindingsv1.Reader
	NewPermissionBindingWriter() permissionbindingsv1.Writer
	NewAggReader() aggsv1.Reader
}

type SQLFactory struct {
	db           *pgxpool.Pool
	cacheFactory factory.CacheFactory
}

// func (f *SQLFactory) NewPermissionReader() permdata.Reader {
// return permdata.NewSQLReader(f.cacheFactory.NewCache("permissions", time.Hour*24), f.db)
// }

// func (f *SQLFactory) NewPermissionWriter() permdata.Writer {
// return permdata.NewSQLWriter(f.db)
// }

func (f *SQLFactory) NewRoleBindingsReader() rolebindings.Reader {
	return rolebindings.NewSQLReader(f.cacheFactory.NewCache("rolebindings", time.Second*15), rolebindingsdatagen.New(f.db))
}

func (f *SQLFactory) NewRoleBindingsWriter() rolebindings.Writer {
	return rolebindings.NewSQLWriter(rolebindingsdatagen.New(f.db))
}

func (f *SQLFactory) NewRoleReader() roles.Reader {
	cache := f.cacheFactory.NewCache("roles", time.Hour*24)
	queries := rolesdatagen.New(f.db)
	return roles.NewSQLRoleReader(cache, queries)
}

func (f *SQLFactory) NewUserReader() users.Reader {
	cache := f.cacheFactory.NewCache("users", time.Hour*24)
	queries := usersdatagen.New(f.db)
	return users.NewSQLReader(cache, queries)
}

func (f *SQLFactory) NewUserWriter() users.Writer {
	queries := usersdatagen.New(f.db)
	return users.NewSQLWriter(queries)
}

// func (f *SQLFactory) NewUserAggReader() userdata.AggReader {
// 	return userdata.NewSQLUserAggReader(f.cacheFactory.NewCache("user_agg", time.Hour*24), f.db)
// }

// func (f *SQLFactory) NewRoleAggReader() roledata.AggReader {
// 	return roledata.NewSQLAggReader(f.cacheFactory.NewCache("role_agg", time.Hour*24), f.db)
// }

func (f *SQLFactory) NewResourceReader() resources.Reader {
	cache := f.cacheFactory.NewCache("resources", time.Hour*24)
	queries := resourcesdatagen.New(f.db)
	return resources.NewSQLReader(&cache, queries)
}

func (f *SQLFactory) NewResourceWriter() resources.Writer {
	queries := resourcesdatagen.New(f.db)
	return resources.NewSQLWriter(queries)
}

func (f *SQLFactory) NewRoleWriter() rolesv1.Writer {
	queries := rolesdatagen.New(f.db)
	return rolesv1.NewSQLWriter(queries)
}

func (f *SQLFactory) NewPermissionReader() permissionsv1.Reader {
	cache := f.cacheFactory.NewCache("permissions", time.Hour*24)
	queries := permissionsdatagen.New(f.db)
	return permissionsv1.NewSQLReader(&cache, queries)
}

func (f *SQLFactory) NewPermissionWriter() permissionsv1.Writer {
	queries := permissionsdatagen.New(f.db)
	return permissionsv1.NewSQLWriter(queries)
}

func (f *SQLFactory) NewPermissionBindingReader() permissionbindingsv1.Reader {
	cache := f.cacheFactory.NewCache("permission_bindings", time.Hour*24)
	queries := permissionbindingsdatagen.New(f.db)
	return permissionbindingsv1.NewSQLReader(&cache, *queries)
}

func (f *SQLFactory) NewPermissionBindingWriter() permissionbindingsv1.Writer {
	queries := permissionbindingsdatagen.New(f.db)
	return permissionbindingsv1.NewSQLWriter(*queries)
}

func (f *SQLFactory) NewAggReader() aggsv1.Reader {
	cache := f.cacheFactory.NewCache("aggs_reader", time.Hour*24)
	queries := aggsdatagen.New(f.db)
	return aggsv1.NewSQLReader(queries, &cache)
}
