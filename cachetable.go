package levonCache

import (
	"log"
	"sync"
	"time"
)

type CacheTable struct {
	sync.RWMutex

	name  string
	items map[interface{}]*CacheItem

	cleanupTimer    *time.Timer
	cleanupInterval time.Duration
	logger          *log.Logger

	loadData func(key interface{}, args ...interface{}) *CacheItem
	addItem  func(item *CacheItem)
	delItem  func(item *CacheItem)
}

func (table *CacheTable) Count() int {
	table.RLock()
	defer table.RUnlock()
	return len(table.items)
}

func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.RUnlock()

	for k, v : range table.items {
		trans(k, v)
	}
}


func (table *CacheTable) SetLoadData(f func(interface{}, ...interface{}) *CacheItem){
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}
