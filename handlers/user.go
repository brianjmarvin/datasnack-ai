package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"datasnack/db"
	"datasnack/middleware"
	"datasnack/models"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	DB *db.Database
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(database *db.Database) *UserHandler {
	return &UserHandler{
		DB: database,
	}
}

// CreateUser handles POST /users - Create a new user
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	userClient, ok := middleware.GetUserFromContext(r)
	if !ok {
		handleUserError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var userReq models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		handleUserError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Validate required fields
	if userReq.Email == "" {
		handleUserError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	// Use user ID from authenticated context
	userID := userClient.UserID

	// Set default values
	user := models.User{
		UserID:           userID,
		FirstName:        userReq.FirstName,
		LastName:         userReq.LastName,
		FullName:         userReq.FullName,
		Email:            strings.ToLower(strings.TrimSpace(userReq.Email)),
		SubscriptionPlan: userReq.SubscriptionPlan,
		AccountValidated: userReq.AccountValidated,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		LastLogin:        time.Now(),
		CompanyID:        userReq.CompanyID,
		CompanyName:      userReq.CompanyName,
		Credits:          userReq.Credits,
		IsActive:         true,
	}

	// Set default subscription plan if not provided
	if user.SubscriptionPlan == "" {
		user.SubscriptionPlan = models.SPARK
	}

	// Set default credits if not provided
	if user.Credits == 0 {
		user.Credits = 20
	}

	// Create user in database
	createdUser, err := h.DB.CreateUser(user)
	if err != nil {
		handleUserError(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User created successfully",
		User:    &createdUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetUser handles GET /users/{id} - Get a specific user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	if userID == "" {
		handleUserError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Get user from database
	user, err := h.DB.GetUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to get user", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    &user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUser handles GET /users/me - Get current authenticated user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	userClient, ok := middleware.GetUserFromContext(r)
	if !ok {
		handleUserError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Get user from database
	user, err := h.DB.GetUser(userClient.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to get user", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    &user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAllUsers handles GET /users - Get all users (admin only)
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all users from database
	users, err := h.DB.GetAllUsers()
	if err != nil {
		handleUserError(w, http.StatusInternalServerError, "Failed to get users", err)
		return
	}

	response := models.UsersResponse{
		Success: true,
		Message: fmt.Sprintf("Retrieved %d users", len(users)),
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUser handles PUT /users/{id} - Update a specific user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	if userID == "" {
		handleUserError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	var userReq models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		handleUserError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Convert request to user model
	user := models.User{
		FirstName:        userReq.FirstName,
		LastName:         userReq.LastName,
		FullName:         userReq.FullName,
		Email:            strings.ToLower(strings.TrimSpace(userReq.Email)),
		SubscriptionPlan: userReq.SubscriptionPlan,
		AccountValidated: userReq.AccountValidated,
		CompanyID:        userReq.CompanyID,
		CompanyName:      userReq.CompanyName,
		Credits:          userReq.Credits,
		IsActive:         userReq.IsActive,
	}

	// Update user in database
	updatedUser, err := h.DB.UpdateUser(userID, user)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to update user", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    &updatedUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCurrentUser handles PUT /users/me - Update current authenticated user
func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	userClient, ok := middleware.GetUserFromContext(r)
	if !ok {
		handleUserError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var userReq models.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		handleUserError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Convert request to user model
	user := models.User{
		FirstName:        userReq.FirstName,
		LastName:         userReq.LastName,
		FullName:         userReq.FullName,
		Email:            strings.ToLower(strings.TrimSpace(userReq.Email)),
		SubscriptionPlan: userReq.SubscriptionPlan,
		AccountValidated: userReq.AccountValidated,
		CompanyID:        userReq.CompanyID,
		CompanyName:      userReq.CompanyName,
		Credits:          userReq.Credits,
		IsActive:         userReq.IsActive,
	}

	// Update user in database
	updatedUser, err := h.DB.UpdateUser(userClient.UserID, user)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to update user", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    &updatedUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE /users/{id} - Delete a specific user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	if userID == "" {
		handleUserError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Delete user from database
	err := h.DB.DeleteUser(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to delete user", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateUserCredits handles PUT /users/{id}/credits - Update user credits
func (h *UserHandler) UpdateUserCredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	userID = strings.TrimSuffix(userID, "/credits")
	if userID == "" {
		handleUserError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		handleUserError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}

	var creditUpdate struct {
		Credits int `json:"credits"`
	}
	if err := json.Unmarshal(body, &creditUpdate); err != nil {
		handleUserError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Update user credits
	err = h.DB.UpdateUserCredits(userID, creditUpdate.Credits)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to update user credits", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: fmt.Sprintf("User credits updated by %d", creditUpdate.Credits),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateCurrentUserCredits handles PUT /users/me/credits - Update current user's credits
func (h *UserHandler) UpdateCurrentUserCredits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	userClient, ok := middleware.GetUserFromContext(r)
	if !ok {
		handleUserError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		handleUserError(w, http.StatusBadRequest, "Failed to read request body", err)
		return
	}

	var creditUpdate struct {
		Credits int `json:"credits"`
	}
	if err := json.Unmarshal(body, &creditUpdate); err != nil {
		handleUserError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	// Update user credits
	err = h.DB.UpdateUserCredits(userClient.UserID, creditUpdate.Credits)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			handleUserError(w, http.StatusNotFound, "User not found", err)
		} else {
			handleUserError(w, http.StatusInternalServerError, "Failed to update user credits", err)
		}
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: fmt.Sprintf("User credits updated by %d", creditUpdate.Credits),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUserError sends an error response for user operations
func handleUserError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("User handler error: %s - %v", message, err)

	response := models.UserResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
