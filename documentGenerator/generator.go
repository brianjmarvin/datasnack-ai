package documentGenerator

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// DocumentGenerator handles the generation of different document types
type DocumentGenerator struct {
	aiClient *DocumentAIClient
}

// NewDocumentGenerator creates a new document generator
func NewDocumentGenerator(aiClient AIClient) *DocumentGenerator {
	return &DocumentGenerator{
		aiClient: NewDocumentAIClient(aiClient),
	}
}

// NewDocumentGeneratorWithAIClient creates a new document generator with a specific DocumentAIClient
func NewDocumentGeneratorWithAIClient(aiClient *DocumentAIClient) *DocumentGenerator {
	return &DocumentGenerator{
		aiClient: aiClient,
	}
}

// GenerateDocument generates a document based on the request
func (dg *DocumentGenerator) GenerateDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	// Set default values
	if req.Metadata.CreatedAt.IsZero() {
		req.Metadata.CreatedAt = time.Now()
	}

	switch req.DocumentType {
	case DocumentTypeText:
		return dg.generateTextDocument(ctx, req)
	case DocumentTypeCSV:
		return dg.generateCSVDocument(ctx, req)
	case DocumentTypePDF:
		return dg.generatePDFDocument(ctx, req)
	case DocumentTypeImage:
		return dg.generateImageDocument(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported document type: %s", req.DocumentType)
	}
}

// generateTextDocument generates a plain text document
func (dg *DocumentGenerator) generateTextDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	// Use AI to enhance the content if it's a prompt
	content := req.Content
	if strings.Contains(strings.ToLower(content), "generate") || strings.Contains(strings.ToLower(content), "create") {
		enhancedContent, err := dg.aiClient.GenerateText(ctx, "", content)
		if err != nil {
			return nil, fmt.Errorf("failed to generate text content: %w", err)
		}
		content = enhancedContent
	}

	// Add metadata as header if provided
	var finalContent strings.Builder
	if req.Metadata.Title != "" {
		finalContent.WriteString(fmt.Sprintf("Title: %s\n", req.Metadata.Title))
	}
	if req.Metadata.Author != "" {
		finalContent.WriteString(fmt.Sprintf("Author: %s\n", req.Metadata.Author))
	}
	if req.Metadata.Subject != "" {
		finalContent.WriteString(fmt.Sprintf("Subject: %s\n", req.Metadata.Subject))
	}
	if len(req.Metadata.Keywords) > 0 {
		finalContent.WriteString(fmt.Sprintf("Keywords: %s\n", strings.Join(req.Metadata.Keywords, ", ")))
	}
	if !req.Metadata.CreatedAt.IsZero() {
		finalContent.WriteString(fmt.Sprintf("Created: %s\n", req.Metadata.CreatedAt.Format(time.RFC3339)))
	}
	if finalContent.Len() > 0 {
		finalContent.WriteString("\n")
	}
	finalContent.WriteString(content)

	fileName := fmt.Sprintf("document_%d.txt", time.Now().Unix())
	if req.Metadata.Title != "" {
		fileName = fmt.Sprintf("%s.txt", sanitizeFileName(req.Metadata.Title))
	}

	return &DocumentResponse{
		DocumentID: generateDocumentID(),
		FileName:   fileName,
		FileSize:   int64(len(finalContent.String())),
		MimeType:   "text/plain",
		Content:    []byte(finalContent.String()),
		Metadata:   req.Metadata,
	}, nil
}

// generateCSVDocument generates a CSV document
func (dg *DocumentGenerator) generateCSVDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	// Create a prompt using the working format: "Generate synthetic [content] csv file"
	prompt := fmt.Sprintf("Generate synthetic csv file data based on the following request: %s", req.Content)

	aiContent, err := dg.aiClient.GenerateText(ctx, "", prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSV content: %w", err)
	}

	// Clean up the AI response - remove any markdown formatting or extra text
	aiContent = strings.TrimSpace(aiContent)

	// Remove markdown code blocks if present
	if strings.HasPrefix(aiContent, "```") {
		lines := strings.Split(aiContent, "\n")
		var csvLines []string
		inCodeBlock := false
		for _, line := range lines {
			if strings.HasPrefix(line, "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			if inCodeBlock {
				csvLines = append(csvLines, line)
			}
		}
		aiContent = strings.Join(csvLines, "\n")
	}

	// Parse and validate CSV content
	reader := csv.NewReader(strings.NewReader(aiContent))
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("failed to generate valid CSV data: %w", err)
	}

	// Write CSV to buffer
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	delimiter := ','
	if req.Format.Delimiter != "" {
		delimiter = rune(req.Format.Delimiter[0])
	}
	writer.Comma = delimiter

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	fileName := fmt.Sprintf("document_%d.csv", time.Now().Unix())
	if req.Metadata.Title != "" {
		fileName = fmt.Sprintf("%s.csv", sanitizeFileName(req.Metadata.Title))
	}

	return &DocumentResponse{
		DocumentID: generateDocumentID(),
		FileName:   fileName,
		FileSize:   int64(buf.Len()),
		MimeType:   "text/csv",
		Content:    buf.Bytes(),
		Metadata:   req.Metadata,
	}, nil
}

