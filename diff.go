package gotlai

// DiffResult represents the difference between two content versions.
type DiffResult struct {
	// Added contains text nodes that are new (not in the previous version).
	Added []TextNode

	// Removed contains text nodes that were removed (not in the new version).
	Removed []TextNode

	// Unchanged contains text nodes that exist in both versions.
	Unchanged []TextNode

	// Modified contains pairs of nodes where the text changed but context suggests same element.
	// This is a heuristic based on position/context similarity.
	Modified []ModifiedNode
}

// ModifiedNode represents a text node that was modified.
type ModifiedNode struct {
	Old TextNode
	New TextNode
}

// Stats returns summary statistics for the diff.
func (d *DiffResult) Stats() DiffStats {
	return DiffStats{
		Added:     len(d.Added),
		Removed:   len(d.Removed),
		Unchanged: len(d.Unchanged),
		Modified:  len(d.Modified),
	}
}

// DiffStats contains summary statistics for a diff.
type DiffStats struct {
	Added     int
	Removed   int
	Unchanged int
	Modified  int
}

// HasChanges returns true if there are any differences.
func (d *DiffResult) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Modified) > 0
}

// NeedsTranslation returns the nodes that need to be translated.
// This includes new nodes and modified nodes.
func (d *DiffResult) NeedsTranslation() []TextNode {
	result := make([]TextNode, 0, len(d.Added)+len(d.Modified))
	result = append(result, d.Added...)
	for _, m := range d.Modified {
		result = append(result, m.New)
	}
	return result
}

// DiffContent compares two sets of text nodes and returns the differences.
// This is useful for incremental translation - only translate what changed.
func DiffContent(oldNodes, newNodes []TextNode) *DiffResult {
	result := &DiffResult{}

	// Build maps for efficient lookup
	oldByHash := make(map[string]TextNode)
	newByHash := make(map[string]TextNode)

	for _, node := range oldNodes {
		oldByHash[node.Hash] = node
	}
	for _, node := range newNodes {
		newByHash[node.Hash] = node
	}

	// Find unchanged and removed
	for hash, oldNode := range oldByHash {
		if _, exists := newByHash[hash]; exists {
			result.Unchanged = append(result.Unchanged, oldNode)
		} else {
			result.Removed = append(result.Removed, oldNode)
		}
	}

	// Find added
	for hash, newNode := range newByHash {
		if _, exists := oldByHash[hash]; !exists {
			result.Added = append(result.Added, newNode)
		}
	}

	return result
}

// DiffContentWithContext performs a more sophisticated diff that tries to detect
// modified nodes (same position/context, different text).
func DiffContentWithContext(oldNodes, newNodes []TextNode) *DiffResult {
	result := DiffContent(oldNodes, newNodes)

	// Try to match removed nodes with added nodes based on context/ID
	if len(result.Added) > 0 && len(result.Removed) > 0 {
		matched := make(map[int]bool) // indices of matched added nodes
		removedMatched := make(map[int]bool)

		for ri, removed := range result.Removed {
			for ai, added := range result.Added {
				if matched[ai] {
					continue
				}

				// Match by ID (same position in document)
				if removed.ID == added.ID {
					result.Modified = append(result.Modified, ModifiedNode{
						Old: removed,
						New: added,
					})
					matched[ai] = true
					removedMatched[ri] = true
					break
				}

				// Match by similar context
				if removed.Context != "" && removed.Context == added.Context {
					result.Modified = append(result.Modified, ModifiedNode{
						Old: removed,
						New: added,
					})
					matched[ai] = true
					removedMatched[ri] = true
					break
				}
			}
		}

		// Filter out matched nodes from Added and Removed
		newAdded := make([]TextNode, 0)
		for i, node := range result.Added {
			if !matched[i] {
				newAdded = append(newAdded, node)
			}
		}
		result.Added = newAdded

		newRemoved := make([]TextNode, 0)
		for i, node := range result.Removed {
			if !removedMatched[i] {
				newRemoved = append(newRemoved, node)
			}
		}
		result.Removed = newRemoved
	}

	return result
}
