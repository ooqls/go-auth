package rolebindings

import "sync"

var roleBindingsService Service
var m sync.Mutex

func GetRoleBindingsService() Service {
	m.Lock()
	defer m.Unlock()
	return roleBindingsService
}

func SetRoleBindingsService(s Service) {
	m.Lock()
	defer m.Unlock()
	roleBindingsService = s
}
