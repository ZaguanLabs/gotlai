package gotlai

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_TryAcquire(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         3,
	})

	// Should be able to acquire burst size immediately
	for i := 0; i < 3; i++ {
		if !limiter.TryAcquire() {
			t.Errorf("Expected to acquire token %d", i)
		}
	}

	// Fourth should fail
	if limiter.TryAcquire() {
		t.Error("Expected fourth acquire to fail")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 600, // 10 per second
		BurstSize:         1,
	})

	// Drain the bucket
	limiter.TryAcquire()

	// Should fail immediately
	if limiter.TryAcquire() {
		t.Error("Expected acquire to fail after drain")
	}

	// Wait for refill (100ms for 1 token at 10/sec)
	time.Sleep(150 * time.Millisecond)

	// Should succeed now
	if !limiter.TryAcquire() {
		t.Error("Expected acquire to succeed after refill")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 600, // 10 per second
		BurstSize:         1,
	})

	// Drain the bucket
	limiter.TryAcquire()

	// Wait should block then succeed
	ctx := context.Background()
	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}

	// Should have waited ~100ms
	if elapsed < 50*time.Millisecond {
		t.Errorf("Wait returned too quickly: %v", elapsed)
	}
}

func TestRateLimiter_WaitCancelled(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 1, // Very slow
		BurstSize:         1,
	})

	// Drain the bucket
	limiter.TryAcquire()

	// Cancel context quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Expected error when context cancelled")
	}
}

func TestRateLimiter_Available(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	})

	available := limiter.Available()
	if available != 5 {
		t.Errorf("Expected 5 available, got %f", available)
	}

	limiter.TryAcquire()
	limiter.TryAcquire()

	available = limiter.Available()
	if available < 2.9 || available > 3.1 {
		t.Errorf("Expected ~3 available, got %f", available)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	limiter := NewRateLimiter(RateLimitConfig{
		RequestsPerMinute: 6000, // 100 per second
		BurstSize:         10,
	})

	var wg sync.WaitGroup
	acquired := int64(0)
	var mu sync.Mutex

	// Launch 20 goroutines trying to acquire
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.TryAcquire() {
				mu.Lock()
				acquired++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have acquired exactly burst size
	if acquired != 10 {
		t.Errorf("Expected 10 acquired, got %d", acquired)
	}
}

func TestRateLimitedProvider(t *testing.T) {
	inner := &mockProviderForRateLimit{
		response: []string{"translated"},
	}

	provider := NewRateLimitedProvider(inner, RateLimitConfig{
		RequestsPerMinute: 600,
		BurstSize:         2,
	})

	ctx := context.Background()

	// First two should succeed immediately
	_, err := provider.Translate(ctx, TranslateRequest{Texts: []string{"a"}})
	if err != nil {
		t.Errorf("First translate failed: %v", err)
	}

	_, err = provider.Translate(ctx, TranslateRequest{Texts: []string{"b"}})
	if err != nil {
		t.Errorf("Second translate failed: %v", err)
	}

	// Third should wait for rate limit
	start := time.Now()
	_, err = provider.Translate(ctx, TranslateRequest{Texts: []string{"c"}})
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Third translate failed: %v", err)
	}

	// Should have waited
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected rate limit wait, but returned in %v", elapsed)
	}
}

func TestRateLimitedProvider_ContextCancelled(t *testing.T) {
	inner := &mockProviderForRateLimit{
		response: []string{"translated"},
	}

	provider := NewRateLimitedProvider(inner, RateLimitConfig{
		RequestsPerMinute: 1, // Very slow
		BurstSize:         1,
	})

	// Drain the bucket
	provider.Translate(context.Background(), TranslateRequest{Texts: []string{"a"}})

	// Try with cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := provider.Translate(ctx, TranslateRequest{Texts: []string{"b"}})
	if err == nil {
		t.Error("Expected error when context cancelled")
	}
}

// Mock provider for rate limit tests
type mockProviderForRateLimit struct {
	response []string
	calls    int
}

func (m *mockProviderForRateLimit) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	m.calls++
	return m.response, nil
}
