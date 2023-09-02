package go_cache

import (
	"errors"
	"sync"
	"time"
)

const (
	// DefaultExpiration For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to NewCache() to (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
	// NoExpiration For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
)

var ErrItemNotFound = errors.New("item not found")

type item struct {
	object     any
	expiration int64
}

type Cache struct {
	stop chan struct{}
	wg   sync.WaitGroup

	mu                sync.RWMutex
	items             map[string]item
	defaultExpiration time.Duration
}

// NewCache Return a new cache instance with the provided default expiration time and cleanup interval.
// If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache.
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	if defaultExpiration <= 0 {
		defaultExpiration = NoExpiration
	}

	c := &Cache{
		stop:              make(chan struct{}),
		mu:                sync.RWMutex{},
		items:             make(map[string]item),
		defaultExpiration: defaultExpiration,
	}

	if cleanupInterval > 0 {
		c.wg.Add(1)
		go func(cleanupInterval time.Duration) {
			defer c.wg.Done()
			c.cleanUp(cleanupInterval)
		}(cleanupInterval)
	}

	return c
}

func (c *Cache) cleanUp(cleanupInterval time.Duration) {
	t := time.NewTicker(cleanupInterval)
	defer t.Stop()

	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			c.mu.Lock()
			for key, object := range c.items {
				if object.expiration > 0 && object.expiration <= time.Now().UnixNano() {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Cache) Stop() {
	close(c.stop)
	c.wg.Wait()
}

func (c *Cache) Set(key string, object any, duration time.Duration) {
	var expiration int64
	if duration == DefaultExpiration {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		object:     object,
		expiration: expiration,
	}
}

func (c *Cache) Get(key string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, ErrItemNotFound
	}

	return item.object, nil
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = map[string]item{}
}

func (c *Cache) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
