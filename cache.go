package go_cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrItemAlreadyExists = errors.New("item already exists")
	ErrItemNotFound      = errors.New("item not found")
)

const (
	// DefaultExpiration For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to NewCache (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
	// NoExpiration For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
)

type Cache struct {
	stop chan struct{}
	wg   sync.WaitGroup

	mu                sync.RWMutex
	items             map[string]item
	defaultExpiration time.Duration
}

type item struct {
	object     any
	expiration int64
}

// NewCache Returns a new cache with a given default expiration duration and cleanup interval.
// If the expiration duration is less than 1, the items in the cache never expire (by default),
// and must be deleted manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling DeleteExpired().
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

// cleanUp Deletes all expired items from the cache. This can be used if the
// cleanupInterval passed to NewCache() is set to less than 1.
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

// Stop This will stop the cleanup goroutine and free up resources.
func (c *Cache) Stop() {
	close(c.stop)
	c.wg.Wait()
}

// Set Adds an item to the cache, replacing any existing item.
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
// If the duration is positive, the item expires after that time has passed.
func (c *Cache) Set(key string, object any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(key, object, duration)
}

// Add Inserts an item to the cache only if an item doesn't already exist for the given key,
// or if the existing item has expired. Returns ErrItemAlreadyExists error otherwise.
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
// If the duration is positive, the item expires after that time has passed.
func (c *Cache) Add(key string, object any, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	isExpired := item.expiration > 0 && item.expiration <= time.Now().UnixNano()
	if found && !isExpired {
		return fmt.Errorf("%w: %s", ErrItemAlreadyExists, key)
	}
	c.set(key, object, duration)

	return nil
}

// Replace Sets a new value for the cache only if the given key already exists,
// and the existing item has not expired. Returns ErrItemNotFound error otherwise.
// If the duration is 0 (DefaultExpiration), the cache's default expiration time is used.
// If it is -1 (NoExpiration), the item never expires.
// If the duration is positive, the item expires after that time has passed.
func (c *Cache) Replace(key string, object any, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	isExpired := item.expiration > 0 && item.expiration <= time.Now().UnixNano()
	if !found || isExpired {
		return fmt.Errorf("%w: %s", ErrItemNotFound, key)
	}
	c.set(key, object, duration)

	return nil
}

func (c *Cache) set(key string, object any, duration time.Duration) {
	var expiration int64
	if duration == DefaultExpiration {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.items[key] = item{
		object:     object,
		expiration: expiration,
	}
}

// Get Looks up a key's value from the cache.
// If the key corresponds to an item in the cache, a copy of the value is returned.
// If the key does not exist, nil is returned.
// If the key is found but has expired, it is deleted from the cache and nil is returned.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	isExpired := item.expiration > 0 && item.expiration <= time.Now().UnixNano()
	if !found || isExpired {
		return nil, false
	}

	return item.object, true
}

// Delete Removes the provided key from the cache.
// If the key was not found, Delete is a no-op.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Flush Completely clears the cache.
// This will delete all items in the cache, including ones that have not yet expired.
// This is a no-op if the cache is already empty.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = map[string]item{}
}

// ItemCount Returns the number of items in the cache. This may include items that have expired,
// but have not yet been cleaned up.
func (c *Cache) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
