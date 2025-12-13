package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	GroqAPIKey      string
	DatabaseURL     string
	EmbeddingModel  string
	ChatModel       string
	ChunkSize       int
	ChunkOverlap    int
	SystemPrompt    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if groqAPIKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required (use Neon PostgreSQL connection string)")
	}

	return &Config{
		GroqAPIKey:     groqAPIKey,
		DatabaseURL:    databaseURL,
		EmbeddingModel: "llama-3.3-70b-versatile", // Groq model for embeddings
		ChatModel:      "llama-3.3-70b-versatile", // Groq model for chat
		ChunkSize:      500,                        // Characters per chunk
		ChunkOverlap:   50,                         // Overlap between chunks
		SystemPrompt:   "You are a helpful assistant.",
	}, nil
}
