// Package llm provides a pluggable interface for LLM providers.
package llm

import "context"

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // text content
}

// LLMClient is the common interface implemented by all LLM providers.
type LLMClient interface {
	// Generate returns the model's response for the given messages.
	Generate(ctx context.Context, messages []Message) (string, error)
}
