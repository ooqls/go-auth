package usercachev1

import "sync"

var m sync.Mutex = sync.Mutex{}
var c UserNameCache

func Get() UserNameCache {
	m.Lock()
	defer m.Unlock()

	return c
}

func Set(newC UserNameCache) {
	m.Lock()
	defer m.Unlock()

	c = newC
}