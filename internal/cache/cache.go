package cache

import (
	"container/list"
	"sync"
	"sync/atomic"

	"wbl0/internal/model"
)

type Cache struct {
	mu  sync.RWMutex
	cap int
	ll  *list.List
	m   map[string]*list.Element

	hits   atomic.Int64
	misses atomic.Int64
}

type entry struct {
	key string
	val model.Order
}

func New(capacity int) *Cache {
	if capacity <= 0 {
		capacity = 10
	}
	return &Cache{
		cap: capacity,
		ll:  list.New(),
		m:   make(map[string]*list.Element),
	}
}

func (c *Cache) Get(id string) (model.Order, bool) {
	c.mu.RLock()
	ele := c.m[id]
	defer c.mu.RUnlock()
	if ele == nil {
		c.misses.Add(1)
		return model.Order{}, false
	}
	//c.ll.MoveToFront(ele)  типа чтобы актульное всегда было первым надо ли оставить?
	val := ele.Value.(entry).val
	c.hits.Add(1)
	return val, true
}

func (c *Cache) Set(order model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele := c.m[order.OrderUID]; ele != nil {
		en := ele.Value.(entry)
		ele.Value = entry{key: en.key, val: order}
		c.ll.MoveToFront(ele)
		return
	}

	ele := c.ll.PushFront(entry{key: order.OrderUID, val: order})
	c.m[order.OrderUID] = ele

	if c.ll.Len() > c.cap {
		tail := c.ll.Back()
		if tail != nil {
			en := tail.Value.(entry)
			delete(c.m, en.key)
			c.ll.Remove(tail)
		}
	}
}

func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ll.Len()
}

func (c *Cache) Capacity() int { return c.cap }
func (c *Cache) Stats() (hits, misses int64) {
	return c.hits.Load(), c.misses.Load()
}
