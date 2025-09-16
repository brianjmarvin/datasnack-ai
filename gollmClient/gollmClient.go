package gollmClient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/teilomillet/gollm"
)

// GollmClient implements the AIClient interface using the gollm library
// This provides unified access to multiple LLM providers (OpenAI, Anthropic, Groq, Ollama, etc.)
type GollmClient struct {
	llm     gollm.LLM
	context context.Context
}

// Config holds configuration for the gollm client
type Config struct {
	Provider  string `json:"provider"`  // openai, anthropic, groq, ollama, etc.
	Model     string `json:"model"`     // Model name (e.g., gpt-4o, claude-3-5-sonnet)
	APIKey    string `json:"apiKey"`    // API key for the provider
	BaseURL   string `json:"baseURL"`   // Optional: custom base URL (for Ollama, etc.)
	MaxTokens int    `json:"maxTokens"` // Optional: max tokens per request
}

// New creates a new GollmClient instance
func New(config Config) (*GollmClient, error) {
	// Set default values
	if config.MaxTokens == 0 {
		config.MaxTokens = 4000
	}

	// Create the LLM instance using gollm configuration options
	llm, err := gollm.NewLLM(
		gollm.SetProvider(config.Provider),
		gollm.SetModel(config.Model),
		gollm.SetAPIKey(config.APIKey),
		gollm.SetMaxTokens(config.MaxTokens),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm LLM: %w", err)
	}

	// Set base URL if provided (useful for local Ollama instances)
	if config.BaseURL != "" {
		llm.SetOption("base_url", config.BaseURL)
	}

	return &GollmClient{
		llm:     llm,
		context: context.Background(),
	}, nil
}

// NewFromEnv creates a new GollmClient using environment variables
// Expected env vars: GOLLM_PROVIDER, GOLLM_MODEL, GOLLM_API_KEY, GOLLM_BASE_URL (optional)
func NewFromEnv() (*GollmClient, error) {
	config := Config{
		Provider:  os.Getenv("GOLLM_PROVIDER"),
		Model:     os.Getenv("GOLLM_MODEL"),
		APIKey:    os.Getenv("GOLLM_API_KEY"),
		BaseURL:   os.Getenv("GOLLM_BASE_URL"),
		MaxTokens: 4000, // Default
	}

	if config.Provider == "" {
		return nil, fmt.Errorf("GOLLM_PROVIDER environment variable is required")
	}
	if config.Model == "" {
		return nil, fmt.Errorf("GOLLM_MODEL environment variable is required")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("GOLLM_API_KEY environment variable is required")
	}

	return New(config)
}

// GenerateAI implements the AIClient interface
// request: the user's prompt/request
// system: the system prompt/instructions
// pastMsgs: previous conversation messages (format: [{"role": "user", "content": "..."}, {"role": "assistant", "content": "..."}])
func (g *GollmClient) GenerateAI(request string, system string, pastMsgs []map[string]string) (string, error) {
	// Build the conversation history
	var messages []gollm.PromptMessage

	// Add system message if provided
	if system != "" {
		messages = append(messages, gollm.PromptMessage{
			Role:    "system",
			Content: system,
		})
	}

	// Add past messages
	for _, msg := range pastMsgs {
		role, hasRole := msg["role"]
		content, hasContent := msg["content"]

		if !hasRole || !hasContent {
			continue // Skip malformed messages
		}

		// Normalize role names
		switch strings.ToLower(role) {
		case "user", "human":
			messages = append(messages, gollm.PromptMessage{
				Role:    "user",
				Content: content,
			})
		case "assistant", "ai", "bot":
			messages = append(messages, gollm.PromptMessage{
				Role:    "assistant",
				Content: content,
			})
		case "system":
			messages = append(messages, gollm.PromptMessage{
				Role:    "system",
				Content: content,
			})
		}
	}

	// Add the current request
	messages = append(messages, gollm.PromptMessage{
		Role:    "user",
		Content: request,
	})

	// Create prompt with conversation history
	prompt := gollm.NewPrompt("", gollm.WithMessages(messages))

	// Generate response
	response, err := g.llm.Generate(g.context, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}

	return response, nil
}

