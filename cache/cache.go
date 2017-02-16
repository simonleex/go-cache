package cache

import "sync"

type Cache struct {
	*cache
}

// need mu because map isn't implemented thread safe
type cache struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

// set or replacing a existing item
func (c *cache) Set(k string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = v
}

// return a item with the given k, and a bool indicate if the item exist
func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.items[k]; ok {
		return item, true
	} else {
		return nil, false
	}
}

func NewCache() *Cache {
	items := make(map[string]interface{})
	return &cache{
		items: items,
	}
}
