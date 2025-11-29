package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Note: time is used for ExportedAt timestamp

// ExportFormat represents the JSON structure for cache export/import.
type ExportFormat struct {
	Version    string            `json:"version"`
	ExportedAt string            `json:"exported_at"`
	Entries    []ExportEntry     `json:"entries"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ExportEntry represents a single cache entry.
type ExportEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Exporter provides cache export functionality.
type Exporter struct {
	cache TranslationCache
}

// NewExporter creates a new cache exporter.
func NewExporter(cache TranslationCache) *Exporter {
	return &Exporter{cache: cache}
}

// Export writes the cache contents to a writer in JSON format.
func (e *Exporter) Export(w io.Writer, metadata map[string]string) error {
	// Get all entries from cache
	entries, err := e.getAllEntries()
	if err != nil {
		return fmt.Errorf("getting cache entries: %w", err)
	}

	export := ExportFormat{
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Entries:    entries,
		Metadata:   metadata,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(export); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return nil
}

// ExportToFile exports the cache to a file.
// The path is provided by the caller and is intentionally user-controlled.
func (e *Exporter) ExportToFile(path string, metadata map[string]string) error {
	f, err := os.Create(path) // #nosec G304 - path is intentionally user-provided
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	return e.Export(f, metadata)
}

// getAllEntries extracts all entries from the cache.
func (e *Exporter) getAllEntries() ([]ExportEntry, error) {
	// Type assert to get internal data
	switch c := e.cache.(type) {
	case *InMemoryCache:
		return e.exportInMemoryCache(c), nil
	default:
		return nil, fmt.Errorf("cache type %T does not support export", e.cache)
	}
}

// exportInMemoryCache exports entries from an in-memory cache.
func (e *Exporter) exportInMemoryCache(c *InMemoryCache) []ExportEntry {
	data := c.Entries()
	entries := make([]ExportEntry, 0, len(data))

	for key, value := range data {
		entries = append(entries, ExportEntry{
			Key:   key,
			Value: value,
		})
	}

	return entries
}

// Importer provides cache import functionality.
type Importer struct {
	cache TranslationCache
}

// NewImporter creates a new cache importer.
func NewImporter(cache TranslationCache) *Importer {
	return &Importer{cache: cache}
}

// Import reads cache entries from a reader and loads them into the cache.
func (i *Importer) Import(r io.Reader) (*ImportResult, error) {
	var export ExportFormat
	if err := json.NewDecoder(r).Decode(&export); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	result := &ImportResult{
		Version:  export.Version,
		Metadata: export.Metadata,
	}

	for _, entry := range export.Entries {
		if err := i.cache.Set(entry.Key, entry.Value); err != nil {
			result.Failed++
			continue
		}
		result.Imported++
	}

	return result, nil
}

// ImportFromFile imports cache entries from a file.
// The path is provided by the caller and is intentionally user-controlled.
func (i *Importer) ImportFromFile(path string) (*ImportResult, error) {
	f, err := os.Open(path) // #nosec G304 - path is intentionally user-provided
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	return i.Import(f)
}

// ImportResult contains statistics about the import operation.
type ImportResult struct {
	Version  string
	Metadata map[string]string
	Imported int
	Failed   int
}

// ExportableCache is an interface for caches that support export.
type ExportableCache interface {
	TranslationCache
	// Keys returns all keys in the cache.
	Keys() []string
}
