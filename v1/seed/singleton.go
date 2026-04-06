package seed

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/v1/rolebindings"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/ooqls/go-auth/v1/users"
)

var seedService Service

var m sync.Mutex

func GetSeedService() Service {
	m.Lock()
	defer m.Unlock()
	if seedService == nil {
		panic("seed service not initialized")
	}
	return seedService
}

func InitSeedService(factory datav1.Factory) {
	m.Lock()
	defer m.Unlock()
	seedService = NewServiceImpl(
		roles.GetRoleService(),
		rolebindings.GetRoleBindingsService(),
		users.GetUserService(),
	)
}
