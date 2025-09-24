package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"datasnack/cloneAttack"
	"datasnack/db"
	"datasnack/handlers"
	"datasnack/middleware"

	"github.com/spf13/cobra"
)

// EvaluationRequest represents the request payload for the evaluation API
type EvaluationRequest struct {
	AgentConfig    AgentConfigRequest    `json:"agentConfig"`
	EndpointConfig EndpointConfigRequest `json:"endpointConfig"`
}

// AgentConfigRequest represents the agent configuration in the API request
type AgentConfigRequest struct {
	BaseURL           string                   `json:"baseURL"`
	Endpoint          string                   `json:"endpoint"`
	AgentRootFolder   string                   `json:"agentRootFolder"`
	TrackingEnabled   bool                     `json:"trackingEnabled"`
	AgentPurpose      string                   `json:"agentPurpose"`
	TestConfiguration TestConfigurationRequest `json:"testConfiguration"`
}

// TestConfigurationRequest represents the test configuration in the API request
type TestConfigurationRequest struct {
	DataLeakageTests     int `json:"dataLeakageTests"`
	PromptInjectionTests int `json:"promptInjectionTests"`
	ConsistencyTests     int `json:"consistencyTests"`
	IterationsPerTest    int `json:"iterationsPerTest"`
}

// EndpointConfigRequest represents the endpoint configuration in the API request
type EndpointConfigRequest struct {
	Service   ServiceConfigRequest       `json:"service"`
	Endpoints map[string]EndpointDetails `json:"endpoints"`
}

// ServiceConfigRequest represents the service configuration in the API request
type ServiceConfigRequest struct {
	Name        string `json:"name"`
	BaseURL     string `json:"baseURL"`
	Description string `json:"description"`
}

// EndpointDetails represents endpoint details
type EndpointDetails struct {
	Description   string                 `json:"description"`
	Method        string                 `json:"method"`
	Path          string                 `json:"path"`
	RequestSchema map[string]interface{} `json:"request_schema,omitempty"`
}

