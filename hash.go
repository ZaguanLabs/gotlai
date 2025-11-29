package gotlai

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashText computes the SHA-256 hash of the trimmed text.
func HashText(text string) string {
	trimmed := strings.TrimSpace(text)
	hash := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(hash[:])
}

// CacheKey generates a cache key from a text hash and target language.
func CacheKey(hash, targetLang string) string {
	return hash + ":" + targetLang
}

// CacheKeyExtended generates an extended cache key including source language and model.
// Use this when you need to differentiate translations by source language or AI model.
func CacheKeyExtended(hash, sourceLang, targetLang, model string) string {
	return hash + ":" + sourceLang + ":" + targetLang + ":" + model
}
