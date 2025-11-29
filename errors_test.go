package gotlai

import (
	"errors"
	"testing"
)

func TestTranslationError(t *testing.T) {
	cause := errors.New("underlying error")
	err := &TranslationError{Message: "translation failed", Cause: cause}

	if err.Error() != "translation failed: underlying error" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if err.Unwrap() != cause {
		t.Error("Unwrap() should return the cause")
	}

	// Without cause
	err2 := &TranslationError{Message: "simple error"}
	if err2.Error() != "simple error" {
		t.Errorf("unexpected error message: %s", err2.Error())
	}
}

func TestProviderError(t *testing.T) {
	err := &ProviderError{Message: "rate limited", Retryable: true}

	if err.Error() != "provider error: rate limited" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !err.Retryable {
		t.Error("error should be retryable")
	}
}

func TestCacheError(t *testing.T) {
	err := &CacheError{Message: "connection failed"}

	if err.Error() != "cache error: connection failed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestProcessorError(t *testing.T) {
	err := &ProcessorError{Message: "parse failed", ContentType: "html"}

	if err.Error() != "processor error (html): parse failed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestCountMismatchError(t *testing.T) {
	err := &CountMismatchError{Expected: 5, Got: 3}

	expected := "translation count mismatch: expected 5, got 3"
	if err.Error() != expected {
		t.Errorf("unexpected error message: %s, want %s", err.Error(), expected)
	}
}
