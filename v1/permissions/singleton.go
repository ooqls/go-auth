package permissions

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
)

var permissionsService Service

var m sync.Mutex

func GetPermissionsService() Service {
	m.Lock()
	defer m.Unlock()
	if permissionsService == nil {
		panic("permissions service not initialized")
	}
	return permissionsService
}

func InitPermissionsService(factory datav1.Factory) {
	m.Lock()
	defer m.Unlock()
	permissionsService = NewServiceImpl(factory)
}
