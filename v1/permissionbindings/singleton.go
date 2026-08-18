package permissionbindings

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
)

var permissionBindingsService Service

var m sync.Mutex

func GetPermissionBindingsService() Service {
	m.Lock()
	defer m.Unlock()
	if permissionBindingsService == nil {
		panic("permission bindings service not initialized")
	}
	return permissionBindingsService
}

func InitPermissionBindingsService(factory datav1.Factory) {
	m.Lock()
	defer m.Unlock()
	permissionBindingsService = NewServiceImpl(factory)
}
