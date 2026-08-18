package seed

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
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
	seedService = NewServiceImpl(factory)
}
