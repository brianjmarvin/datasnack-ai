package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"datasnack/db"
	"datasnack/email"
	"datasnack/models"
)

// ContactHandler handles contact-related HTTP requests
type ContactHandler struct {
	DB *db.Database
}

// NewContactHandler creates a new ContactHandler instance
func NewContactHandler(database *db.Database) *ContactHandler {
	return &ContactHandler{
		DB: database,
	}
}

// SendNote handles POST /api/v1/sendNote - Send a contact message
func (h *ContactHandler) SendNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("In SendNote")

	var contactReq models.ContactUsRequest
	if err := json.NewDecoder(r.Body).Decode(&contactReq); err != nil {
		handleContactError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Validate required fields
	if contactReq.Note == "" {
		handleContactError(w, http.StatusBadRequest, "Note is required", nil)
		return
	}

	// Convert request to contact model
	contact := models.ContactUs{
		FullName:    strings.TrimSpace(contactReq.FullName),
		CompanyName: strings.TrimSpace(contactReq.CompanyName),
		Note:        strings.TrimSpace(contactReq.Note),
		Email:       strings.ToLower(strings.TrimSpace(contactReq.Email)),
		ChatbotURL:  strings.TrimSpace(contactReq.ChatbotURL),
	}

	// Create contact message in database
	createdContact, err := h.DB.CreateContactUs(contact)
	if err != nil {
		handleContactError(w, http.StatusInternalServerError, "Failed to create contact message", err)
		return
	}

	// Send email notification
	go func() {
		err := email.SendContactEmail(createdContact)
		if err != nil {
			log.Printf("Failed to send contact email: %v", err)
		}
	}()

	response := models.ContactUsResponse{
		Success: true,
		Message: "Contact message sent successfully",
		Contact: &createdContact,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetContactMessage handles GET /api/v1/contact/{id} - Get a specific contact message
func (h *ContactHandler) GetContactMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract contact ID from URL path
	contactID := strings.TrimPrefix(r.URL.Path, "/api/v1/contact/")
	if contactID == "" {
		handleContactError(w, http.StatusBadRequest, "Contact ID is required", nil)
		return
	}

	// Get contact message from database
	contact, err := h.DB.GetContactUs(contactID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleContactError(w, http.StatusNotFound, "Contact message not found", err)
		} else {
			handleContactError(w, http.StatusInternalServerError, "Failed to get contact message", err)
		}
		return
	}

	response := models.ContactUsResponse{
		Success: true,
		Message: "Contact message retrieved successfully",
		Contact: &contact,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllContactMessages handles GET /api/v1/contacts - Get all contact messages (admin only)
func (h *ContactHandler) GetAllContactMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all contact messages from database
	contacts, err := h.DB.GetAllContactUs()
	if err != nil {
		handleContactError(w, http.StatusInternalServerError, "Failed to get contact messages", err)
		return
	}

	response := struct {
		Success  bool               `json:"success"`
		Message  string             `json:"message"`
		Contacts []models.ContactUs `json:"contacts"`
	}{
		Success:  true,
		Message:  "Contact messages retrieved successfully",
		Contacts: contacts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleContactError sends an error response for contact operations
func handleContactError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("Contact handler error: %s - %v", message, err)

	response := models.ContactUsResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
