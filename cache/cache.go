// Package cache provides translation caching implementations.
package cache

// TranslationCache is the interface for translation caching.
type TranslationCache interface {
	// Get retrieves a cached translation. Returns empty string and false if not found or expired.
	Get(key string) (string, bool)

	// Set stores a translation in the cache.
	Set(key string, value string) error
}
