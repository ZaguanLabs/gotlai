package gotlai

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of API requests using a token bucket algorithm.
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	RequestsPerMinute int // Maximum requests per minute
	BurstSize         int // Maximum burst size (default: same as RPM)
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	rpm := float64(cfg.RequestsPerMinute)
	if rpm <= 0 {
		rpm = 60 // Default: 60 RPM
	}

	burst := float64(cfg.BurstSize)
	if burst <= 0 {
		burst = rpm // Default burst = RPM
	}

	return &RateLimiter{
		tokens:     burst, // Start with full bucket
		maxTokens:  burst,
		refillRate: rpm / 60.0, // Convert to tokens per second
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.TryAcquire() {
			return nil
		}

		// Calculate wait time for next token
		r.mu.Lock()
		waitTime := time.Duration(float64(time.Second) / r.refillRate)
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens >= 1 {
		r.tokens--
		return true
	}

	return false
}

// refill adds tokens based on elapsed time (must be called with lock held).
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.lastRefill = now

	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
}

// Available returns the current number of available tokens.
func (r *RateLimiter) Available() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	return r.tokens
}

// RateLimitedProvider wraps an AIProvider with rate limiting.
type RateLimitedProvider struct {
	provider AIProvider
	limiter  *RateLimiter
}

// NewRateLimitedProvider creates a new rate-limited provider.
func NewRateLimitedProvider(provider AIProvider, cfg RateLimitConfig) *RateLimitedProvider {
	return &RateLimitedProvider{
		provider: provider,
		limiter:  NewRateLimiter(cfg),
	}
}

// Translate implements AIProvider with rate limiting.
func (p *RateLimitedProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	// Wait for rate limit
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, &ProviderError{
			Message:   "rate limit wait cancelled",
			Cause:     err,
			Retryable: false,
		}
	}

	return p.provider.Translate(ctx, req)
}

// Limiter returns the underlying rate limiter for inspection.
func (p *RateLimitedProvider) Limiter() *RateLimiter {
	return p.limiter
}
