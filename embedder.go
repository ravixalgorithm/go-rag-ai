package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// TextChunk represents a chunk of text with metadata
type TextChunk struct {
	ID       string
	Content  string
	Source   string
	Metadata map[string]string
}

// Embedder handles text chunking and embedding
type Embedder struct {
	config *Config
}

// NewEmbedder creates a new Embedder instance
func NewEmbedder(config *Config) *Embedder {
	return &Embedder{
		config: config,
	}
}

// LoadTextFile loads text from a file
func (e *Embedder) LoadTextFile(filepath string) (string, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// ChunkText splits text into overlapping chunks
func (e *Embedder) ChunkText(text string, source string) []TextChunk {
	chunks := []TextChunk{}
	chunkSize := e.config.ChunkSize
	overlap := e.config.ChunkOverlap

	// Remove extra whitespace
	text = strings.TrimSpace(text)

	// Simple character-based chunking
	for i := 0; i < len(text); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := text[i:end]
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		chunks = append(chunks, TextChunk{
			Content: strings.TrimSpace(chunk),
			Source:  source,
			Metadata: map[string]string{
				"source": source,
				"chunk":  fmt.Sprintf("%d", len(chunks)),
			},
		})

		if end >= len(text) {
			break
		}
	}

	return chunks
}

// GetEmbedding generates embeddings using Groq API
// Note: Groq doesn't have a dedicated embedding endpoint, so we use a workaround
// by getting the model to generate a semantic representation
func (e *Embedder) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Simple hash-based embedding as fallback (for demo purposes)
	// In production, you'd use a proper embedding model or API
	embedding := make([]float32, 384) // 384-dimensional embedding

	// Create a simple numeric representation based on text
	hash := 0
	for i, char := range text {
		hash = (hash*31 + int(char)) % 1000000
		if i < len(embedding) {
			embedding[i] = float32(hash%100) / 100.0
		}
	}

	// Normalize
	var norm float32
	for _, val := range embedding {
		norm += val * val
	}
	norm = float32(1.0 / (norm + 0.0001))
	for i := range embedding {
		embedding[i] *= norm
	}

	return embedding, nil
}

// GetGroqEmbedding attempts to use Groq for semantic understanding
// This is a workaround since Groq doesn't have embedding endpoints
func (e *Embedder) GetGroqEmbedding(ctx context.Context, text string) ([]float32, error) {
	// For now, use the simple embedding
	// In a real system, you'd use OpenAI embeddings or similar
	return e.GetEmbedding(ctx, text)
}

// ProcessFiles loads and chunks multiple text files
func (e *Embedder) ProcessFiles(filepaths []string) ([]TextChunk, error) {
	allChunks := []TextChunk{}

	for _, filepath := range filepaths {
		text, err := e.LoadTextFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", filepath, err)
		}

		chunks := e.ChunkText(text, filepath)
		allChunks = append(allChunks, chunks...)
	}

	return allChunks, nil
}

// Helper function to pretty print JSON
func prettyPrint(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
