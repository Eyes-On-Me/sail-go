package cache

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	_VERSION = "2015.03.07"
	_URL     = "https://github.com/pmylund/go-cache"
)

type Item struct {
	Object     interface{}
	Expiration *time.Time
}

func (item *Item) Expired() bool {
	if item.Expiration == nil {
		return false
	}
	return item.Expiration.Before(time.Now())
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Cache struct {
	*cache
}

type cache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	items             map[string]*Item
	janitor           *janitor
}

func (c *cache) Set(k string, x interface{}, d time.Duration) {
	c.Lock()
	c.set(k, x, d)
	c.Unlock()
}

func (c *cache) set(k string, x interface{}, d time.Duration) {
	var e *time.Time
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		t := time.Now().Add(d)
		e = &t
	}
	c.items[k] = &Item{
		Object:     x,
		Expiration: e,
	}
}

func (c *cache) Add(k string, x interface{}, d time.Duration) error {
	c.Lock()
	_, found := c.get(k)
	if found {
		c.Unlock()
		return fmt.Errorf("Item %s already exists", k)
	}
	c.set(k, x, d)
	c.Unlock()
	return nil
}

func (c *cache) Replace(k string, x interface{}, d time.Duration) error {
	c.Lock()
	_, found := c.get(k)
	if !found {
		c.Unlock()
		return fmt.Errorf("Item %s doesn't exist", k)
	}
	c.set(k, x, d)
	c.Unlock()
	return nil
}

func (c *cache) Get(k string) (interface{}, bool) {
	c.RLock()
	x, found := c.get(k)
	c.RUnlock()
	return x, found
}

func (c *cache) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found || item.Expired() {
		return nil, false
	}
	return item.Object, true
}

func (c *cache) Increment(k string, n int64) error {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) + int(n)
	case int8:
		v.Object = v.Object.(int8) + int8(n)
	case int16:
		v.Object = v.Object.(int16) + int16(n)
	case int32:
		v.Object = v.Object.(int32) + int32(n)
	case int64:
		v.Object = v.Object.(int64) + n
	case uint:
		v.Object = v.Object.(uint) + uint(n)
	case uintptr:
		v.Object = v.Object.(uintptr) + uintptr(n)
	case uint8:
		v.Object = v.Object.(uint8) + uint8(n)
	case uint16:
		v.Object = v.Object.(uint16) + uint16(n)
	case uint32:
		v.Object = v.Object.(uint32) + uint32(n)
	case uint64:
		v.Object = v.Object.(uint64) + uint64(n)
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + float64(n)
	default:
		c.Unlock()
		return fmt.Errorf("The value for %s is not an integer", k)
	}
	c.Unlock()
	return nil
}

func (c *cache) IncrementFloat(k string, n float64) error {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + n
	default:
		c.Unlock()
		return fmt.Errorf("The value for %s does not have type float32 or float64", k)
	}
	c.Unlock()
	return nil
}

