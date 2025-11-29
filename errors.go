package gotlai

import "fmt"

// TranslationError is the base error type for translation failures.
type TranslationError struct {
	Message string
	Cause   error
}

func (e *TranslationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *TranslationError) Unwrap() error {
	return e.Cause
}

// ProviderError indicates an AI provider failure (API error, rate limit, etc.).
type ProviderError struct {
	Message   string
	Cause     error
	Retryable bool // Whether the operation can be retried
}

func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("provider error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("provider error: %s", e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// CacheError indicates a cache operation failure.
type CacheError struct {
	Message string
	Cause   error
}

func (e *CacheError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("cache error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("cache error: %s", e.Message)
}

func (e *CacheError) Unwrap() error {
	return e.Cause
}

// ProcessorError indicates a content processing failure (parse error, etc.).
type ProcessorError struct {
	Message     string
	Cause       error
	ContentType string // The type of content that failed to process
}

func (e *ProcessorError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("processor error (%s): %s: %v", e.ContentType, e.Message, e.Cause)
	}
	return fmt.Sprintf("processor error (%s): %s", e.ContentType, e.Message)
}

func (e *ProcessorError) Unwrap() error {
	return e.Cause
}

// CountMismatchError indicates the AI returned a different number of translations than expected.
type CountMismatchError struct {
	Expected int
	Got      int
}

func (e *CountMismatchError) Error() string {
	return fmt.Sprintf("translation count mismatch: expected %d, got %d", e.Expected, e.Got)
}
