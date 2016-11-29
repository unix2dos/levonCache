package levonCache

import (
	"log"
	"sort"
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
			table.Del(key)
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
	table.items[key] = &item

	expDur := table.cleanupInterval
	addItem := table.addItem

	table.Unlock()

	if addItem != nil {
		addItem(&item)
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

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := CreateCacheItem(key, data, lifeSpan)
	table.items[key] = &item

	expDur := table.cleanupInterval
	addItem := table.addItem
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

	if ok {
		r.KeepAlive()
		return r, nil
	}

	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.data, item.lifeSpan)
			return item, nil
		}
		return nil, ErrKeyNotFoundOrLoadTable
	}

	return nil, ErrKeyNotFound
}

func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.items = make(map[interface{}]*CacheItem)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

func (table *CacheTable) log(v ...interface{}) {

	if table.logger == nil {
		return
	}

	table.logger.Print(v)
}

//--------------------------------------
type CacheItemPair struct {
	Key           interface{}
	AccessedCount int64
}
type CacheItemPairArray []CacheItemPair

func (p CacheItemPairArray) Len() int {
	return len(p)
}
func (p CacheItemPairArray) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p CacheItemPairArray) Less(i, j int) bool {
	return p[i].AccessedCount > p[j].AccessedCount
}

//获取访问次数最多的前几名
func (table *CacheTable) MostAccessed(count int64) []*CacheItem {
	table.RLock()
	defer table.RUnlock()

	p := make(CacheItemPairArray, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{k, v.accessCount}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
	c := int64(0)
	for _, v := range p {
		if c >= count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}
