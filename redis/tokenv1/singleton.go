package tokenv1

import "sync"

var m sync.Mutex
var c TokenCache

func Get() TokenCache {
	m.Lock()
	defer m.Unlock()

	return c
}

func Set(newC TokenCache) {
	m.Lock()
	defer m.Unlock()

	c = newC
}
