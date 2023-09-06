# go-cache

A thread safe in-memory key-value cache suitable for single instance microservices. Any object can be stored with a given duration or no expiration time at all. Since the cache is thread safe, it can be safely used by multiple goroutines.

# Installation

    go get github.com/J4NN0/go-cache

# Usage

```go
package main

import (
	"fmt"
	"time"

	cache "github.com/J4NN0/go-cache"
)

func main() {
	// Create a cache with a default expiration time of 10 minutes, and which
	// purges expired items every 1 second
	c := cache.NewCache(10*time.Minute, 1*time.Second)
	defer c.Stop()

	// Set a new entry with key "foo" and value "someValue", with the default 
	// expiration time prior set (i.e. 10 minutes)
	c.Set("foo", "someValue", cache.DefaultExpiration)

	// Add a new entry - if an item doesn't already exist for the given key -  
	// with key "bar" and value "1", with no expiration time
	err := c.Add("bar", 1, cache.NoExpiration)
	if err != nil {
		fmt.Printf("Could not add 'bar': %v\n", err)
		return
    }
	
	// Replace an existing entry only if it hasn't expired yet
	err = c.Replace("foo", "someValue2", cache.DefaultExpiration)
	if err != nil {
		fmt.Printf("Could not replace 'foo': %v\n", err)
		return
	}

	// Since Go is statically typed, and cache values can be anything, type
	// assertion might be needed in some case
	var foo string
	x, found := c.Get("foo")
	if !found {
		fmt.Printf("Could not find 'foo'\n")
		return
	}
	foo = x.(string)
	fmt.Printf("Got 'foo': %s\n", foo) // someValue2

	// Current cache size can be checked with following method
	ic := c.ItemCount()
	fmt.Printf("Current cache size: %d\n", ic) // 2

	// Entry can be deleted before it will expire
	// (i.e. 10 minutes after being set in this case)
	c.Delete("foo")

	// Delete entire cache content
	c.Flush()
}
```
