package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize chatbot with conversation memory
	chatBot := NewChatBot(config)

	// Run interactive chat
	if err := chatBot.RunInteractive(ctx); err != nil {
		log.Fatalf("Chat error: %v", err)
	}
}
