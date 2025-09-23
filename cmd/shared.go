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
	FallbackToBedrock    bool             `json:"fallbackToBedrock"`
	LogProviderSelection bool             `json:"logProviderSelection"`
}

type AIClientOption struct {
	Provider    string `json:"provider"`
	Type        string `json:"type"`
	Model       string `json:"model"`
	EnvKey      string `json:"envKey"`
	Endpoint    string `json:"endpoint,omitempty"`
	Description string `json:"description"`
}

// initializeAIClient creates an AI client based on configuration and available API keys
func initializeAIClient() (cloneAttack.AIClient, error) {
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

	// Try each provider in the preferred order
	for i, option := range aiConfig.PreferredOrder {
		if aiConfig.LogProviderSelection {
			log.Printf("Trying AI provider %d/%d: %s (%s)", i+1, len(aiConfig.PreferredOrder), option.Description, option.Type)
		}

		// Check if the required environment variable/key is available
		apiKey := os.Getenv(option.EnvKey)
		if apiKey == "" && option.Provider != "gollm" || option.Type == "ollama" {
			// For Ollama, we don't need an API key, just check if endpoint is accessible
			if option.Type == "ollama" {
				if option.Endpoint == "" {
					option.Endpoint = "http://localhost:11434"
				}
				// We'll try to create the client and let it fail gracefully if Ollama isn't running
			} else {
				if aiConfig.LogProviderSelection {
					log.Printf("Skipping %s: %s not set", option.Description, option.EnvKey)
				}
				continue
			}
		}

		// Try to create the AI client based on provider type
		var client cloneAttack.AIClient
		var clientErr error

		switch option.Provider {
		case "gollm":
			client, clientErr = createGollmClient(option, apiKey)
		case "awsbedrock":
			client, clientErr = createAWSBedrockClient(option)
		default:
			if aiConfig.LogProviderSelection {
				log.Printf("Unknown provider: %s", option.Provider)
			}
			continue
		}

		if clientErr != nil {
			if aiConfig.LogProviderSelection {
				log.Printf("Failed to create %s client: %v", option.Description, clientErr)
			}
			continue
		}

		// Test the client with a simple request
		if testErr := testAIClient(client); testErr != nil {
			if aiConfig.LogProviderSelection {
				log.Printf("Client test failed for %s: %v", option.Description, testErr)
			}
			continue
		}

		if aiConfig.LogProviderSelection {
			log.Printf("Successfully initialized AI client: %s", option.Description)
		}
		return client, nil
	}

	// If no provider worked and fallback is enabled, try AWS Bedrock
	if aiConfig.FallbackToBedrock {
		if aiConfig.LogProviderSelection {
			log.Println("All configured providers failed, falling back to AWS Bedrock")
		}
		bedrockClient := awsbedrock.New()
		if testErr := testAIClient(bedrockClient); testErr != nil {
			return nil, fmt.Errorf("all AI providers failed, including AWS Bedrock fallback: %w", testErr)
		}
		return bedrockClient, nil
	}

	return nil, fmt.Errorf("no AI providers could be initialized - check your API keys and configuration")
}

// createGollmClient creates a gollm client based on the configuration option
func createGollmClient(option AIClientOption, apiKey string) (cloneAttack.AIClient, error) {
	switch option.Type {
	case "openai":
		return gollmClient.NewOpenAIClient(apiKey, option.Model)
	case "anthropic":
		return gollmClient.NewAnthropicClient(apiKey, option.Model)
	case "groq":
		return gollmClient.NewGroqClient(apiKey, option.Model)
	case "ollama":
		endpoint := option.Endpoint
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
			BaseURL:   option.Endpoint,
			MaxTokens: 4000,
		}
		return gollmClient.New(config)
	}
}

// createAWSBedrockClient creates an AWS Bedrock client
func createAWSBedrockClient(option AIClientOption) (cloneAttack.AIClient, error) {
	// AWS Bedrock client doesn't need specific configuration in our current implementation
	// The region and credentials are handled by AWS SDK automatically
	return awsbedrock.New(), nil
}

// testAIClient performs a simple test to verify the AI client is working
func testAIClient(client cloneAttack.AIClient) error {
	// Simple test request
	_, err := client.GenerateAI("Hello", "You are a helpful assistant.", nil)
	return err
}

