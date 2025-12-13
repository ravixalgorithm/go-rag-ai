package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
)

// ChatBot handles RAG-based chat interactions
type ChatBot struct {
	vectorStore *VectorStore
	embedder    *Embedder
	config      *Config
}

// NewChatBot creates a new ChatBot instance
func NewChatBot(vectorStore *VectorStore, embedder *Embedder, config *Config) *ChatBot {
	return &ChatBot{
		vectorStore: vectorStore,
		embedder:    embedder,
		config:      config,
	}
}

// GroqChatRequest represents a request to Groq API
type GroqChatRequest struct {
	Model       string        `json:"model"`
	Messages    []GroqMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// GroqMessage represents a message in Groq API
type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GroqChatResponse represents a response from Groq API
type GroqChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Query performs a RAG query
func (cb *ChatBot) Query(ctx context.Context, question string) (string, error) {
	// 1. Get embedding for the question
	queryEmbedding, err := cb.embedder.GetEmbedding(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to get query embedding: %w", err)
	}

	// 2. Search for relevant documents
	results, err := cb.vectorStore.Search(ctx, queryEmbedding, 3)
	if err != nil {
		return "", fmt.Errorf("failed to search documents: %w", err)
	}

	// 3. Build context from results
	context_text := ""
	if len(results) > 0 {
		context_text = "Context:\n"
		for i, result := range results {
			context_text += fmt.Sprintf("\n[Document %d - Similarity: %.2f]\n%s\n",
				i+1, result.Similarity, result.Content)
		}
	}

	// 4. Create prompt with context
	prompt := fmt.Sprintf("%s\n\nQuestion: %s\n\nAnswer based on the context above:",
		context_text, question)

	// 5. Call Groq API for completion
	reqBody := GroqChatRequest{
		Model: cb.config.ChatModel,
		Messages: []GroqMessage{
			{Role: "system", Content: cb.config.SystemPrompt},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cb.config.GroqAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	var groqResp GroqChatResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return groqResp.Choices[0].Message.Content, nil
}

// RunInteractive starts an interactive chat session
func (cb *ChatBot) RunInteractive(ctx context.Context) error {
	// Color definitions
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	// Print welcome message
	cyan.Println("\n╔════════════════════════════════════════╗")
	cyan.Println("║   RAG Chatbot with Groq & pgvector    ║")
	cyan.Println("╚════════════════════════════════════════╝")
	yellow.Println("\nType your questions, or 'exit' to quit.")
	yellow.Println("Commands: 'clear' to clear chat, 'stats' for database info\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Print prompt
		green.Print("You: ")

		// Read user input
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Handle empty input
		if input == "" {
			continue
		}

		// Handle exit command
		if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
			cyan.Println("\nGoodbye! 👋")
			break
		}

		// Handle stats command
		if strings.ToLower(input) == "stats" {
			count, err := cb.vectorStore.Count(ctx)
			if err != nil {
				red.Printf("Error getting stats: %v\n", err)
				continue
			}
			yellow.Printf("\n📊 Database contains %d document chunks\n\n", count)
			continue
		}

		// Handle clear command
		if strings.ToLower(input) == "clear" {
			// Clear screen (simple approach)
			fmt.Print("\033[H\033[2J")
			cyan.Println("\n╔════════════════════════════════════════╗")
			cyan.Println("║   RAG Chatbot with Groq & pgvector    ║")
			cyan.Println("╚════════════════════════════════════════╝\n")
			continue
		}

		// Process question
		cyan.Print("\nBot: ")
		answer, err := cb.Query(ctx, input)
		if err != nil {
			red.Printf("Error: %v\n\n", err)
			continue
		}

		// Print answer
		fmt.Println(answer)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
