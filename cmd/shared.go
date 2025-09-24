/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	awsbedrock "datasnack/awsBedrock"
	"datasnack/cloneAttack"
	"datasnack/gollmClient"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type TestConfiguration struct {
	DataLeakageTests     int `json:"dataLeakageTests"`
	PromptInjectionTests int `json:"promptInjectionTests"`
	ConsistencyTests     int `json:"consistencyTests"`
	IterationsPerTest    int `json:"iterationsPerTest"`
}

type AIClientConfig struct {
	PreferredOrder       []AIClientOption `json:"preferredOrder"`
	LogProviderSelection bool             `json:"logProviderSelection"`
}

type AIClientOption struct {
	Provider     string   `json:"provider"`
	Type         string   `json:"type"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`
	EnvKey       string   `json:"envKey"`
	RegionKey    string   `json:"regionKey"`
	SecretKey    string   `json:"secretKey"`
	Description  string   `json:"description"`
}

// AIClientManager manages multiple AI clients for different capabilities
type AIClientManager struct {
	textClient  cloneAttack.AIClient
	imageClient cloneAttack.AIClient
	config      AIClientConfig
}

// NewAIClientManager creates a new AI client manager with clients for different capabilities
func NewAIClientManager() (*AIClientManager, error) {
	// Load AI client configuration
	configPath := os.Getenv("AI_CLIENT_CONFIG")
	if configPath == "" {
		configPath = "config/aiClientConfig.json"
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read AI client config file: %w", err)
	}

	var aiConfig AIClientConfig
	if err := json.Unmarshal(configData, &aiConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI client config: %w", err)
	}

	manager := &AIClientManager{
		config: aiConfig,
	}

	// Initialize text client
	textClient, err := manager.initializeClientForCapability("text")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize text AI client: %w", err)
	}
	manager.textClient = textClient

	// Initialize image client (optional)
	imageClient, err := manager.initializeClientForCapability("image")
	if err != nil {
		if aiConfig.LogProviderSelection {
			log.Printf("No image AI client available: %v", err)
		}
		// Image client is optional, so we continue without it
	} else {
		manager.imageClient = imageClient
	}

	return manager, nil
}

// GetTextClient returns the text AI client
func (m *AIClientManager) GetTextClient() cloneAttack.AIClient {
	return m.textClient
}

// GetImageClient returns the image AI client (may be nil)
func (m *AIClientManager) GetImageClient() cloneAttack.AIClient {
	return m.imageClient
}

// HasImageCapability returns true if image generation is available
func (m *AIClientManager) HasImageCapability() bool {
	return m.imageClient != nil
}

// initializeClientForCapability finds and initializes a client for the specified capability
func (m *AIClientManager) initializeClientForCapability(capability string) (cloneAttack.AIClient, error) {
	// Try each provider in the preferred order
	for i, option := range m.config.PreferredOrder {
		// Check if this option supports the required capability
		if !contains(option.Capabilities, capability) {
			continue
		}

		if m.config.LogProviderSelection {
			log.Printf("Trying %s AI provider %d/%d: %s (%s)", capability, i+1, len(m.config.PreferredOrder), option.Description, option.Type)
		}

		// Check if the required environment variables are available
		if !m.hasRequiredEnvVars(option) {
			if m.config.LogProviderSelection {
				log.Printf("Skipping %s: missing required environment variables", option.Description)
			}
			continue
		}

		// Try to create the AI client based on provider type
		client, clientErr := m.createClient(option)
		if clientErr != nil {
			if m.config.LogProviderSelection {
				log.Printf("Failed to create %s client: %v", option.Description, clientErr)
			}
			continue
		}

		// Test the client with a simple request
		if testErr := m.testClient(client, capability); testErr != nil {
			if m.config.LogProviderSelection {
				log.Printf("Client test failed for %s: %v", option.Description, testErr)
			}
			continue
		}

		if m.config.LogProviderSelection {
			log.Printf("Successfully initialized %s AI client: %s", capability, option.Description)
		}
		return client, nil
	}

	return nil, fmt.Errorf("no %s AI providers could be initialized - check your API keys and configuration", capability)
}

