package users

import (
	"sync"

	"github.com/ooqls/go-auth/internal/datav1"
)

var userService Service

var m sync.Mutex

func GetUserService() Service {
	m.Lock()
	defer m.Unlock()
	if userService == nil {
		panic("user service not initialized")
	}
	return userService
}

func InitUserService(factory datav1.Factory) {
	m.Lock()
	defer m.Unlock()
	userService = NewServiceImpl(factory)
}
