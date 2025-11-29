package gotlai

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// slowCache simulates a slow cache for testing parallel lookups
type slowCache struct {
	data    map[string]string
	mu      sync.RWMutex
	delay   time.Duration
	lookups int64
}

func newSlowCache(delay time.Duration) *slowCache {
	return &slowCache{
		data:  make(map[string]string),
		delay: delay,
	}
}

func (c *slowCache) Get(key string) (string, bool) {
	atomic.AddInt64(&c.lookups, 1)
	time.Sleep(c.delay)
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *slowCache) Set(key string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func TestParallelCacheLookup_Basic(t *testing.T) {
	cache := newSlowCache(0)
	cache.Set("hash1:es_ES", "Hola")
	cache.Set("hash2:es_ES", "Mundo")

	nodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
		{Hash: "hash3", Text: "Missing"},
	}

	translations, misses := ParallelCacheLookup(cache, nodes, "es_ES")

	if len(translations) != 2 {
		t.Errorf("Expected 2 translations, got %d", len(translations))
	}

	if translations["hash1"] != "Hola" {
		t.Errorf("Expected 'Hola', got %q", translations["hash1"])
	}

	if len(misses) != 1 {
		t.Errorf("Expected 1 miss, got %d", len(misses))
	}

	if misses[0].Hash != "hash3" {
		t.Errorf("Expected miss hash 'hash3', got %q", misses[0].Hash)
	}
}

func TestParallelCacheLookup_Deduplication(t *testing.T) {
	cache := newSlowCache(0)

	// Same hash appears multiple times
	nodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash1", Text: "Hello"},
	}

	_, misses := ParallelCacheLookup(cache, nodes, "es_ES")

	// Should only have one miss (deduplicated)
	if len(misses) != 1 {
		t.Errorf("Expected 1 deduplicated miss, got %d", len(misses))
	}
}

func TestParallelCacheLookup_NilCache(t *testing.T) {
	nodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
	}

	translations, misses := ParallelCacheLookup(nil, nodes, "es_ES")

	if len(translations) != 0 {
		t.Errorf("Expected 0 translations with nil cache, got %d", len(translations))
	}

	if len(misses) != 1 {
		t.Errorf("Expected all nodes as misses with nil cache, got %d", len(misses))
	}
}

func TestParallelCacheLookup_EmptyNodes(t *testing.T) {
	cache := newSlowCache(0)
	translations, misses := ParallelCacheLookup(cache, []TextNode{}, "es_ES")

	if len(translations) != 0 {
		t.Errorf("Expected 0 translations for empty nodes, got %d", len(translations))
	}

	if len(misses) != 0 {
		t.Errorf("Expected 0 misses for empty nodes, got %d", len(misses))
	}
}

func TestParallelCacheLookup_FasterThanSequential(t *testing.T) {
	delay := 10 * time.Millisecond
	cache := newSlowCache(delay)

	// Pre-populate cache
	for i := 0; i < 10; i++ {
		cache.Set(CacheKey(string(rune('a'+i)), "es_ES"), "translated")
	}

	nodes := make([]TextNode, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = TextNode{Hash: string(rune('a' + i)), Text: "text"}
	}

	start := time.Now()
	ParallelCacheLookup(cache, nodes, "es_ES")
	elapsed := time.Since(start)

	// Sequential would take 10 * 10ms = 100ms
	// Parallel should be much faster (close to 10ms + overhead)
	maxExpected := 50 * time.Millisecond
	if elapsed > maxExpected {
		t.Errorf("Parallel lookup took %v, expected < %v", elapsed, maxExpected)
	}
}

func BenchmarkParallelCacheLookup(b *testing.B) {
	cache := newSlowCache(0)
	for i := 0; i < 100; i++ {
		cache.Set(CacheKey(string(rune(i)), "es_ES"), "translated")
	}

	nodes := make([]TextNode, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = TextNode{Hash: string(rune(i)), Text: "text"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParallelCacheLookup(cache, nodes, "es_ES")
	}
}
