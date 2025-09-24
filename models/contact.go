package models

import (
	"time"
)

// ContactUs represents a contact message from a user
type ContactUs struct {
	ContactID   string    `json:"contactId,omitempty"`
	FullName    string    `json:"fullName,omitempty"`
	CompanyName string    `json:"companyName,omitempty"`
	Note        string    `json:"note,omitempty"`
	Email       string    `json:"email,omitempty"`
	ChatbotURL  string    `json:"chatbotUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// ContactUsRequest represents the request payload for contact messages
type ContactUsRequest struct {
	FullName    string `json:"fullName,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	Note        string `json:"note,omitempty"`
	Email       string `json:"email,omitempty"`
	ChatbotURL  string `json:"chatbotUrl,omitempty"`
}

// ContactUsResponse represents the response from contact operations
type ContactUsResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Contact *ContactUs `json:"contact,omitempty"`
	Error   string     `json:"error,omitempty"`
}
