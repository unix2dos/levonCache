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

	for k, v := range table.items {
		trans(k, v)
	}
}

func (table *CacheTable) SetLoadData(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

func (table *CacheTable) setAddCallBack(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.addItem = f
}

func (table *CacheTable) setDelCallBack(f func(*CacheItem)) {
	table.Lock()
	defer table.Unlock()
	table.delItem = f
}

func (table *CacheTable) setLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

func (table *CacheTable) expirationCheck() {
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log(table.cleanupInterval)
	} else {
		table.log(table.name)
	}
	items := table.items
	table.Unlock()

	now := time.Now()
	smallestDuration := 0 * time.Second //最快的一个key期限的日期
	for key, item := range items {
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 { //0代表没有期限限制
			continue
		}

		if now.Sub(accessedOn) >= lifeSpan {
			table.Delete(key)
		} else {
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	table.Lock()
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck() //再来一次检查
		})
	}
	table.Unlock()
}

func (table *CacheTable) Add(key interface{}, data interface{}, lifeSpan time.Duration) *CacheItem {
	item := CreateCacheItem(key, data, lifeSpan)
	table.Lock()
	table.items[key] = &data

	expDur := table.cleanupInterval
	addItem := table.addItem

	table.Unlock()

	if addItem != nil {
		addItem(&key)
	}

	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.expirationCheck()
	}

	return &item
}

func (table *CacheTable) Del(key interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	if !ok {
		table.RUnlock()
		return nil, ErrKeyNotFound
	}
	delItem := table.delItem
	table.RUnlock()

	if delItem != nil {
		delItem(r)
	}

	r.RLock()
	defer r.RUnlock()
	if r.expireCallBack != nil {
		r.expireCallBack(key)
	}

	table.Lock()
	defer table.Unlock()
	delete(table.items, key)

	return r, nil
}

func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]
	return ok
}

func (table *CacheTable) NotFoundAdd(key interface{}, data interface{}, lifeSpan time.Duration) bool {
	table.Lock()

	if r, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := CreateCacheItem(key, data, lifeSpan)
	table.items[key] = &data

	expDur = table.cleanupInterval
	addItem = table.addItem
	table.Unlock()

	if addItem != nil {
		addItem(&item)
	}

	if lifeSpan > 0 && (expDur == 0 || lifeSpan < expDur) {
		table.expirationCheck()
	}
	return true
}

func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()
	/*
		if ok {
			r.KeepAlive()
			return r, nil
		}
			return nil,ErrKeyNotFound
	*/
}
