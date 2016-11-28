package levonCache

import "sync"

var (
	mutex sync.RWMutex
	cache = make(map[string]*CacheTable)
)

func Cache(table string) *CacheTable {
	mutex.RLock()
	t, ok = cache[table]
	mutex.RUnlock()

	if !ok {
		t = &CacheTable{
			name:  table,
			items: make(map[interface{}]*CacheItem),
		}
		mutex.Lock()
		cache[table] = t
		mutex.Unlock()
	}

	return t
}
