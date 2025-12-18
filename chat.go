package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"go-groq/internal/llm"

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
	llmClient           llm.LLMClient
}

// NewChatBot creates a new ChatBot instance
func NewChatBot(config *Config) *ChatBot {
	client, err := llm.NewClient(config.Provider, config.APIKey, config.ChatModel)
	if err != nil {
		panic(fmt.Sprintf("failed to create LLM client: %v", err))
	}
	return &ChatBot{
		config:              config,
		conversationHistory: make([]ConversationMessage, 0),
		llmClient:           client,
	}
}

// SwitchModel switches to a different provider and/or model at runtime.
// provider can be "groq" or "openai"; model is the model name (e.g. "gpt-4o").
func (cb *ChatBot) SwitchModel(provider, model, apiKey string) error {
	client, err := llm.NewClient(provider, apiKey, model)
	if err != nil {
		return err
	}
	cb.llmClient = client
	cb.config.Provider = provider
	cb.config.ChatModel = model
	return nil
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

	// Build messages for LLM including conversation history
	messages := []llm.Message{
		{Role: "system", Content: cb.config.SystemPrompt},
	}

	// Add conversation history (keep last 20 messages to avoid token limits)
	historyStart := 0
	if len(cb.conversationHistory) > 20 {
		historyStart = len(cb.conversationHistory) - 20
	}

	for _, msg := range cb.conversationHistory[historyStart:] {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Call LLM via the pluggable client
	answer, err := cb.llmClient.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

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

// StreamResponseWithCodeHighlight streams response with simple code highlighting
func StreamResponseWithCodeHighlight(text string) {
	white := color.New(color.FgWhite)
	codeBlockColor := color.New(color.FgBlue)
	inlineCodeColor := color.New(color.FgYellow)

	inCodeBlock := false
	inInlineCode := false
	i := 0

	for i < len(text) {
		// Check for code block start/end (```)
		if i+2 < len(text) && text[i:i+3] == "```" {
			if !inCodeBlock {
				// Starting code block
				codeBlockColor.Print("```")
				time.Sleep(5 * time.Millisecond)
				inCodeBlock = true
				i += 3

				// Print language identifier if present (until newline)
				for i < len(text) && text[i] != '\n' {
					codeBlockColor.Print(string(text[i]))
					time.Sleep(5 * time.Millisecond)
					i++
				}
				if i < len(text) && text[i] == '\n' {
					fmt.Println()
					i++
				}
			} else {
				// Ending code block
				codeBlockColor.Print("```")
				time.Sleep(5 * time.Millisecond)
				inCodeBlock = false
				i += 3
			}
			continue
		}

		// Check for inline code (`)
		if text[i] == '`' && !inCodeBlock {
			inlineCodeColor.Print("`")
			time.Sleep(5 * time.Millisecond)
			inInlineCode = !inInlineCode
			i++
			continue
		}

		// Print character with appropriate color
		char := string(text[i])
		if inCodeBlock {
			codeBlockColor.Print(char)
		} else if inInlineCode {
			inlineCodeColor.Print(char)
		} else {
			white.Print(char)
		}

		time.Sleep(5 * time.Millisecond)
		i++
	}
	fmt.Println()
}

// RunInteractive starts an interactive chat session
func (cb *ChatBot) RunInteractive(ctx context.Context) error {
	// Color definitions
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	magenta := color.New(color.FgMagenta, color.Bold)

	// Print welcome message
	cyan.Println("╔════════════════════════╗")
	cyan.Println("║   RAG Chatbot In Go    ║")
	cyan.Println("╚════════════════════════╝")
	yellow.Println("\n💬 I'll remember our conversation! Type your questions.")
	fmt.Print("Commands: ")
	magenta.Print("/clear")
	fmt.Print(", ")
	magenta.Print("/history")
	fmt.Print(", ")
	magenta.Print("/exit")
	fmt.Print(", ")
	magenta.Print("/model <provider> [model]")
	fmt.Println("\n")

	// Channel for user input
	inputChan := make(chan string)
	scanner := bufio.NewScanner(os.Stdin)

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println() // Move to new line after ^C
		cyan.Println("\n👋 Goodbye! (Ctrl+C)")
		os.Exit(0)
	}()

	// Flag to indicate if streaming is in progress
	streaming := false

	for {
		// Print prompt with timestamp only if not streaming
		if !streaming {
			timeStr := GetTimeString()
			green.Printf("You (%s): ", timeStr)
		}

		// Start goroutine to read input
		go func() {
			if scanner.Scan() {
				inputChan <- scanner.Text()
			} else {
				inputChan <- ""
			}
		}()

		// Wait for input (this blocks until user presses enter)
		input := <-inputChan
		input = strings.TrimSpace(input)

		// Handle empty input
		if input == "" {
			continue
		}

		// Handle /exit command
		if strings.ToLower(input) == "/exit" || strings.ToLower(input) == "/quit" {
			cyan.Println("\n👋 Goodbye! It was nice chatting with you!")
			break
		}

		// Handle /history command
		if strings.ToLower(input) == "/history" {
			fmt.Println()
			cyan.Println("╔══════════════════════════════════════════════════════════════╗")
			cyan.Printf("║  📜 Conversation History (%d messages)%s║\n", len(cb.conversationHistory), strings.Repeat(" ", 39-len(fmt.Sprintf("%d", len(cb.conversationHistory)))))
			cyan.Println("╠══════════════════════════════════════════════════════════════╣")
			if len(cb.conversationHistory) == 0 {
				cyan.Println("║  No messages yet.                                            ║")
			}
			for _, msg := range cb.conversationHistory {
				if msg.Role == "user" {
					fmt.Print("║  ")
					green.Printf("You (%s): ", msg.Timestamp.Format("15:04:05"))
					fmt.Printf("%s\n", msg.Content)
				} else {
					fmt.Print("║  ")
					magenta.Printf("%s (%s): ", cb.config.Provider, msg.Timestamp.Format("15:04:05"))
					fmt.Printf("%s\n", msg.Content)
				}
			}
			cyan.Println("╚══════════════════════════════════════════════════════════════╝")
			fmt.Println()
			continue
		}

		// Handle /clear command
		if strings.ToLower(input) == "/clear" {
			// Clear screen
			fmt.Print("\033[H\033[2J")
			cyan.Println("\n╔════════════════════════╗")
			cyan.Println("║   RAG Chatbot In Go    ║")
			cyan.Println("╚════════════════════════╝")
			yellow.Printf("\n✨ Screen cleared! Conversation history: %d messages\n\n", len(cb.conversationHistory))
			continue
		}

		// Handle /model command: /model <provider> [model]
		if strings.HasPrefix(strings.ToLower(input), "/model ") || strings.ToLower(input) == "/model" {
			parts := strings.Fields(input)
			if len(parts) < 2 {
				red.Println("Usage: /model <provider> [model]")
				red.Println("Providers: groq, openai, anthropic, gemini, openrouter")
				continue
			}
			newProvider := strings.ToLower(parts[1])

			// Determine default model if not specified
			var newModel string
			if len(parts) >= 3 {
				newModel = strings.Join(parts[2:], " ") // Allow model names with spaces/slashes
			} else {
				switch newProvider {
				case "groq":
					newModel = "llama-3.3-70b-versatile"
				case "openai":
					newModel = "gpt-4o-mini"
				case "anthropic":
					newModel = "claude-3-5-sonnet-20241022"
				case "gemini":
					newModel = "gemini-1.5-flash"
				case "openrouter":
					newModel = "meta-llama/llama-3.1-8b-instruct:free"
				}
			}

			// Determine API key for the new provider
			var apiKey string
			switch newProvider {
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
				red.Printf("Unknown provider: %s (supported: groq, openai, anthropic, gemini, openrouter)\n", newProvider)
				continue
			}
			if apiKey == "" {
				red.Printf("No API key found for %s. Set %s_API_KEY in your environment.\n", newProvider, strings.ToUpper(newProvider))
				continue
			}

			if err := cb.SwitchModel(newProvider, newModel, apiKey); err != nil {
				red.Printf("Failed to switch model: %v\n", err)
				continue
			}
			green.Printf("✅ Switched to %s / %s\n\n", newProvider, newModel)
			continue
		}

		// Set streaming flag
		streaming = true

		// Show "<provider> is thinking..." indicator
		fmt.Println()
		gray := color.New(color.FgHiBlack)
		gray.Printf("%s is thinking", cb.config.Provider)
		for i := 0; i < 3; i++ {
			time.Sleep(200 * time.Millisecond)
			gray.Print(".")
		}
		fmt.Print("\r\033[K") // Clear the "thinking" line

		// Process question
		answer, err := cb.Query(ctx, input)
		if err != nil {
			red.Printf("\n❌ Error: %v\n\n", err)
			streaming = false
			continue
		}

		// Print answer with timestamp and streaming
		botTimeStr := GetTimeString()
		magenta.Printf("%s (%s): ", cb.config.Provider, botTimeStr)

		// Stream response with simple code highlighting
		StreamResponseWithCodeHighlight(answer)

		fmt.Println()

		// Clear streaming flag - user can now type
		streaming = false
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
