package cache

import (
	"sync"
	"time"
)

type Cache struct {
	Entries map[string]cacheEntry
	mu      sync.RWMutex
}

type cacheEntry struct {
	createdAt time.Time
	Val       []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		Entries: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Entries[key] = cacheEntry{
		createdAt: time.Now(),
		Val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.Entries[key]
	if !ok {
		return nil, false
	}
	return entry.Val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		for key, value := range c.Entries {
			if time.Since(value.createdAt) > interval {
				c.mu.Lock()
				delete(c.Entries, key)
				c.mu.Unlock()
			}
		}
	}
}
