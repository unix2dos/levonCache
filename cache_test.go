package levonCache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	k = "testkey"
	v = "testvalue"
)

func TestCache(t *testing.T) {
	table := Cache("testCache")
	table.Add(k+"_1", v+"_1", 0*time.Second)
	table.Add(k+"_2", v+"_2", 0*time.Second)

	p1, _ := table.Value(k + "_1")
	p2, _ := table.Value(k + "_2")

	t.Log(p1.Data(), p2.Data())
}

func TestCacheExpire(t *testing.T) {

	table := Cache("testCache")
	table.Add(k+"_1", v+"_1", 200*time.Millisecond)
	table.Add(k+"_2", v+"_2", 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)
	p1, err1 := table.Value(k + "_1")
	_, err2 := table.Value(k + "_2")

	t.Log(p1.Data(), err1)
	t.Log(err2)
}

func TestExists(t *testing.T) {
	table := Cache("testExist")
	table.Add(k, v, 0)
	if table.Exists(k) {
		t.Log("key exist")
	}
}

func TestNotFoundAdd(t *testing.T) {
	table := Cache("NotFound")
	t.Log(table.NotFoundAdd(k, v, 0))
	t.Log(table.NotFoundAdd(k, v, 0))
	t.Log(table.NotFoundAdd(k, v, 0))
}

func TestNotFoundAndConcurrency(t *testing.T) {
	table := Cache("Concurrency")

	var finish sync.WaitGroup
	var added int32
	var idle int32

	fn := func(id int) {
		for i := 0; i < 100; i++ {
			if table.NotFoundAdd(i, i+id, 0) {
				atomic.AddInt32(&added, 1)
			} else {
				atomic.AddInt32(&idle, 1)
			}
			time.Sleep(0)
		}
		finish.Done()
	}

	finish.Add(10)
	go fn(0000)
	go fn(1000)
	go fn(2000)
	go fn(3000)
	go fn(4000)
	go fn(5000)
	go fn(6000)
	go fn(7000)
	go fn(8000)
	go fn(9000)
	finish.Wait()

	t.Log(added, idle)

	table.Foreach(func(key interface{}, item *CacheItem) {
		k := key.(int)
		v := item.Data().(int)
		t.Log(k, v)
	})
}
