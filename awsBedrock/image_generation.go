package awsbedrock

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// ImageGenerationConfig represents the configuration for image generation
type ImageGenerationConfig struct {
	NumberOfImages int     `json:"numberOfImages"`
	Quality        string  `json:"quality"`
	Height         int     `json:"height"`
	Width          int     `json:"width"`
	CfgScale       float64 `json:"cfgScale"`
	Seed           int     `json:"seed"`
}

// ImageGenerationRequest represents the request payload for image generation
type ImageGenerationRequest struct {
	TaskType              string                `json:"taskType"`
	TextToImageParams     map[string]string     `json:"textToImageParams"`
	ImageGenerationConfig ImageGenerationConfig `json:"imageGenerationConfig"`
}

// ImageGenerationResponse represents the response from image generation
type ImageGenerationResponse struct {
	Images []string `json:"images"`
}

// GenerateImage creates and returns an image based on text description using Amazon Titan Image Generator
func (wrapper BedrockClient) GenerateImage(textDescription string) ([]byte, error) {
	modelId := "amazon.titan-image-generator-v2:0"
	ctx := context.TODO()

	// Create the request payload
	request := ImageGenerationRequest{
		TaskType: "TEXT_IMAGE",
		TextToImageParams: map[string]string{
			"text": textDescription,
		},
		ImageGenerationConfig: ImageGenerationConfig{
			NumberOfImages: 1,
			Quality:        "standard",
			Height:         1024,
			Width:          1024,
			CfgScale:       8.0,
			Seed:           0,
		},
	}

	// Marshal the request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke the model
	output, err := wrapper.BedrockRuntimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil {
		return nil, fmt.Errorf("model error: %s : %w", modelId, err)
	}

	// Parse the response
	var response ImageGenerationResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if we have any images
	if len(response.Images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	// Decode the base64 image data
	imageBytes, err := base64.StdEncoding.DecodeString(response.Images[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	return imageBytes, nil
}

// GenerateImageWithConfig creates and returns an image based on text description with custom configuration
func (wrapper BedrockClient) GenerateImageWithConfig(textDescription string, config ImageGenerationConfig) ([]byte, error) {
	modelId := "amazon.titan-image-generator-v2:0"
	ctx := context.TODO()

	// Create the request payload
	request := ImageGenerationRequest{
		TaskType: "TEXT_IMAGE",
		TextToImageParams: map[string]string{
			"text": textDescription,
		},
		ImageGenerationConfig: config,
	}

	// Marshal the request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke the model
	output, err := wrapper.BedrockRuntimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil {
		return nil, fmt.Errorf("model error: %s : %w", modelId, err)
	}

	// Parse the response
	var response ImageGenerationResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if we have any images
	if len(response.Images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	// Decode the base64 image data
	imageBytes, err := base64.StdEncoding.DecodeString(response.Images[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	return imageBytes, nil
}
