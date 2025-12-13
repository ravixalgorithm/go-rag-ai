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
	"time"

	"github.com/fatih/color"
)

// ConversationMessage stores a single message in the conversation
type ConversationMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

// ChatBot handles RAG-based chat interactions with conversation memory
type ChatBot struct {
	config              *Config
	conversationHistory []ConversationMessage
}

// NewChatBot creates a new ChatBot instance
func NewChatBot(config *Config) *ChatBot {
	return &ChatBot{
		config:              config,
		conversationHistory: make([]ConversationMessage, 0),
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

// AddToHistory adds a message to the conversation history
func (cb *ChatBot) AddToHistory(role, content string) {
	cb.conversationHistory = append(cb.conversationHistory, ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// Query performs a RAG query with conversation context
func (cb *ChatBot) Query(ctx context.Context, question string) (string, error) {
	// Add user message to history
	cb.AddToHistory("user", question)

	// Build messages for Groq API including conversation history
	messages := []GroqMessage{
		{Role: "system", Content: cb.config.SystemPrompt},
	}

	// Add conversation history (keep last 10 exchanges to avoid token limits)
	historyStart := 0
	if len(cb.conversationHistory) > 20 {
		historyStart = len(cb.conversationHistory) - 20
	}

	for _, msg := range cb.conversationHistory[historyStart:] {
		messages = append(messages, GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Call Groq API
	reqBody := GroqChatRequest{
		Model:       cb.config.ChatModel,
		Messages:    messages,
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

	answer := groqResp.Choices[0].Message.Content

	// Add assistant response to history
	cb.AddToHistory("assistant", answer)

	return answer, nil
}

// StreamText prints text with a typing effect
func StreamText(text string, textColor *color.Color) {
	for _, char := range text {
		textColor.Print(string(char))
		time.Sleep(5 * time.Millisecond) // 0.005 seconds per character
	}
	fmt.Println()
}

// GetTimeString returns formatted current time
func GetTimeString() string {
	return time.Now().Format("15:04:05")
}

// RunInteractive starts an interactive chat session
func (cb *ChatBot) RunInteractive(ctx context.Context) error {
	// Color definitions
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	magenta := color.New(color.FgMagenta, color.Bold)
	white := color.New(color.FgWhite)

	// Print welcome message
	cyan.Println("╔════════════════════════╗")
	cyan.Println("║   RAG Chatbot In Go    ║")
	cyan.Println("╚════════════════════════╝")
	yellow.Println("\n💬 I'll remember our conversation! Type your questions.")
	yellow.Println("Commands: 'clear' to clear screen, 'history' to view conversation, 'exit' to quit\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Print prompt with timestamp
		timeStr := GetTimeString()
		green.Printf("You (%s): ", timeStr)

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
			cyan.Println("\n👋 Goodbye! It was nice chatting with you!")
			break
		}

		// Handle history command
		if strings.ToLower(input) == "history" {
			yellow.Printf("\n📜 Conversation History (%d messages):\n\n", len(cb.conversationHistory))
			for _, msg := range cb.conversationHistory {
				if msg.Role == "user" {
					green.Printf("You (%s): %s\n", msg.Timestamp.Format("15:04:05"), msg.Content)
				} else {
					magenta.Printf("Bot (%s): %s\n", msg.Timestamp.Format("15:04:05"), msg.Content)
				}
			}
			fmt.Println()
			continue
		}

		// Handle clear command
		if strings.ToLower(input) == "clear" {
			// Clear screen
			fmt.Print("\033[H\033[2J")
			cyan.Println("\n╔════════════════════════════════════════════╗")
			cyan.Println("║   RAG Chatbot with Conversation Memory    ║")
			cyan.Println("╚════════════════════════════════════════════╝")
			yellow.Printf("\n✨ Screen cleared! Conversation history: %d messages\n\n", len(cb.conversationHistory))
			continue
		}

		// Process question
		answer, err := cb.Query(ctx, input)
		if err != nil {
			red.Printf("\n❌ Error: %v\n\n", err)
			continue
		}

		// Print answer with timestamp and streaming effect
		fmt.Println()
		botTimeStr := GetTimeString()
		magenta.Printf("Bot (%s): ", botTimeStr)
		StreamText(answer, white)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
