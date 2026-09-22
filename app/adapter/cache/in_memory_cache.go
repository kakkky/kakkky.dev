package cache

import (
	"sync"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type inMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]inMemoryEntry
}

type inMemoryEntry struct {
	value any
	exp   time.Time
}

func (c *Cache) NewInMemoryCacheClient() domain.CacheClient {
	return &inMemoryCache{entries: map[string]inMemoryEntry{}}
}

func (c *inMemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.exp) {
		return nil, false
	}
	return e.value, true
}

func (c *inMemoryCache) GetStale(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return e.value, true
}

func (c *inMemoryCache) Set(key string, value any, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = inMemoryEntry{value: value, exp: expiresAt}
}
