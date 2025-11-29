package gotlai

import (
	"context"
	"errors"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxRetries int           // Maximum number of retry attempts
	BaseDelay  time.Duration // Initial delay between retries
	MaxDelay   time.Duration // Maximum delay between retries
}

// DefaultRetryConfig returns sensible defaults for retry behavior.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
}

// RetryFunc is a function that can be retried.
type RetryFunc[T any] func() (T, error)

// WithRetry executes a function with exponential backoff retry.
func WithRetry[T any](ctx context.Context, cfg RetryConfig, fn RetryFunc[T]) (T, error) {
	var lastErr error
	var zero T

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			return zero, err
		}

		// Don't sleep after the last attempt
		if attempt < cfg.MaxRetries {
			delay := cfg.BaseDelay * time.Duration(1<<attempt)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return zero, lastErr
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for ProviderError with Retryable flag
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Retryable
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return false
}

// RetryableProvider wraps an AIProvider with retry logic.
type RetryableProvider struct {
	provider AIProvider
	config   RetryConfig
}

// NewRetryableProvider creates a new provider with retry logic.
func NewRetryableProvider(provider AIProvider, cfg RetryConfig) *RetryableProvider {
	return &RetryableProvider{
		provider: provider,
		config:   cfg,
	}
}

// Translate implements AIProvider with retry logic.
func (p *RetryableProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	return WithRetry(ctx, p.config, func() ([]string, error) {
		return p.provider.Translate(ctx, req)
	})
}
