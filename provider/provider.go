// Package provider defines the AI provider interface and implementations.
package provider

import "github.com/ZaguanLabs/gotlai"

// AIProvider is the interface for AI translation backends.
// This is an alias to the main package interface for convenience.
type AIProvider = gotlai.AIProvider

// TranslateRequest is an alias to the main package type.
type TranslateRequest = gotlai.TranslateRequest
