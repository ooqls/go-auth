package roles

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
)

var roleService Service

var m sync.Mutex

func GetRoleService() Service {
	m.Lock()
	defer m.Unlock()
	if roleService == nil {
		panic("role service not initialized")
	}
	return roleService
}

func InitRoleService(factory datav1.Factory) {
	m.Lock()
	defer m.Unlock()
	roleService = NewServiceImpl(factory)
}