func (c *cache) IncrementInt(k string, n int) (int, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementInt8(k string, n int8) (int8, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int8)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int8", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementInt16(k string, n int16) (int16, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int16)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int16", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementInt32(k string, n int32) (int32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int32", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementInt64(k string, n int64) (int64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int64", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUint(k string, n uint) (uint, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUintptr(k string, n uintptr) (uintptr, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uintptr)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uintptr", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUint8(k string, n uint8) (uint8, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint8)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint8", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUint16(k string, n uint16) (uint16, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint16)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint16", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUint32(k string, n uint32) (uint32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint32", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementUint64(k string, n uint64) (uint64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint64", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementFloat32(k string, n float32) (float32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(float32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an float32", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) IncrementFloat64(k string, n float64) (float64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(float64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an float64", k)
	}
	nv := rv + n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) Decrement(k string, n int64) error {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return fmt.Errorf("Item not found")
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) - int(n)
	case int8:
		v.Object = v.Object.(int8) - int8(n)
	case int16:
		v.Object = v.Object.(int16) - int16(n)
	case int32:
		v.Object = v.Object.(int32) - int32(n)
	case int64:
		v.Object = v.Object.(int64) - n
	case uint:
		v.Object = v.Object.(uint) - uint(n)
	case uintptr:
		v.Object = v.Object.(uintptr) - uintptr(n)
	case uint8:
		v.Object = v.Object.(uint8) - uint8(n)
	case uint16:
		v.Object = v.Object.(uint16) - uint16(n)
	case uint32:
		v.Object = v.Object.(uint32) - uint32(n)
	case uint64:
		v.Object = v.Object.(uint64) - uint64(n)
	case float32:
		v.Object = v.Object.(float32) - float32(n)
	case float64:
		v.Object = v.Object.(float64) - float64(n)
	default:
		c.Unlock()
		return fmt.Errorf("The value for %s is not an integer", k)
	}
	c.Unlock()
	return nil
}

func (c *cache) DecrementFloat(k string, n float64) error {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) - float32(n)
	case float64:
		v.Object = v.Object.(float64) - n
	default:
		c.Unlock()
		return fmt.Errorf("The value for %s does not have type float32 or float64", k)
	}
	c.Unlock()
	return nil
}

func (c *cache) DecrementInt(k string, n int) (int, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementInt8(k string, n int8) (int8, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int8)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int8", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementInt16(k string, n int16) (int16, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int16)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int16", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementInt32(k string, n int32) (int32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int32", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementInt64(k string, n int64) (int64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(int64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an int64", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUint(k string, n uint) (uint, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUintptr(k string, n uintptr) (uintptr, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uintptr)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uintptr", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUint8(k string, n uint8) (uint8, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint8)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint8", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUint16(k string, n uint16) (uint16, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint16)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint16", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUint32(k string, n uint32) (uint32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint32", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementUint64(k string, n uint64) (uint64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an uint64", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementFloat32(k string, n float32) (float32, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(float32)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an float32", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) DecrementFloat64(k string, n float64) (float64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.Unlock()
		return 0, fmt.Errorf("Item %s not found", k)
	}
	rv, ok := v.Object.(float64)
	if !ok {
		c.Unlock()
		return 0, fmt.Errorf("The value for %s is not an float64", k)
	}
	nv := rv - n
	v.Object = nv
	c.Unlock()
	return nv, nil
}

func (c *cache) Delete(k string) {
	c.Lock()
	c.delete(k)
	c.Unlock()
}

func (c *cache) delete(k string) {
	delete(c.items, k)
}

func (c *cache) DeleteExpired() {
	c.Lock()
	for k, v := range c.items {
		if v.Expired() {
			c.delete(k)
		}
	}
	c.Unlock()
}

func (c *cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	c.RLock()
	defer c.RUnlock()
	for _, v := range c.items {
		gob.Register(v.Object)
	}
	err = enc.Encode(&c.items)
	return
}

func (c *cache) SaveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = c.Save(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func (c *cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]*Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.Lock()
		defer c.Unlock()
		for k, v := range items {
			ov, found := c.items[k]
			if !found || ov.Expired() {
				c.items[k] = v
			}
		}
	}
	return err
}

func (c *cache) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = c.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func (c *cache) Items() map[string]*Item {
	c.RLock()
	defer c.RUnlock()
	return c.items
}

func (c *cache) ItemCount() int {
	c.RLock()
	n := len(c.items)
	c.RUnlock()
	return n
}

func (c *cache) Flush() {
	c.Lock()
	c.items = map[string]*Item{}
	c.Unlock()
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) Run(c *cache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			c.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
	}
	c.janitor = j
	go j.Run(c)
}

func newCache(de time.Duration, m map[string]*Item) *cache {
	if de == 0 {
		de = -1
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newCacheWithJanitor(de time.Duration, ci time.Duration, m map[string]*Item) *Cache {
	c := newCache(de, m)
	C := &Cache{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]*Item)
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

func NewFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]*Item) *Cache {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}
