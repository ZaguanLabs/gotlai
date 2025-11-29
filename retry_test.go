package gotlai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetry_Success(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %q", result)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_RetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	result, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", &ProviderError{Message: "rate limited", Retryable: true}
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Expected no error after retries, got: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %q", result)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	_, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "", &ProviderError{Message: "invalid API key", Retryable: false}
	})

	if err == nil {
		t.Fatal("Expected error for non-retryable error")
	}

	// Should not retry non-retryable errors
	if callCount != 1 {
		t.Errorf("Expected 1 call for non-retryable error, got %d", callCount)
	}
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 2,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	_, err := WithRetry(context.Background(), cfg, func() (string, error) {
		callCount++
		return "", &ProviderError{Message: "rate limited", Retryable: true}
	})

	if err == nil {
		t.Fatal("Expected error after max retries")
	}

	// Initial attempt + 2 retries = 3 calls
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 + 2 retries), got %d", callCount)
	}
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second, // Long delay
		MaxDelay:   10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := WithRetry(ctx, cfg, func() (string, error) {
		callCount++
		return "", &ProviderError{Message: "rate limited", Retryable: true}
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"retryable provider error", &ProviderError{Retryable: true}, true},
		{"non-retryable provider error", &ProviderError{Retryable: false}, false},
		{"generic error", errors.New("some error"), false},
		{"context canceled", context.Canceled, false},
		{"context deadline", context.DeadlineExceeded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", cfg.MaxRetries)
	}

	if cfg.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay 1s, got %v", cfg.BaseDelay)
	}

	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay 30s, got %v", cfg.MaxDelay)
	}
}

// Test RetryableProvider
type failingProvider struct {
	failCount int
	callCount int
}

func (p *failingProvider) Translate(ctx context.Context, req TranslateRequest) ([]string, error) {
	p.callCount++
	if p.callCount <= p.failCount {
		return nil, &ProviderError{Message: "temporary failure", Retryable: true}
	}
	return []string{"translated"}, nil
}

func TestRetryableProvider(t *testing.T) {
	inner := &failingProvider{failCount: 2}
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	provider := NewRetryableProvider(inner, cfg)

	result, err := provider.Translate(context.Background(), TranslateRequest{
		Texts:      []string{"hello"},
		TargetLang: "es_ES",
	})

	if err != nil {
		t.Fatalf("Expected success after retries, got: %v", err)
	}

	if len(result) != 1 || result[0] != "translated" {
		t.Errorf("Unexpected result: %v", result)
	}

	if inner.callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", inner.callCount)
	}
}
