package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	GroqAPIKey   string
	ChatModel    string
	SystemPrompt string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if groqAPIKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	return &Config{
		GroqAPIKey:   groqAPIKey,
		ChatModel:    "llama-3.3-70b-versatile",
		SystemPrompt: "You are a helpful assistant. Use the conversation history to provide contextual responses.",
	}, nil
}
