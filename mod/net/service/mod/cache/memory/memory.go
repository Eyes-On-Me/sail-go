package cache

import (
	"errors"
	"github.com/sail-services/sail-go/foundation/framework/service/cache"
	"sync"
	"time"
)

type MemoryItem struct {
	val     interface{}
	created int64
	expire  int64
}

type MemoryCacher struct {
	lock     sync.RWMutex
	items    map[string]*MemoryItem
	interval int
}

func init() {
	cache.Register("memory", NewMemoryCacher())
}

func NewMemoryCacher() *MemoryCacher {
	return &MemoryCacher{items: make(map[string]*MemoryItem)}
}

func (c *MemoryCacher) Put(key string, val interface{}, expire int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items[key] = &MemoryItem{
		val:     val,
		created: time.Now().Unix(),
		expire:  expire,
	}
	return nil
}

func (c *MemoryCacher) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return nil
	}
	if item.hasExpired() {
		go c.Delete(key)
		return nil
	}
	return item.val
}

func (c *MemoryCacher) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.items, key)
	return nil
}

func (c *MemoryCacher) Incr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}
	item.val, err = cache.Incr(item.val)
	return err
}

func (c *MemoryCacher) Decr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}
	item.val, err = cache.Decr(item.val)
	return err
}

func (c *MemoryCacher) IsExist(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.items[key]
	return ok
}

func (c *MemoryCacher) Flush() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items = make(map[string]*MemoryItem)
	return nil
}

func (c *MemoryCacher) checkRawExpiration(key string) {
	item, ok := c.items[key]
	if !ok {
		return
	}
	if item.hasExpired() {
		delete(c.items, key)
	}
}

func (c *MemoryCacher) checkExpiration(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.checkRawExpiration()
}

func (c *MemoryCacher) startGC() {
	if c.interval < 1 {
		return
	}
	if c.items != nil {
		c.lock.Lock()
		defer c.lock.Unlock()
		for key, _ := range c.items {
			c.checkRawExpiration(key)
		}
	}
	time.AfterFunc(time.Duration(c.interval)*time.Second, func() { c.startGC() })
}

func (c *MemoryCacher) StartAndGC(opt cache.Options) error {
	c.interval = opt.Interval
	go c.startGC()
	return nil
}

func (item *MemoryItem) hasExpired() bool {
	return item.expire > 0 &&
		(time.Now().Unix()-item.created) >= item.expire
}
