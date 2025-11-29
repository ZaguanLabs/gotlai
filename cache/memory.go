package cache

import (
	"sync"
	"time"
)

// cacheEntry holds a cached value with its timestamp.
type cacheEntry struct {
	value     string
	timestamp time.Time
}

// InMemoryCache is a thread-safe in-memory cache with TTL support.
type InMemoryCache struct {
	cache map[string]cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewInMemoryCache creates a new in-memory cache with the specified TTL.
// If ttlSeconds is 0 or negative, entries never expire.
func NewInMemoryCache(ttlSeconds int) *InMemoryCache {
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttlSeconds <= 0 {
		ttl = 0 // No expiration
	}
	return &InMemoryCache{
		cache: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired, empty string and false otherwise.
func (c *InMemoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if !ok {
		return "", false
	}

	// Check TTL if enabled
	if c.ttl > 0 && time.Since(entry.timestamp) > c.ttl {
		// Entry expired - clean it up
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		return "", false
	}

	return entry.value, true
}

// Set stores a value in the cache.
func (c *InMemoryCache) Set(key string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
	}
	return nil
}

// Len returns the number of entries in the cache (including expired ones).
func (c *InMemoryCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Clear removes all entries from the cache.
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]cacheEntry)
}

// Entries returns all non-expired entries as key-value pairs.
// This is used for cache export.
func (c *InMemoryCache) Entries() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string)
	now := time.Now()

	for key, entry := range c.cache {
		// Skip expired entries
		if c.ttl > 0 && now.Sub(entry.timestamp) > c.ttl {
			continue
		}
		result[key] = entry.value
	}

	return result
}
