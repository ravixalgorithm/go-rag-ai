package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenRouterClient implements LLMClient for the OpenRouter API.
// OpenRouter provides access to many models via a unified API.
type OpenRouterClient struct {
	apiKey string
	model  string
	http   *http.Client
}

// NewOpenRouterClient creates a new OpenRouter LLM client.
func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey: apiKey,
		model:  model,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// openrouterRequest is the request payload for the OpenRouter API (OpenAI-compatible).
type openrouterRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// openrouterResponse is the response payload from the OpenRouter API.
type openrouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate sends the messages to the OpenRouter API and returns the model's response.
func (c *OpenRouterClient) Generate(ctx context.Context, messages []Message) (string, error) {
	reqBody := openrouterRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/ravixalgorithm/go-rag-ai")
	req.Header.Set("X-Title", "Go RAG AI Chatbot")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("call OpenRouter API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenRouter API error %d: %s", resp.StatusCode, string(body))
	}

	var orResp openrouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(orResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenRouter response")
	}
	return orResp.Choices[0].Message.Content, nil
}
