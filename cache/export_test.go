package cache

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestExporter_Export(t *testing.T) {
	c := NewInMemoryCache(3600)
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	exporter := NewExporter(c)
	var buf bytes.Buffer

	err := exporter.Export(&buf, map[string]string{"lang": "es_ES"})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse the output
	var export ExportFormat
	if err := json.Unmarshal(buf.Bytes(), &export); err != nil {
		t.Fatalf("Failed to parse export: %v", err)
	}

	if export.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", export.Version)
	}

	if len(export.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(export.Entries))
	}

	if export.Metadata["lang"] != "es_ES" {
		t.Errorf("Expected metadata lang=es_ES, got %v", export.Metadata)
	}
}

func TestImporter_Import(t *testing.T) {
	jsonData := `{
		"version": "1.0",
		"exported_at": "2024-01-01T00:00:00Z",
		"entries": [
			{"key": "key1", "value": "value1"},
			{"key": "key2", "value": "value2"}
		],
		"metadata": {"lang": "es_ES"}
	}`

	c := NewInMemoryCache(3600)
	importer := NewImporter(c)

	result, err := importer.Import(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 2 {
		t.Errorf("Expected 2 imported, got %d", result.Imported)
	}

	if result.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", result.Failed)
	}

	// Verify entries are in cache
	if val, ok := c.Get("key1"); !ok || val != "value1" {
		t.Errorf("key1 not found or wrong value: %s", val)
	}

	if val, ok := c.Get("key2"); !ok || val != "value2" {
		t.Errorf("key2 not found or wrong value: %s", val)
	}
}

func TestExportImport_RoundTrip(t *testing.T) {
	// Create and populate source cache
	src := NewInMemoryCache(3600)
	src.Set("hash1:es_ES", "Hola")
	src.Set("hash2:es_ES", "Mundo")

	// Export
	exporter := NewExporter(src)
	var buf bytes.Buffer
	if err := exporter.Export(&buf, nil); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import into new cache
	dst := NewInMemoryCache(3600)
	importer := NewImporter(dst)
	result, err := importer.Import(&buf)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Imported != 2 {
		t.Errorf("Expected 2 imported, got %d", result.Imported)
	}

	// Verify
	if val, ok := dst.Get("hash1:es_ES"); !ok || val != "Hola" {
		t.Errorf("hash1:es_ES not found or wrong value")
	}
}

func TestExporter_EmptyCache(t *testing.T) {
	c := NewInMemoryCache(3600)
	exporter := NewExporter(c)

	var buf bytes.Buffer
	err := exporter.Export(&buf, nil)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var export ExportFormat
	json.Unmarshal(buf.Bytes(), &export)

	if len(export.Entries) != 0 {
		t.Errorf("Expected 0 entries for empty cache, got %d", len(export.Entries))
	}
}

func TestImporter_InvalidJSON(t *testing.T) {
	c := NewInMemoryCache(3600)
	importer := NewImporter(c)

	_, err := importer.Import(strings.NewReader("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