// hasRequiredEnvVars checks if all required environment variables are set for an option
func (m *AIClientManager) hasRequiredEnvVars(option AIClientOption) bool {
	// Check primary API key
	if option.EnvKey != "" {
		if os.Getenv(option.EnvKey) == "" {
			return false
		}
	}

	// For AWS Bedrock, check region and secret key
	if option.Provider == "awsbedrock" {
		if option.RegionKey != "" && os.Getenv(option.RegionKey) == "" {
			return false
		}
		if option.SecretKey != "" && os.Getenv(option.SecretKey) == "" {
			return false
		}
	}

	// For Ollama, check endpoint
	if option.Type == "ollama" {
		if os.Getenv(option.EnvKey) == "" {
			return false
		}
	}

	return true
}

// createClient creates an AI client based on the configuration option
func (m *AIClientManager) createClient(option AIClientOption) (cloneAttack.AIClient, error) {
	switch option.Provider {
	case "gollm":
		return m.createGollmClient(option)
	case "awsbedrock":
		return m.createAWSBedrockClient(option)
	default:
		return nil, fmt.Errorf("unknown provider: %s", option.Provider)
	}
}

// testClient tests a client with a simple request appropriate for its capability
func (m *AIClientManager) testClient(client cloneAttack.AIClient, capability string) error {
	switch capability {
	case "text":
		return m.testTextClient(client)
	case "image":
		return m.testImageClient(client)
	default:
		return fmt.Errorf("unknown capability: %s", capability)
	}
}

// testTextClient tests a text AI client
func (m *AIClientManager) testTextClient(client cloneAttack.AIClient) error {
	_, err := client.GenerateAI("Hello", "You are a helpful assistant.", nil)
	return err
}

// testImageClient tests an image AI client
func (m *AIClientManager) testImageClient(client cloneAttack.AIClient) error {
	_, err := client.GenerateImage("test")
	return err
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// initializeAIClient creates an AI client based on configuration and available API keys
// This function is kept for backward compatibility but now uses the new manager
func initializeAIClient() (cloneAttack.AIClient, error) {
	manager, err := NewAIClientManager()
	if err != nil {
		return nil, err
	}
	return manager.GetTextClient(), nil
}

// createGollmClient creates a gollm client based on the configuration option
func (m *AIClientManager) createGollmClient(option AIClientOption) (cloneAttack.AIClient, error) {
	apiKey := os.Getenv(option.EnvKey)

	switch option.Type {
	case "openai":
		return gollmClient.NewOpenAIClient(apiKey, option.Model)
	case "anthropic":
		return gollmClient.NewAnthropicClient(apiKey, option.Model)
	case "groq":
		return gollmClient.NewGroqClient(apiKey, option.Model)
	case "ollama":
		endpoint := os.Getenv(option.EnvKey) // OLLAMA_ENDPOINT
		if endpoint == "" {
			endpoint = "http://localhost:11434"
		}
		return gollmClient.NewOllamaClient(option.Model, endpoint)
	default:
		// Generic gollm client creation
		config := gollmClient.Config{
			Provider:  option.Type,
			Model:     option.Model,
			APIKey:    apiKey,
			BaseURL:   os.Getenv(option.EnvKey), // For Ollama, this is the endpoint
			MaxTokens: 4000,
		}
		return gollmClient.New(config)
	}
}

// createAWSBedrockClient creates an AWS Bedrock client with specific model configuration
func (m *AIClientManager) createAWSBedrockClient(option AIClientOption) (cloneAttack.AIClient, error) {
	// Create a new Bedrock client with the specific model
	client := awsbedrock.NewWithModel(option.Model)
	if client == nil {
		return nil, fmt.Errorf("failed to create AWS Bedrock client for model: %s", option.Model)
	}
	return client, nil
}
