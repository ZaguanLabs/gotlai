package gotlai

import (
	"testing"
)

func TestDiffContent_NoChanges(t *testing.T) {
	nodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
	}

	diff := DiffContent(nodes, nodes)

	if diff.HasChanges() {
		t.Error("Expected no changes for identical content")
	}

	if len(diff.Unchanged) != 2 {
		t.Errorf("Expected 2 unchanged, got %d", len(diff.Unchanged))
	}
}

func TestDiffContent_AllNew(t *testing.T) {
	oldNodes := []TextNode{}
	newNodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
	}

	diff := DiffContent(oldNodes, newNodes)

	if len(diff.Added) != 2 {
		t.Errorf("Expected 2 added, got %d", len(diff.Added))
	}

	if len(diff.Removed) != 0 {
		t.Errorf("Expected 0 removed, got %d", len(diff.Removed))
	}
}

func TestDiffContent_AllRemoved(t *testing.T) {
	oldNodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
	}
	newNodes := []TextNode{}

	diff := DiffContent(oldNodes, newNodes)

	if len(diff.Added) != 0 {
		t.Errorf("Expected 0 added, got %d", len(diff.Added))
	}

	if len(diff.Removed) != 2 {
		t.Errorf("Expected 2 removed, got %d", len(diff.Removed))
	}
}

func TestDiffContent_Mixed(t *testing.T) {
	oldNodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
		{Hash: "hash3", Text: "Removed"},
	}
	newNodes := []TextNode{
		{Hash: "hash1", Text: "Hello"},
		{Hash: "hash2", Text: "World"},
		{Hash: "hash4", Text: "Added"},
	}

	diff := DiffContent(oldNodes, newNodes)

	if len(diff.Unchanged) != 2 {
		t.Errorf("Expected 2 unchanged, got %d", len(diff.Unchanged))
	}

	if len(diff.Added) != 1 {
		t.Errorf("Expected 1 added, got %d", len(diff.Added))
	}

	if len(diff.Removed) != 1 {
		t.Errorf("Expected 1 removed, got %d", len(diff.Removed))
	}
}

func TestDiffContentWithContext_DetectsModified(t *testing.T) {
	oldNodes := []TextNode{
		{ID: "node-1", Hash: "hash1", Text: "Hello", Context: "in <h1>"},
		{ID: "node-2", Hash: "hash2", Text: "Welcome", Context: "in <p>"},
	}
	newNodes := []TextNode{
		{ID: "node-1", Hash: "hash3", Text: "Hi", Context: "in <h1>"},     // Modified
		{ID: "node-2", Hash: "hash2", Text: "Welcome", Context: "in <p>"}, // Unchanged
	}

	diff := DiffContentWithContext(oldNodes, newNodes)

	if len(diff.Modified) != 1 {
		t.Errorf("Expected 1 modified, got %d", len(diff.Modified))
	}

	if len(diff.Unchanged) != 1 {
		t.Errorf("Expected 1 unchanged, got %d", len(diff.Unchanged))
	}

	if len(diff.Added) != 0 {
		t.Errorf("Expected 0 added after matching, got %d", len(diff.Added))
	}

	if len(diff.Removed) != 0 {
		t.Errorf("Expected 0 removed after matching, got %d", len(diff.Removed))
	}

	// Check the modified node
	if diff.Modified[0].Old.Text != "Hello" || diff.Modified[0].New.Text != "Hi" {
		t.Errorf("Modified node mismatch: %v", diff.Modified[0])
	}
}

func TestDiffResult_NeedsTranslation(t *testing.T) {
	diff := &DiffResult{
		Added: []TextNode{
			{Hash: "hash1", Text: "New text"},
		},
		Modified: []ModifiedNode{
			{
				Old: TextNode{Hash: "hash2", Text: "Old text"},
				New: TextNode{Hash: "hash3", Text: "Updated text"},
			},
		},
		Unchanged: []TextNode{
			{Hash: "hash4", Text: "Same text"},
		},
	}

	needs := diff.NeedsTranslation()

	if len(needs) != 2 {
		t.Errorf("Expected 2 nodes needing translation, got %d", len(needs))
	}
}

func TestDiffResult_Stats(t *testing.T) {
	diff := &DiffResult{
		Added:     make([]TextNode, 3),
		Removed:   make([]TextNode, 2),
		Unchanged: make([]TextNode, 10),
		Modified:  make([]ModifiedNode, 1),
	}

	stats := diff.Stats()

	if stats.Added != 3 || stats.Removed != 2 || stats.Unchanged != 10 || stats.Modified != 1 {
		t.Errorf("Stats mismatch: %+v", stats)
	}
}

func TestDiffResult_HasChanges(t *testing.T) {
	tests := []struct {
		name     string
		diff     DiffResult
		expected bool
	}{
		{"no changes", DiffResult{Unchanged: make([]TextNode, 5)}, false},
		{"has added", DiffResult{Added: make([]TextNode, 1)}, true},
		{"has removed", DiffResult{Removed: make([]TextNode, 1)}, true},
		{"has modified", DiffResult{Modified: make([]ModifiedNode, 1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.diff.HasChanges() != tt.expected {
				t.Errorf("HasChanges() = %v, want %v", tt.diff.HasChanges(), tt.expected)
			}
		})
	}
}
