package gollmClient

import (
	"encoding/json"
	"os"
	"testing"
)

// TestGollmClientBasic tests basic functionality of the gollm client
func TestGollmClientBasic(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a new gollm client
	config := Config{
		Provider:  "openai",
		Model:     "gpt-4o-mini",
		APIKey:    apiKey,
		MaxTokens: 100,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create gollm client: %v", err)
	}

	// Test basic generation
	response, err := client.GenerateAI("What is 2+2?", "", nil)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}

// TestGollmClientWithSchema tests schema validation
func TestGollmClientWithSchema(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: OPENAI_API_KEY not set")
	}

	// Create a new gollm client
	config := Config{
		Provider:  "openai",
		Model:     "gpt-4o-mini",
		APIKey:    apiKey,
		MaxTokens: 100,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create gollm client: %v", err)
	}

	// Define a simple JSON schema
	schema := `{
		"type": "object",
		"properties": {
			"answer": {
				"type": "string"
			},
			"confidence": {
				"type": "number",
				"minimum": 0,
				"maximum": 1
			}
		},
		"required": ["answer", "confidence"]
	}`

	// Test schema generation
	response, err := client.GenerateAISchema("What is 2+2? Respond with your answer and confidence level.", "", nil, schema)
	if err != nil {
		t.Fatalf("Failed to generate response with schema: %v", err)
	}

	// Validate that the response is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	// Check required fields
	if _, ok := result["answer"]; !ok {
		t.Error("Response missing 'answer' field")
	}
	if _, ok := result["confidence"]; !ok {
		t.Error("Response missing 'confidence' field")
	}

	t.Logf("Schema response: %s", response)
}

// TestGollmClientFromEnv tests creating client from environment variables
func TestGollmClientFromEnv(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("GOLLM_PROVIDER", "openai")
	os.Setenv("GOLLM_MODEL", "gpt-4o-mini")
	os.Setenv("GOLLM_API_KEY", os.Getenv("OPENAI_API_KEY"))

	// Clean up after test
	defer func() {
		os.Unsetenv("GOLLM_PROVIDER")
		os.Unsetenv("GOLLM_MODEL")
		os.Unsetenv("GOLLM_API_KEY")
	}()

	// Skip if no API key is provided
	if os.Getenv("GOLLM_API_KEY") == "" {
		t.Skip("Skipping test: GOLLM_API_KEY not set")
	}

	// Create client from environment
	client, err := NewFromEnv()
	if err != nil {
		t.Fatalf("Failed to create gollm client from env: %v", err)
	}

	// Test basic functionality
	response, err := client.GenerateAI("Hello, world!", "", nil)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Env client response: %s", response)
}