// generatePDFDocument generates a PDF document
func (dg *DocumentGenerator) generatePDFDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	// Use AI to generate content
	content, err := dg.aiClient.GenerateText(ctx, "", req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF content: %w", err)
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	fontSize := 12.0
	if req.Format.FontSize > 0 {
		fontSize = float64(req.Format.FontSize)
	}
	pdf.SetFont("Arial", "", fontSize)

	// Add metadata
	if req.Metadata.Title != "" {
		pdf.SetFont("Arial", "B", fontSize+2.0)
		pdf.Cell(0, 10, req.Metadata.Title)
		pdf.Ln(10)
		pdf.SetFont("Arial", "", fontSize)
	}

	if req.Metadata.Author != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Author: %s", req.Metadata.Author))
		pdf.Ln(6)
	}

	if req.Metadata.Subject != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Subject: %s", req.Metadata.Subject))
		pdf.Ln(6)
	}

	if len(req.Metadata.Keywords) > 0 {
		pdf.Cell(0, 6, fmt.Sprintf("Keywords: %s", strings.Join(req.Metadata.Keywords, ", ")))
		pdf.Ln(6)
	}

	pdf.Ln(10)

	// Add content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		pdf.Cell(0, 6, line)
		pdf.Ln(6)
	}

	// Get PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	fileName := fmt.Sprintf("document_%d.pdf", time.Now().Unix())
	if req.Metadata.Title != "" {
		fileName = fmt.Sprintf("%s.pdf", sanitizeFileName(req.Metadata.Title))
	}

	return &DocumentResponse{
		DocumentID: generateDocumentID(),
		FileName:   fileName,
		FileSize:   int64(buf.Len()),
		MimeType:   "application/pdf",
		Content:    buf.Bytes(),
		Metadata:   req.Metadata,
	}, nil
}

// generateImageDocument generates an image document
func (dg *DocumentGenerator) generateImageDocument(ctx context.Context, req *DocumentRequest) (*DocumentResponse, error) {
	// Use AI to generate image
	imageReader, err := dg.aiClient.GenerateImage(ctx, req.Content)
	if err != nil {
		// If image generation fails, create a placeholder text file explaining the issue
		errorContent := fmt.Sprintf(`Image Generation Failed

Request: %s
Error: %s

This is a placeholder document because the image generation service is currently unavailable.
The AI image generation feature requires proper configuration and valid model access.

To resolve this issue:
1. Verify your AI client configuration includes an image-capable model
2. Ensure the required environment variables are set (e.g., AWS credentials for Bedrock)
3. Check that your AI provider has image generation capabilities

Generated on: %s`, req.Content, err.Error(), time.Now().Format(time.RFC3339))

		fileName := fmt.Sprintf("image_placeholder_%d.txt", time.Now().Unix())
		if req.Metadata.Title != "" {
			fileName = fmt.Sprintf("%s_placeholder.txt", sanitizeFileName(req.Metadata.Title))
		}

		return &DocumentResponse{
			DocumentID: generateDocumentID(),
			FileName:   fileName,
			FileSize:   int64(len(errorContent)),
			MimeType:   "text/plain",
			Content:    []byte(errorContent),
			Metadata:   req.Metadata,
		}, nil
	}

	// Read image data
	imageData, err := io.ReadAll(imageReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	fileName := fmt.Sprintf("document_%d.png", time.Now().Unix())
	if req.Metadata.Title != "" {
		fileName = fmt.Sprintf("%s.png", sanitizeFileName(req.Metadata.Title))
	}

	return &DocumentResponse{
		DocumentID: generateDocumentID(),
		FileName:   fileName,
		FileSize:   int64(len(imageData)),
		MimeType:   "image/png",
		Content:    imageData,
		Metadata:   req.Metadata,
	}, nil
}

// Helper functions
func generateDocumentID() string {
	return fmt.Sprintf("doc_%d", time.Now().UnixNano())
}

func sanitizeFileName(name string) string {
	// Remove or replace invalid characters for filenames
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}
