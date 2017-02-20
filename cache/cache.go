package cache

import (
	"fmt"
	"sync"
	"time"
	"runtime"
)

type Cache struct {
	*cache
}

type Item struct {
	Object     interface{}
	Expiration int64
}

// need mutex because map isn't implemented thread safe
type cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	janitor           *janitor
	mu                sync.RWMutex
}


var h sync.RWMutex
const (
	DefaultExpiration time.Duration = 0
	NoExpiration      time.Duration = -1
)

// check if item has expired
func (item *Item) hasExpired() bool {
	return item.Expiration > 0 && item.Expiration < time.Now().UnixNano()
}

// get expiration timestamp by time.Duration
func calcExpirationTime(defaultDuration, duration time.Duration) int64 {
	if duration > 0 {
		return time.Now().Add(duration).UnixNano()
	} else if duration == DefaultExpiration && defaultDuration > 0{
		return time.Now().Add(defaultDuration).UnixNano()
	}
	// here duration must < 0 indicate that there is no expiration
	return -1
}

// set or replacing a existing item
func (c *cache) Set(k string, v interface{}, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(k, v, d)
}

// add if there isn't an existing item or unexpired item by given k
// otherwise an error return
func (c *cache) Add(k string, v interface{}, d time.Duration) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[k]
	if found && item.hasExpired() {
		return fmt.Errorf("there is an item existing by given k %s", k)
	}

	c.set(k, v, d)

	return nil
}

// replace if there is an existing item or unexpired item by given k,
// otherwise return an error
func (c *cache) Replace(k string, v interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[k]
	if !found || item.hasExpired() {
		return fmt.Errorf("can't find item by given k %s", k)
	}

	c.set(k, v, d)

	return nil
}

// set operation without lock
func (c *cache) set(k string, v interface{}, d time.Duration) {
	c.items[k] = Item{
		Object:     v,
		Expiration: calcExpirationTime(c.defaultExpiration, d),
	}
}

// return a item with the given k, and a bool indicate if the item exist
func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.hasExpired() {
		return nil, false
	}

	return item.Object, true
}

func (c *cache) DeleteExpired() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.Expiration > 0 && v.Expiration < now {
			delete(c.items, k)
		}
	}
}

func (c *cache) Delete(k string) error{
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[k]
	if !found || item.hasExpired(){
		return fmt.Errorf("can't find item %s:k", k)
	}
	delete(c.items, k)

	return nil
}

// delete all items
func (c *cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = map[string]Item{}
}

func (c *cache) ItemCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

type janitor struct {
	stopC    chan bool
	interval time.Duration
}

func (j *janitor) Run(c *cache) {
	j.stopC = make(chan bool)
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stopC:
			ticker.Stop()
			return
		}
	}
}


func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		interval: ci,
	}
	c.janitor = j

	go j.Run(c)
}

func stopJanitor(c *Cache) {
	c.janitor.stopC <- true
}

func newCacheWithJanitor(de, ci time.Duration, items map[string]Item) *Cache {
	if de == 0 {
		de = NoExpiration
	}
	c := &cache{
		defaultExpiration: de,
		items:             items,
	}

	C := &Cache{c}
	// to make sure after C gc'ed , the Janitor will stop
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}

	return C
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := map[string]Item{}
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}
