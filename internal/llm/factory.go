package llm

import "fmt"

// NewClient returns an LLMClient for the specified provider.
// Supported providers: "groq", "openai", "anthropic", "gemini", "openrouter".
func NewClient(provider, apiKey, model string) (LLMClient, error) {
	switch provider {
	case "groq":
		return NewGroqClient(apiKey, model), nil
	case "openai":
		return NewOpenAIClient(apiKey, model), nil
	case "anthropic":
		return NewAnthropicClient(apiKey, model), nil
	case "gemini":
		return NewGeminiClient(apiKey, model), nil
	case "openrouter":
		return NewOpenRouterClient(apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q", provider)
	}
}
