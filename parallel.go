package gotlai

import (
	"context"
	"sync"
)

// ParallelCacheLookup performs cache lookups in parallel using goroutines.
// Returns a map of hash to cached value, and a slice of cache misses.
func ParallelCacheLookup(cache TranslationCache, nodes []TextNode, targetLang string) (map[string]string, []TextNode) {
	if cache == nil || len(nodes) == 0 {
		return make(map[string]string), nodes
	}

	type lookupResult struct {
		hash  string
		value string
		found bool
	}

	// Deduplicate nodes by hash first
	uniqueNodes := make(map[string]TextNode)
	for _, node := range nodes {
		if _, exists := uniqueNodes[node.Hash]; !exists {
			uniqueNodes[node.Hash] = node
		}
	}

	// Create channels for results
	results := make(chan lookupResult, len(uniqueNodes))
	var wg sync.WaitGroup

	// Launch goroutines for parallel lookups
	for hash := range uniqueNodes {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			key := CacheKey(h, targetLang)
			if val, ok := cache.Get(key); ok {
				results <- lookupResult{hash: h, value: val, found: true}
			} else {
				results <- lookupResult{hash: h, found: false}
			}
		}(hash)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	translations := make(map[string]string)
	missedHashes := make(map[string]bool)

	for result := range results {
		if result.found {
			translations[result.hash] = result.value
		} else {
			missedHashes[result.hash] = true
		}
	}

	// Build cache misses slice (preserving original order)
	var cacheMisses []TextNode
	seenMisses := make(map[string]bool)
	for _, node := range nodes {
		if missedHashes[node.Hash] && !seenMisses[node.Hash] {
			cacheMisses = append(cacheMisses, node)
			seenMisses[node.Hash] = true
		}
	}

	return translations, cacheMisses
}

// ParallelTranslator is a translator that uses parallel cache lookups.
type ParallelTranslator struct {
	*Translator
	parallelThreshold int // Minimum nodes to trigger parallel lookup
}

// NewParallelTranslator creates a translator with parallel cache lookups.
func NewParallelTranslator(targetLang string, provider AIProvider, opts ...TranslatorOption) *ParallelTranslator {
	return &ParallelTranslator{
		Translator:        NewTranslator(targetLang, provider, opts...),
		parallelThreshold: 5, // Use parallel for 5+ nodes
	}
}

// WithParallelThreshold sets the minimum nodes for parallel lookup.
func (t *ParallelTranslator) WithParallelThreshold(n int) *ParallelTranslator {
	t.parallelThreshold = n
	return t
}

// TranslateBatchParallel translates nodes using parallel cache lookups.
// This is an exported method for advanced use cases.
func (t *ParallelTranslator) TranslateBatchParallel(ctx context.Context, nodes []TextNode) (map[string]string, int, int, error) {
	if t.cache == nil || len(nodes) < t.parallelThreshold {
		// Fall back to sequential for small batches or no cache
		return t.translateBatch(ctx, nodes)
	}

	// Parallel cache lookup
	translations, cacheMisses := ParallelCacheLookup(t.cache, nodes, t.targetLang)
	cachedCount := len(translations)

	// Translate cache misses via AI
	translatedCount := 0
	if len(cacheMisses) > 0 && t.provider != nil {
		texts := make([]string, len(cacheMisses))
		textContexts := make([]string, len(cacheMisses))
		for i, node := range cacheMisses {
			texts[i] = node.Text
			textContexts[i] = node.Context
		}

		results, err := t.provider.Translate(ctx, TranslateRequest{
			Texts:         texts,
			TargetLang:    t.targetLang,
			SourceLang:    t.sourceLang,
			ExcludedTerms: t.excludedTerms,
			Context:       t.context,
			TextContexts:  textContexts,
		})
		if err != nil {
			return nil, 0, 0, err
		}

		// Cache and store results
		for i, node := range cacheMisses {
			translations[node.Hash] = results[i]
			if t.cache != nil {
				cacheKey := CacheKey(node.Hash, t.targetLang)
				_ = t.cache.Set(cacheKey, results[i]) // Ignore cache set errors
			}
			translatedCount++
		}
	}

	return translations, cachedCount, translatedCount, nil
}
