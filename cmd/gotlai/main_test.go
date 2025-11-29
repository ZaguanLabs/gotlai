package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--version"}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "gotlai") {
		t.Errorf("expected version output, got: %s", stdout.String())
	}
}

func TestRun_MissingLang(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error for missing --lang")
	}

	if !strings.Contains(err.Error(), "--lang is required") {
		t.Errorf("expected '--lang is required' error, got: %v", err)
	}
}

func TestRun_MissingAPIKey(t *testing.T) {
	// Temporarily unset OPENAI_API_KEY
	t.Setenv("OPENAI_API_KEY", "")

	var stdout, stderr bytes.Buffer
	err := run([]string{"--lang", "es_ES"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	if !strings.Contains(err.Error(), "API key required") {
		t.Errorf("expected API key error, got: %v", err)
	}
}

func TestRun_DryRun(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.html")
	os.WriteFile(inputFile, []byte("<p>Hello</p><p>World</p>"), 0644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--lang", "es_ES", "--dry-run", inputFile}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Hello") {
		t.Error("dry-run should show 'Hello'")
	}
	if !strings.Contains(output, "World") {
		t.Error("dry-run should show 'World'")
	}
	if !strings.Contains(output, "2 translatable") {
		t.Errorf("dry-run should show node count, got: %s", output)
	}
}

func TestRun_DryRunJSON(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.html")
	os.WriteFile(inputFile, []byte("<p>Hello</p>"), 0644)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--lang", "es_ES", "--dry-run", "--json", inputFile}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("dry-run JSON failed: %v", err)
	}

	var result struct {
		NodeCount int      `json:"node_count"`
		Texts     []string `json:"texts"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.NodeCount != 1 {
		t.Errorf("expected 1 node, got %d", result.NodeCount)
	}
	if len(result.Texts) != 1 || result.Texts[0] != "Hello" {
		t.Errorf("expected ['Hello'], got %v", result.Texts)
	}
}

func TestRun_OutputShortFlag(t *testing.T) {
	// Test that -o is recognized as an alias for --output
	// We can't fully test file output without API key, but we can verify flag parsing
	var stdout, stderr bytes.Buffer

	// This should fail with "API key required" not "unknown flag"
	t.Setenv("OPENAI_API_KEY", "")
	err := run([]string{"--lang", "es_ES", "-o", "output.html", "input.html"}, &stdout, &stderr)

	if err == nil {
		t.Fatal("expected error")
	}

	// Should fail on API key, not on flag parsing
	if !strings.Contains(err.Error(), "API key") && !strings.Contains(err.Error(), "reading file") {
		t.Errorf("expected API key or file error, got: %v", err)
	}
}
