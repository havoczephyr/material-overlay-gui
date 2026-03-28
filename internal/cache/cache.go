package cache

import (
	"sync"
	"time"

	"github.com/havoczephyr/material-overlay-gui/internal/card"
)

const (
	CardTTL = 7 * 24 * time.Hour // 7 days
	SetTTL  = 24 * time.Hour     // 24 hours
)

type entry struct {
	data      interface{}
	expiresAt time.Time
}

func (e *entry) expired() bool {
	return time.Now().After(e.expiresAt)
}

// Cache is an in-memory cache with TTL eviction.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
}

func New() *Cache {
	c := &Cache{
		entries: make(map[string]entry),
	}
	go c.evictLoop()
	return c
}

func (c *Cache) GetCard(name string) (*card.Card, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries["card:"+name]
	if !ok || e.expired() {
		return nil, false
	}

	cd, ok := e.data.(*card.Card)
	return cd, ok
}

func (c *Cache) SetCard(name string, cd *card.Card) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries["card:"+name] = entry{
		data:      cd,
		expiresAt: time.Now().Add(CardTTL),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || e.expired() {
		return nil, false
	}

	return e.data, true
}

func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

func (c *Cache) evictLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.evictExpired()
	}
}

func (c *Cache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
}
