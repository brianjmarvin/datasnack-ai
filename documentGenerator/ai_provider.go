package documentGenerator

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

// AIClient defines the interface for AI clients used in document generation
// This matches the cloneAttack.AIClient interface
type AIClient interface {
	GenerateAI(request string, system string, pastMsgs []map[string]string) (string, error)
	GenerateAISchema(request string, system string, pastMsgs []map[string]string, schema string) (string, error)
	GenerateImage(prompt string) ([]byte, error)
}

// DocumentAIClient wraps the cloneAttack.AIClient to provide document-specific methods
type DocumentAIClient struct {
	textClient  AIClient
	imageClient AIClient
}

// NewDocumentAIClient creates a new document AI client wrapper with a single client
func NewDocumentAIClient(client AIClient) *DocumentAIClient {
	return &DocumentAIClient{
		textClient:  client,
		imageClient: client, // Use the same client for both text and image
	}
}

// NewDocumentAIClientWithManager creates a new document AI client wrapper with separate text and image clients
func NewDocumentAIClientWithManager(textClient, imageClient AIClient) *DocumentAIClient {
	return &DocumentAIClient{
		textClient:  textClient,
		imageClient: imageClient,
	}
}

// GenerateText generates text content using the text AI client
func (d *DocumentAIClient) GenerateText(ctx context.Context, model string, prompt string) (string, error) {
	if d.textClient == nil {
		return "", fmt.Errorf("text AI client not available")
	}
	// Use the text AI client's GenerateAI method
	systemPrompt := "You are a helpful assistant that generates content for documents."
	return d.textClient.GenerateAI(prompt, systemPrompt, nil)
}

// GenerateImage generates an image based on the prompt
func (d *DocumentAIClient) GenerateImage(ctx context.Context, prompt string) (io.Reader, error) {
	if d.imageClient == nil {
		return nil, fmt.Errorf("image AI client not available")
	}
	// Use the image AI client's GenerateImage method
	imageData, err := d.imageClient.GenerateImage(prompt)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(imageData), nil
}