// EvaluationResponse represents the response from the evaluation API
type EvaluationResponse struct {
	Success       bool                           `json:"success"`
	Message       string                         `json:"message,omitempty"`
	Results       *cloneAttack.StressTestResults `json:"results,omitempty"`
	CallHistory   []cloneAttack.CallMetadata     `json:"callHistory,omitempty"`
	Error         string                         `json:"error,omitempty"`
	ExecutionTime float64                        `json:"executionTime"`
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server [port]",
	Short: "Start HTTP server for API-based endpoint evaluation",
	Long: `Start an HTTP server that provides an API endpoint for running endpoint evaluations.

The server accepts POST requests to /evaluate with a JSON payload containing:
- agentConfig: Agent configuration (baseURL, endpoint, test settings, etc.)
- endpointConfig: Endpoint configuration (service info, endpoints, schemas)

Example request:
  POST /evaluate
  {
    "agentConfig": {
      "baseURL": "http://localhost:8080",
      "endpoint": "/api/v1/chatHandler",
      "agentRootFolder": "/path/to/agent",
      "trackingEnabled": true,
      "agentPurpose": "AI chat service",
      "testConfiguration": {
        "dataLeakageTests": 5,
        "promptInjectionTests": 5,
        "consistencyTests": 5,
        "iterationsPerTest": 3
      }
    },
    "endpointConfig": {
      "service": {
        "name": "My AI Service",
        "baseURL": "http://localhost:8080",
        "description": "AI chat service"
      },
      "endpoints": {
        "main": {
          "description": "Main chat endpoint",
          "method": "POST",
          "path": "/api/v1/chatHandler",
          "request_schema": { ... }
        }
      }
    }
  }

The server returns the evaluation results as JSON.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		port := "8080"
		if len(args) > 0 {
			port = args[0]
		}

		// Initialize database
		database, err := db.Start()
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer database.Close()

		// Create users table
		if err := database.CreateUsersTable(); err != nil {
			log.Fatalf("Failed to create users table: %v", err)
		}

		// Create contact_us table
		if err := database.CreateContactUsTable(); err != nil {
			log.Fatalf("Failed to create contact_us table: %v", err)
		}

		// Initialize Firebase authentication service
		if err := middleware.InitAuthService(); err != nil {
			log.Fatalf("Failed to initialize authentication service: %v", err)
		}
		log.Println("Firebase authentication service initialized")

		// Initialize handlers
		userHandler := handlers.NewUserHandler(database)
		contactHandler := handlers.NewContactHandler(database)

		// Set up routes with middleware
		http.HandleFunc("/evaluate", middleware.CORSMiddleware(handleEvaluation))
		http.HandleFunc("/health", middleware.CORSMiddleware(handleHealth))

		// Contact endpoints (public - no authentication required)
		http.HandleFunc("/api/v1/sendNote", middleware.CORSMiddleware(middleware.PublicMiddleware(contactHandler.SendNote)))
		http.HandleFunc("/api/v1/contact/", middleware.CORSMiddleware(middleware.AdminMiddleware(contactHandler.GetContactMessage)))
		http.HandleFunc("/api/v1/contacts", middleware.CORSMiddleware(middleware.AdminMiddleware(contactHandler.GetAllContactMessages)))

		// User management endpoints
		http.HandleFunc("/users", middleware.CORSMiddleware(middleware.AuthMiddleware(userHandler.CreateUser)))
		http.HandleFunc("/users/", middleware.CORSMiddleware(middleware.AuthMiddleware(userHandler.GetUser)))
		http.HandleFunc("/users/me", middleware.CORSMiddleware(middleware.AuthMiddleware(userHandler.GetCurrentUser)))
		http.HandleFunc("/users/all", middleware.CORSMiddleware(middleware.AdminMiddleware(userHandler.GetAllUsers)))

		// User update endpoints
		http.HandleFunc("/users/update/", middleware.CORSMiddleware(middleware.AuthMiddleware(userHandler.UpdateUser)))
		http.HandleFunc("/users/me/update", middleware.CORSMiddleware(middleware.AuthMiddleware(userHandler.UpdateCurrentUser)))
		http.HandleFunc("/users/delete/", middleware.CORSMiddleware(middleware.AdminMiddleware(userHandler.DeleteUser)))
		http.HandleFunc("/users/credits/", middleware.CORSMiddleware(middleware.AdminMiddleware(userHandler.UpdateUserCredits)))

		log.Printf("Starting server on port %s", port)
		log.Printf("API endpoints:")
		log.Printf("  - Evaluation: http://localhost:%s/evaluate", port)
		log.Printf("  - Health check: http://localhost:%s/health", port)
		log.Printf("  - Contact form: http://localhost:%s/api/v1/sendNote", port)
		log.Printf("  - User management: http://localhost:%s/users", port)
		log.Printf("  - Current user: http://localhost:%s/users/me", port)
		log.Printf("  - All users (admin): http://localhost:%s/users/all", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	},
}

// handleHealth handles health check requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// handleEvaluation handles evaluation requests
func handleEvaluation(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req EvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := EvaluationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if err := validateRequest(&req); err != nil {
		response := EvaluationResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert request to internal structures
	agentConfig := convertAgentConfig(&req.AgentConfig)
	endpointConfig := convertEndpointConfig(&req.EndpointConfig)

	// Initialize AI client
	ai, err := initializeAIClient()
	if err != nil {
		response := EvaluationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to initialize AI client: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create endpoint evaluator with the converted endpoint configuration
	evaluator, err := cloneAttack.NewEndpointEvaluatorWithConfig(
		ai,
		endpointConfig,
		agentConfig.AgentPurpose,
		cloneAttack.TestConfiguration{
			DataLeakageTests:     agentConfig.TestConfiguration.DataLeakageTests,
			PromptInjectionTests: agentConfig.TestConfiguration.PromptInjectionTests,
			ConsistencyTests:     agentConfig.TestConfiguration.ConsistencyTests,
			IterationsPerTest:    agentConfig.TestConfiguration.IterationsPerTest,
		},
	)
	if err != nil {
		response := EvaluationResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create evaluator: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set the endpoint configuration directly
	evaluator.SetEndpointConfig(endpointConfig)

	// Run evaluation
	log.Println("Starting endpoint evaluation...")
	results, err := evaluator.RunComprehensiveVulnerabilityTest()
	if err != nil {
		response := EvaluationResponse{
			Success: false,
			Error:   fmt.Sprintf("Evaluation failed: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	executionTime := time.Since(startTime).Seconds()

	// Return successful response
	response := EvaluationResponse{
		Success:       true,
		Message:       "Evaluation completed successfully",
		Results:       results,
		CallHistory:   evaluator.GetCallHistory(),
		ExecutionTime: executionTime,
	}

	log.Printf("Evaluation completed in %.2f seconds", executionTime)
	json.NewEncoder(w).Encode(response)
}

// validateRequest validates the evaluation request
func validateRequest(req *EvaluationRequest) error {
	if req.AgentConfig.BaseURL == "" {
		return fmt.Errorf("agentConfig.baseURL is required")
	}
	if req.AgentConfig.Endpoint == "" {
		return fmt.Errorf("agentConfig.endpoint is required")
	}
	if req.AgentConfig.AgentPurpose == "" {
		return fmt.Errorf("agentConfig.agentPurpose is required")
	}
	if req.EndpointConfig.Service.BaseURL == "" {
		return fmt.Errorf("endpointConfig.service.baseURL is required")
	}
	if len(req.EndpointConfig.Endpoints) == 0 {
		return fmt.Errorf("endpointConfig.endpoints is required")
	}
	return nil
}

// convertAgentConfig converts API request agent config to internal format
func convertAgentConfig(req *AgentConfigRequest) struct {
	BaseURL           string
	Endpoint          string
	AgentRootFolder   string
	TrackingEnabled   bool
	AgentPurpose      string
	TestConfiguration struct {
		DataLeakageTests     int
		PromptInjectionTests int
		ConsistencyTests     int
		IterationsPerTest    int
	}
} {
	return struct {
		BaseURL           string
		Endpoint          string
		AgentRootFolder   string
		TrackingEnabled   bool
		AgentPurpose      string
		TestConfiguration struct {
			DataLeakageTests     int
			PromptInjectionTests int
			ConsistencyTests     int
			IterationsPerTest    int
		}
	}{
		BaseURL:         req.BaseURL,
		Endpoint:        req.Endpoint,
		AgentRootFolder: req.AgentRootFolder,
		TrackingEnabled: req.TrackingEnabled,
		AgentPurpose:    req.AgentPurpose,
		TestConfiguration: struct {
			DataLeakageTests     int
			PromptInjectionTests int
			ConsistencyTests     int
			IterationsPerTest    int
		}{
			DataLeakageTests:     req.TestConfiguration.DataLeakageTests,
			PromptInjectionTests: req.TestConfiguration.PromptInjectionTests,
			ConsistencyTests:     req.TestConfiguration.ConsistencyTests,
			IterationsPerTest:    req.TestConfiguration.IterationsPerTest,
		},
	}
}

// convertEndpointConfig converts API request endpoint config to internal format
func convertEndpointConfig(req *EndpointConfigRequest) *cloneAttack.EndpointConfig {
	endpoints := make(map[string]cloneAttack.EndpointInfo)
	for name, details := range req.Endpoints {
		endpoints[name] = cloneAttack.EndpointInfo{
			Description:   details.Description,
			Method:        details.Method,
			Path:          details.Path,
			RequestSchema: details.RequestSchema,
		}
	}

	return &cloneAttack.EndpointConfig{
		Service: cloneAttack.ServiceConfig{
			Name:        req.Service.Name,
			BaseURL:     req.Service.BaseURL,
			Description: req.Service.Description,
		},
		Endpoints: endpoints,
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
