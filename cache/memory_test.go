package cache

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryCache_GetSet(t *testing.T) {
	c := NewInMemoryCache(3600) // 1 hour TTL

	// Test set and get
	err := c.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, ok := c.Get("key1")
	if !ok {
		t.Error("Get should return true for existing key")
	}
	if val != "value1" {
		t.Errorf("Get returned %q, want %q", val, "value1")
	}

	// Test missing key
	val, ok = c.Get("nonexistent")
	if ok {
		t.Error("Get should return false for missing key")
	}
	if val != "" {
		t.Errorf("Get should return empty string for missing key, got %q", val)
	}
}

func TestInMemoryCache_TTL(t *testing.T) {
	c := NewInMemoryCache(1) // 1 second TTL

	c.Set("key1", "value1")

	// Should be available immediately
	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Error("Value should be available immediately after set")
	}

	// Wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// Should be expired now
	val, ok = c.Get("key1")
	if ok {
		t.Error("Value should be expired after TTL")
	}
	if val != "" {
		t.Errorf("Expired value should return empty string, got %q", val)
	}
}

func TestInMemoryCache_NoTTL(t *testing.T) {
	c := NewInMemoryCache(0) // No TTL

	c.Set("key1", "value1")

	// Should be available
	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Error("Value should be available with no TTL")
	}
}

func TestInMemoryCache_Overwrite(t *testing.T) {
	c := NewInMemoryCache(3600)

	c.Set("key1", "value1")
	c.Set("key1", "value2")

	val, ok := c.Get("key1")
	if !ok {
		t.Error("Key should exist")
	}
	if val != "value2" {
		t.Errorf("Value should be overwritten, got %q, want %q", val, "value2")
	}
}

func TestInMemoryCache_Len(t *testing.T) {
	c := NewInMemoryCache(3600)

	if c.Len() != 0 {
		t.Errorf("Empty cache should have length 0, got %d", c.Len())
	}

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if c.Len() != 2 {
		t.Errorf("Cache should have length 2, got %d", c.Len())
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	c := NewInMemoryCache(3600)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("Cleared cache should have length 0, got %d", c.Len())
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("Cleared cache should not contain any keys")
	}
}

func TestInMemoryCache_Concurrent(t *testing.T) {
	c := NewInMemoryCache(3600)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i%26))
			c.Set(key, "value")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i%26))
			c.Get(key)
		}(i)
	}

	wg.Wait()
	// If we get here without a race condition, the test passes
}

// Verify InMemoryCache implements TranslationCache
var _ TranslationCache = (*InMemoryCache)(nil)
