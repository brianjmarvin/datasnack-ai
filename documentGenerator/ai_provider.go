package documentGenerator

import (
	"bytes"
	"context"
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
	client AIClient
}

// NewDocumentAIClient creates a new document AI client wrapper
func NewDocumentAIClient(client AIClient) *DocumentAIClient {
	return &DocumentAIClient{client: client}
}

// GenerateText generates text content using the AI client
func (d *DocumentAIClient) GenerateText(ctx context.Context, model string, prompt string) (string, error) {
	// Use the existing AI client's GenerateAI method
	systemPrompt := "You are a helpful assistant that generates content for documents."
	return d.client.GenerateAI(prompt, systemPrompt, nil)
}

// GenerateImage generates an image based on the prompt
func (d *DocumentAIClient) GenerateImage(ctx context.Context, prompt string) (io.Reader, error) {
	// Use the AI client's GenerateImage method
	imageData, err := d.client.GenerateImage(prompt)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(imageData), nil
}
