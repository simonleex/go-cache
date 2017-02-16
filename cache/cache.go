package cache

import (
	"sync"
	"time"
	"fmt"
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
	mu                sync.RWMutex
}

const (
	DefaultExpiration time.Duration = 0
	NoExpiration time.Duration = -1
)


// check if item has expired
func (item *Item) hasExpired() bool{
	return item.Expiration > 0 && item.Expiration < time.Now().UnixNano()
}


// get expiration timestamp by time.Duration
func calcExpirationTime(defaultDuration, duration time.Duration) int64 {
	if duration > 0 {
		return time.Now().Add(duration).UnixNano()
	}else if duration == DefaultExpiration {
		return defaultDuration
	}
	// here duration must < 0 indicate that there is no expiration
	return duration
}

// set or replacing a existing item
func (c *cache) Set(k string, v interface{}, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(k, v, d)
}

// add if there isn't an existing item or unexpired item by given k
// otherwise an error return
func (c *cache) Add(k string, v interface{}, d time.Duration) error{
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[k]
	if found && item.hasExpired{
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
		Object: v,
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

	return item, true
}

// delete all items
func (c *cache)Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = map[string]Item{}
}

func (c *cache)ItemCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

func NewCache(defaultExpiration time.Duration) *Cache {
	items := make(map[string]Item)
	return &cache{
		items: items,
		defaultExpiration: defaultExpiration,
	}
}
