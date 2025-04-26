package cache

import (
	"context"
	"sync"
)

// todo now we lock for all our wallets but they should not affect each other
type Cache struct {
	mu   sync.RWMutex
	data map[string]int64
}

func New() *Cache {
	return &Cache{
		data: make(map[string]int64),
	}
}

func (c *Cache) Get(_ context.Context, key string) (int64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) Set(_ context.Context, key string, amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = amount
}

func (c *Cache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