// GenerateAISchema implements the AIClient interface with JSON schema validation
// request: the user's prompt/request
// system: the system prompt/instructions
// pastMsgs: previous conversation messages
// schema: JSON schema string for structured output validation
func (g *GollmClient) GenerateAISchema(request string, system string, pastMsgs []map[string]string, schema string) (string, error) {
	// Build the conversation history (same as GenerateAI)
	var messages []gollm.PromptMessage

	// Add system message if provided
	if system != "" {
		messages = append(messages, gollm.PromptMessage{
			Role:    "system",
			Content: system,
		})
	}

	// Add past messages
	for _, msg := range pastMsgs {
		role, hasRole := msg["role"]
		content, hasContent := msg["content"]

		if !hasRole || !hasContent {
			continue // Skip malformed messages
		}

		// Normalize role names
		switch strings.ToLower(role) {
		case "user", "human":
			messages = append(messages, gollm.PromptMessage{
				Role:    "user",
				Content: content,
			})
		case "assistant", "ai", "bot":
			messages = append(messages, gollm.PromptMessage{
				Role:    "assistant",
				Content: content,
			})
		case "system":
			messages = append(messages, gollm.PromptMessage{
				Role:    "system",
				Content: content,
			})
		}
	}

	// Add the current request with schema instruction
	schemaInstruction := fmt.Sprintf("%s\n\nPlease respond with valid JSON that matches this schema:\n%s", request, schema)
	messages = append(messages, gollm.PromptMessage{
		Role:    "user",
		Content: schemaInstruction,
	})

	// Create prompt with conversation history
	prompt := gollm.NewPrompt("", gollm.WithMessages(messages))

	// Parse schema for structured output
	var schemaInterface interface{}
	if err := json.Unmarshal([]byte(schema), &schemaInterface); err != nil {
		return "", fmt.Errorf("invalid JSON schema: %w", err)
	}

	// Generate response with JSON schema validation
	response, err := g.llm.GenerateWithSchema(g.context, prompt, schemaInterface)
	if err != nil {
		return "", fmt.Errorf("failed to generate AI response with schema validation: %w", err)
	}

	// Validate that the response is valid JSON
	var jsonResponse interface{}
	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		return "", fmt.Errorf("response is not valid JSON: %w", err)
	}

	return response, nil
}

// GetProvider returns the current provider name
func (g *GollmClient) GetProvider() string {
	return g.llm.GetProvider()
}

// GetModel returns the current model name
func (g *GollmClient) GetModel() string {
	return g.llm.GetModel()
}

// SetMaxTokens updates the maximum tokens for requests
func (g *GollmClient) SetMaxTokens(maxTokens int) {
	g.llm.SetOption("max_tokens", maxTokens)
}

// SetTemperature updates the temperature for requests (if supported by provider)
func (g *GollmClient) SetTemperature(temperature float64) {
	g.llm.SetOption("temperature", temperature)
}

// SetTopP updates the top_p for requests (if supported by provider)
func (g *GollmClient) SetTopP(topP float64) {
	g.llm.SetOption("top_p", topP)
}

// Close cleans up resources (if needed)
func (g *GollmClient) Close() error {
	// gollm doesn't require explicit cleanup, but this method is here for interface consistency
	return nil
}

// Helper function to create common provider configurations
func NewOpenAIClient(apiKey, model string) (*GollmClient, error) {
	return New(Config{
		Provider:  "openai",
		Model:     model,
		APIKey:    apiKey,
		MaxTokens: 4000,
	})
}

func NewAnthropicClient(apiKey, model string) (*GollmClient, error) {
	return New(Config{
		Provider:  "anthropic",
		Model:     model,
		APIKey:    apiKey,
		MaxTokens: 4000,
	})
}

func NewGroqClient(apiKey, model string) (*GollmClient, error) {
	return New(Config{
		Provider:  "groq",
		Model:     model,
		APIKey:    apiKey,
		MaxTokens: 4000,
	})
}

func NewOllamaClient(model, baseURL string) (*GollmClient, error) {
	return New(Config{
		Provider:  "ollama",
		Model:     model,
		APIKey:    "", // Ollama doesn't require API key
		BaseURL:   baseURL,
		MaxTokens: 4000,
	})
}
