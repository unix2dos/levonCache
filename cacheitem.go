package levonCache

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex

	key  interface{}
	data interface{}

	lifeSpan time.Duration

	createdOn   time.Time
	accessedOn  time.Time
	accessCount int64

	expireCallBack func(key interface{})
}

func CreateCacheItem(key interface{}, data interface{}, lifeSpan time.Duration) CacheItem {
	t := time.Now()
	return CacheItem{
		key:            key,
		data:           data,
		lifeSpan:       lifeSpan,
		createdOn:      t,
		accessedOn:     t,
		accessCount:    0,
		expireCallBack: nil,
	}
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

func (item *CacheItem) LifeSpan() time.Duration {
	item.RLock()
	defer item.RUnlock()
	return item.lifeSpan
}

func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

func (item *CacheItem) AccessCount() int64 {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *CacheItem) CreatedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.createdOn
}

func (item *CacheItem) Key() interface{} {
	return item.key //外面不能改动,所以不用加锁
}

func (item *CacheItem) Data() interface{} {
	return item.data
}

func (item *CacheItem) setExpireCallBack(f func(interface{})) {
	item.Lock()
	defer item.Unlock()
	item.expireCallBack = f
}
