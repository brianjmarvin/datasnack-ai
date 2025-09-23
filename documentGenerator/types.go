package documentGenerator

import (
	"time"
)

// DocumentType represents the type of document to generate
type DocumentType string

const (
	DocumentTypeText  DocumentType = "text"
	DocumentTypeCSV   DocumentType = "csv"
	DocumentTypePDF   DocumentType = "pdf"
	DocumentTypeImage DocumentType = "image"
)

// DocumentRequest represents the request payload for document generation
type DocumentRequest struct {
	DocumentType DocumentType `json:"document_type" binding:"required"`
	Content      string       `json:"content" binding:"required"`
	Metadata     Metadata     `json:"metadata"`
	Format       Format       `json:"format"`
}

// Metadata contains additional information about the document
type Metadata struct {
	Title     string            `json:"title,omitempty"`
	Author    string            `json:"author,omitempty"`
	Subject   string            `json:"subject,omitempty"`
	Keywords  []string          `json:"keywords,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	Custom    map[string]string `json:"custom,omitempty"`
}

// Format contains formatting options for the document
type Format struct {
	// For text files
	Encoding string `json:"encoding,omitempty"`

	// For CSV files
	Delimiter string `json:"delimiter,omitempty"`
	Headers   bool   `json:"headers,omitempty"`

	// For PDF files
	PageSize string  `json:"page_size,omitempty"`
	FontSize int     `json:"font_size,omitempty"`
	Margins  Margins `json:"margins,omitempty"`

	// For images
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Format string `json:"format,omitempty"` // png, jpg, etc.
}

// Margins for PDF formatting
type Margins struct {
	Top    float64 `json:"top,omitempty"`
	Bottom float64 `json:"bottom,omitempty"`
	Left   float64 `json:"left,omitempty"`
	Right  float64 `json:"right,omitempty"`
}

// DocumentResponse represents the response from document generation
type DocumentResponse struct {
	DocumentID string   `json:"document_id"`
	FileName   string   `json:"file_name"`
	FileSize   int64    `json:"file_size"`
	MimeType   string   `json:"mime_type"`
	Content    []byte   `json:"content"`
	Metadata   Metadata `json:"metadata"`
}
