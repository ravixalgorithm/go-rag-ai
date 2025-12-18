package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	Provider     string // LLM provider: groq, openai, anthropic, gemini, openrouter
	APIKey       string // API key for the selected provider
	ChatModel    string
	SystemPrompt string
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
	var apiKey string
	switch provider {
	case "groq":
		apiKey = os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			log.Fatal("GROQ_API_KEY environment variable is required when using groq provider")
		}
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI_API_KEY environment variable is required when using openai provider")
		}
	case "anthropic":
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY environment variable is required when using anthropic provider")
		}
	case "gemini":
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable is required when using gemini provider")
		}
	case "openrouter":
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENROUTER_API_KEY environment variable is required when using openrouter provider")
		}
	default:
		log.Fatalf("Unsupported LLM_PROVIDER: %s (supported: groq, openai, anthropic, gemini, openrouter)", provider)
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
