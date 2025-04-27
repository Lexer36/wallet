package cache

import (
	"context"
)

type LRUCache interface {
	Add(key interface{}, value interface{}) (evicted bool)
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{}) (present bool)
}

// todo now we lock mutex for all our wallets but they should not affect each other
type Cache struct {
	cache LRUCache
}

func New(c LRUCache) *Cache {
	return &Cache{
		cache: c,
	}
}

func (c *Cache) Get(_ context.Context, key string) (int64, bool) {
	val, ok := c.cache.Get(key)
	if ok {
		return val.(int64), true
	}

	return 0, false
}

func (c *Cache) Set(_ context.Context, key string, amount int64) {
	c.cache.Add(key, amount)
}

func (c *Cache) Delete(_ context.Context, key string) {
	c.cache.Remove(key)
}
