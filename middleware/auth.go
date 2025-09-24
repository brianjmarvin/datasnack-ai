package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"datasnack/models"

	firebase "firebase.google.com/go/v4"
)

type ContextKey string

const ContextUserKey ContextKey = "user"

// Global auth service instance
var authService *AuthService

// AuthService holds the Firebase app for authentication
type AuthService struct {
	App *firebase.App
}

// NewAuthService creates a new authentication service
func NewAuthService() (*AuthService, error) {
	// Set the Google Application Credentials if not already set
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "datasnack-firebase-adminsdk-9i81w-c2064ff92c.json")
	}

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	return &AuthService{
		App: app,
	}, nil
}

// InitAuthService initializes the global auth service
func InitAuthService() error {
	service, err := NewAuthService()
	if err != nil {
		return err
	}
	authService = service
	return nil
}

// checkWebsiteAuth verifies Firebase ID token and returns user client
func (s *AuthService) checkWebsiteAuth(token string) (models.UserClient, error) {
	var userClient models.UserClient
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) < 2 {
		log.Println("token does not exist")
		return models.UserClient{}, fmt.Errorf("token does not exist")
	}
	reqToken := splitToken[1]

	auth, err := s.App.Auth(context.Background())
	if err != nil {
		log.Println(err)
		return models.UserClient{}, err
	}

	result, err := auth.VerifyIDToken(context.Background(), reqToken)
	if err != nil {
		log.Println(err)
		return models.UserClient{}, err
	}

	userClient.UserID = result.Claims["user_id"].(string)
	userClient.IsAuthenticated = true
	// log.Printf("user %v is authenticated %v", userClient.UserID, userClient.IsAuthenticated)
	return userClient, nil
}

// AuthMiddleware provides basic authentication for protected endpoints
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AuthMiddleware: %s %s", r.Method, r.URL.Path)

		var userClient models.UserClient
		apiKeyWithBearer := r.Header.Get("Authorization")

		// Remove "Bearer " prefix if present
		apiKey := strings.ReplaceAll(apiKeyWithBearer, "Bearer ", "")

		if apiKey == "" {
			log.Println("No authorization header provided")
			handleAuthError(w, "Authorization header required")
			return
		}

		// Check for admin auth
		adminKey := os.Getenv("ADMIN_API_KEY")
		if adminKey != "" && apiKey == adminKey {
			userClient.UserID = os.Getenv("ADMIN_USER_ID")
			if userClient.UserID == "" {
				userClient.UserID = "admin"
			}
			userClient.IsAuthenticated = true
			ctx := context.WithValue(r.Context(), ContextUserKey, userClient)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Check for demo auth
		if apiKey == "DEMO" {
			userClient.UserID = "demo-user"
			userClient.IsAuthenticated = true
			userClient.IsDemo = true
			ctx := context.WithValue(r.Context(), ContextUserKey, userClient)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Verify Firebase token
		if authService == nil {
			log.Println("Auth service not initialized")
			handleAuthError(w, "Authentication service not available")
			return
		}

		token := apiKeyWithBearer
		result, err := authService.checkWebsiteAuth(token)
		if err != nil {
			log.Println("Firebase token verification failed:", err)
			handleAuthError(w, "Invalid authentication token")
			return
		}

		userClient = result
		ctx := context.WithValue(r.Context(), ContextUserKey, userClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// PublicMiddleware allows public access without authentication
func PublicMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("PublicMiddleware: %s %s", r.Method, r.URL.Path)

		var userClient models.UserClient

		// Check if user is authenticated
		apiKeyWithBearer := r.Header.Get("Authorization")
		apiKey := strings.ReplaceAll(apiKeyWithBearer, "Bearer ", "")

		if apiKey != "" {
			// User is authenticated
			userClient.UserID = apiKey
			userClient.IsAuthenticated = true
		} else {
			// Public access
			userClient.UserID = "anonymous"
			userClient.IsAuthenticated = false
		}

		ctx := context.WithValue(r.Context(), ContextUserKey, userClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminMiddleware provides admin-only access
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("AdminMiddleware: %s %s", r.Method, r.URL.Path)

		var userClient models.UserClient
		apiKeyWithBearer := r.Header.Get("Authorization")
		apiKey := strings.ReplaceAll(apiKeyWithBearer, "Bearer ", "")

		if apiKey == "" {
			log.Println("No authorization header provided for admin endpoint")
			handleAuthError(w, "Admin authorization required")
			return
		}

		// Check for admin auth
		adminKey := os.Getenv("ADMIN_API_KEY")
		if adminKey == "" {
			adminKey = "admin-secret-key" // Default admin key for development
		}

		if apiKey != adminKey {
			log.Println("Invalid admin authorization")
			handleAuthError(w, "Admin authorization required")
			return
		}

		userClient.UserID = "admin"
		userClient.IsAuthenticated = true

		ctx := context.WithValue(r.Context(), ContextUserKey, userClient)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext extracts the user from the request context
func GetUserFromContext(r *http.Request) (models.UserClient, bool) {
	user, ok := r.Context().Value(ContextUserKey).(models.UserClient)
	return user, ok
}

// handleAuthError sends an authentication error response
func handleAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]string{
		"error":   "Unauthorized",
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// CORS middleware for handling cross-origin requests
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Define allowed origins
		const WEBSITE_URL1 = "https://datasnack.ai"
		const WEBSITE_URL2 = "http://localhost:5173"

		// Get the origin from the request
		origin := r.Header.Get("Origin")

		// Check if the origin is allowed
		var allowedOrigin string
		if origin == WEBSITE_URL1 || origin == WEBSITE_URL2 {
			allowedOrigin = origin
		} else {
			// For development, you might want to allow localhost on different ports
			// Uncomment the following lines if you need more flexibility in development
			// if strings.HasPrefix(origin, "http://localhost:") {
			//     allowedOrigin = origin
			// }
		}

		// Set CORS headers
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}
