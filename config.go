package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	Provider     string // LLM provider: groq, openai, anthropic, gemini, openrouter
	APIKey       string // API key for the selected provider
	ChatModel    string
	SystemPrompt string
}

// GetAPIKey returns the API key for the specified provider
func GetAPIKey(provider string) (string, error) {
	var apiKey string
	switch provider {
	case "groq":
		apiKey = os.Getenv("GROQ_API_KEY")
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
	case "anthropic":
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	case "gemini":
		apiKey = os.Getenv("GEMINI_API_KEY")
	case "openrouter":
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	default:
		return "", fmt.Errorf("unsupported LLM provider: %s (supported: groq, openai, anthropic, gemini, openrouter)", provider)
	}

	if apiKey == "" {
		return "", fmt.Errorf("no API key found for %s. Set %s_API_KEY in your environment", provider, os.Getenv(strings.ToUpper(provider)+"_API_KEY")) // Simple heuristic, improved error message below
	}
	return apiKey, nil
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Determine provider (default to groq)
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "groq"
	}

	// Load the appropriate API key
	apiKey, err := GetAPIKey(provider)
	if err != nil {
		// For the initial load, we want to fail hard if the key is missing or provider is invalid
		log.Fatalf("Config error: %v", err)
	}

	// Determine default model per provider
	chatModel := os.Getenv("LLM_MODEL")
	if chatModel == "" {
		switch provider {
		case "groq":
			chatModel = "llama-3.3-70b-versatile"
		case "openai":
			chatModel = "gpt-4o-mini"
		case "anthropic":
			chatModel = "claude-3-5-sonnet-20241022"
		case "gemini":
			chatModel = "gemini-1.5-flash"
		case "openrouter":
			chatModel = "meta-llama/llama-3.1-8b-instruct:free"
		}
	}

	return &Config{
		Provider:     provider,
		APIKey:       apiKey,
		ChatModel:    chatModel,
		SystemPrompt: "You are a helpful assistant. Use the conversation history to provide contextual responses.",
	}, nil
}
