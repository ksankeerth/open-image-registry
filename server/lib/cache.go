package lib

import (
	"sync"
	"time"
)

type item struct {
	// for the moment, We'll consider values as string
	// if requirement arises, we may use generics
	value      string
	expiration int64
}

type Cache struct {
	items map[string]*item
	mu    sync.RWMutex
	stop  chan struct{}
}

func NewCache(cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*item),
		stop:  make(chan struct{}),
	}

	go c.startCleanup(cleanupInterval)

	return c
}

func (c *Cache) Set(key string, value string, ttl time.Duration) {
	expiration := time.Now().Add(ttl).UnixNano()
	
	c.mu.Lock()
	c.items[key] = &item{
		value: value,
		expiration: expiration,
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()
	if !found {
		return "" 
	}

	if time.Now().UnixNano() > item.expiration {
		c.Delete(key)
		return ""
	}

	return item.value
}


func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) Count() int {
	c.mu.RLock()
	len := len(c.items)
	c.mu.RUnlock()
	return len
}

func (c *Cache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]*item)
	c.mu.Unlock()
}

func (c *Cache) Stop() {
	close(c.stop)
}

func (c *Cache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache) cleanup() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	for key, item := range c.items {
		if now > item.expiration {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}