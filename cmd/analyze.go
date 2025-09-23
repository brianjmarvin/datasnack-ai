/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"datasnack/cloneAttack"

	"github.com/spf13/cobra"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [output-directory]",
	Short: "Analyze a target AI agent and generate endpoint configuration",
	Long: `Analyze a target AI agent (local location, endpoint, and base URL provided in agentConfig.json file) 
and generate an endpoint-config.json file that contains the endpoint, request schema, and response schema 
that can be used to test the AI agent with the "endpointEval" command.

The command analyzes agents written in JavaScript, Go, or Python by:
- Reading agent configuration from config/agentConfig.json
- Using static analysis and OS tools to examine the codebase
- Using AI to understand the API structure and generate schemas
- Creating a complete endpoint-config.json file with timestamp

The output file will be named: endpoint_config_YYYYMMDD_HHMMSS.json

Example:
  datasnack analyze config/`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDirectory := args[0]

		// Read agent configuration
		log.Println("Reading agent configuration from: config/agentConfig.json")
		agentConfigData, err := os.ReadFile("config/agentConfig.json")
		if err != nil {
			log.Fatalf("Failed to read agent config: %v", err)
		}

		var agentConfig struct {
			BaseURL         string `json:"baseURL"`
			Endpoint        string `json:"endpoint"`
			AgentRootFolder string `json:"agentRootFolder"`
			TrackingEnabled bool   `json:"trackingEnabled"`
			AgentPurpose    string `json:"agentPurpose"`
		}

		if err := json.Unmarshal(agentConfigData, &agentConfig); err != nil {
			log.Fatalf("Failed to parse agent config: %v", err)
		}

		// Initialize AI client
		ai, err := initializeAIClient()
		if err != nil {
			log.Fatalf("Failed to initialize AI client: %v", err)
		}

		// Create agent analyzer
		analyzer := NewAgentAnalyzer(ai, agentConfig.AgentRootFolder, agentConfig.BaseURL, agentConfig.Endpoint, agentConfig.AgentPurpose)

		// Analyze the agent
		log.Println("Analyzing agent codebase...")
		analysis, err := analyzer.AnalyzeAgent()
		if err != nil {
			log.Fatalf("Failed to analyze agent: %v", err)
		}

		// Generate endpoint configuration
		log.Println("Generating endpoint configuration...")
		endpointConfig, err := analyzer.GenerateEndpointConfig(analysis)
		if err != nil {
			log.Fatalf("Failed to generate endpoint config: %v", err)
		}

		// Generate timestamped filename
		timestamp := time.Now().Format("20060102_150405")
		outputConfigFile := filepath.Join(outputDirectory, fmt.Sprintf("endpoint_config_%s.json", timestamp))

		// Write the configuration to file
		configJSON, err := json.MarshalIndent(endpointConfig, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal config to JSON: %v", err)
		}

		if err := os.WriteFile(outputConfigFile, configJSON, 0644); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}

		log.Printf("Endpoint configuration generated successfully: %s", outputConfigFile)
		log.Printf("You can now use: datasnack endpointEval %s", outputConfigFile)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

// AgentAnalysis represents the analysis results of an AI agent
type AgentAnalysis struct {
	Language        string                 `json:"language"`
	Framework       string                 `json:"framework"`
	Endpoints       []EndpointInfo         `json:"endpoints"`
	RequestSchemas  map[string]interface{} `json:"requestSchemas"`
	ResponseSchemas map[string]interface{} `json:"responseSchemas"`
	HealthEndpoint  string                 `json:"healthEndpoint"`
	MainFiles       []string               `json:"mainFiles"`
	PackageInfo     map[string]interface{} `json:"packageInfo"`
}

// EndpointInfo represents information about an API endpoint
type EndpointInfo struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Handler     string `json:"handler"`
}

// SchemaHint represents extracted schema information from static analysis
type SchemaHint struct {
	FieldName     string
	FieldType     string
	IsRequired    bool
	Description   string
	Example       string
	Constraints   map[string]interface{}
	IsArray       bool
	ArrayItemType string
	NestedFields  map[string]SchemaHint
}

// AgentAnalyzer handles the analysis of AI agents
type AgentAnalyzer struct {
	ai              cloneAttack.AIClient
	agentRootFolder string
	baseURL         string
	endpoint        string
	agentPurpose    string
}

// NewAgentAnalyzer creates a new agent analyzer
func NewAgentAnalyzer(ai cloneAttack.AIClient, agentRootFolder, baseURL, endpoint, agentPurpose string) *AgentAnalyzer {
	return &AgentAnalyzer{
		ai:              ai,
		agentRootFolder: agentRootFolder,
		baseURL:         baseURL,
		endpoint:        endpoint,
		agentPurpose:    agentPurpose,
	}
}

// AnalyzeAgent performs comprehensive analysis of the agent codebase
func (a *AgentAnalyzer) AnalyzeAgent() (*AgentAnalysis, error) {
	analysis := &AgentAnalysis{
		Endpoints:       []EndpointInfo{},
		RequestSchemas:  make(map[string]interface{}),
		ResponseSchemas: make(map[string]interface{}),
		MainFiles:       []string{},
		PackageInfo:     make(map[string]interface{}),
	}

	// Detect programming language
	language, err := a.detectLanguage()
	if err != nil {
		return nil, fmt.Errorf("failed to detect language: %w", err)
	}
	analysis.Language = language

	// Detect framework
	framework, err := a.detectFramework(language)
	if err != nil {
		return nil, fmt.Errorf("failed to detect framework: %w", err)
	}
	analysis.Framework = framework

	// Find main files
	mainFiles, err := a.findMainFiles(language)
	if err != nil {
		return nil, fmt.Errorf("failed to find main files: %w", err)
	}
	analysis.MainFiles = mainFiles

	// Extract package information
	packageInfo, err := a.extractPackageInfo(language)
	if err != nil {
		log.Printf("Warning: Failed to extract package info: %v", err)
	}
	analysis.PackageInfo = packageInfo

	// Find endpoints using static analysis
	endpoints, err := a.findEndpoints(language, framework)
	if err != nil {
		return nil, fmt.Errorf("failed to find endpoints: %w", err)
	}
	analysis.Endpoints = endpoints

	// Find health endpoint
	healthEndpoint, err := a.findHealthEndpoint(language, framework)
	if err != nil {
		log.Printf("Warning: Failed to find health endpoint: %v", err)
	}
	analysis.HealthEndpoint = healthEndpoint

	// Use AI to analyze and generate schemas
	if err := a.generateSchemasWithAI(analysis); err != nil {
		return nil, fmt.Errorf("failed to generate schemas with AI: %w", err)
	}

	return analysis, nil
}

// detectLanguage detects the programming language of the agent
func (a *AgentAnalyzer) detectLanguage() (string, error) {
	// Check for common files that indicate language
	languageFiles := map[string][]string{
		"javascript": {"package.json", "*.js", "*.ts", "*.jsx", "*.tsx"},
		"go":         {"go.mod", "*.go"},
		"python":     {"requirements.txt", "pyproject.toml", "setup.py", "*.py"},
	}

	for lang, patterns := range languageFiles {
		for _, pattern := range patterns {
			if pattern[0] == '*' {
				// Use find command for file patterns
				cmd := exec.Command("find", a.agentRootFolder, "-name", pattern, "-type", "f")
				output, err := cmd.Output()
				if err == nil && len(output) > 0 {
					return lang, nil
				}
			} else {
				// Check for specific files
				filePath := filepath.Join(a.agentRootFolder, pattern)
				if _, err := os.Stat(filePath); err == nil {
					return lang, nil
				}
			}
		}
	}

	return "unknown", fmt.Errorf("could not detect programming language")
}

// detectFramework detects the framework used by the agent
func (a *AgentAnalyzer) detectFramework(language string) (string, error) {
	switch language {
	case "javascript":
		return a.detectJavaScriptFramework()
	case "go":
		return a.detectGoFramework()
	case "python":
		return a.detectPythonFramework()
	default:
		return "unknown", nil
	}
}

// detectJavaScriptFramework detects JavaScript framework
func (a *AgentAnalyzer) detectJavaScriptFramework() (string, error) {
	packageJsonPath := filepath.Join(a.agentRootFolder, "package.json")
	if _, err := os.Stat(packageJsonPath); err != nil {
		return "node", nil
	}

	packageData, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return "node", nil
	}

	var packageJson map[string]interface{}
	if err := json.Unmarshal(packageData, &packageJson); err != nil {
		return "node", nil
	}

	dependencies := make(map[string]interface{})
	if deps, ok := packageJson["dependencies"].(map[string]interface{}); ok {
		for k, v := range deps {
			dependencies[k] = v
		}
	}
	if devDeps, ok := packageJson["devDependencies"].(map[string]interface{}); ok {
		for k, v := range devDeps {
			dependencies[k] = v
		}
	}

	// Check for common frameworks
	frameworkChecks := map[string][]string{
		"express": {"express"},
		"fastify": {"fastify"},
		"koa":     {"koa"},
		"hapi":    {"@hapi/hapi", "hapi"},
		"next":    {"next"},
		"nuxt":    {"nuxt"},
		"nest":    {"@nestjs/core"},
	}

	for framework, packages := range frameworkChecks {
		for _, pkg := range packages {
			if _, exists := dependencies[pkg]; exists {
				return framework, nil
			}
		}
	}

	return "node", nil
}

// detectGoFramework detects Go framework
func (a *AgentAnalyzer) detectGoFramework() (string, error) {
	// Check go.mod for common frameworks
	goModPath := filepath.Join(a.agentRootFolder, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		return "net/http", nil
	}

	goModData, err := os.ReadFile(goModPath)
	if err != nil {
		return "net/http", nil
	}

	content := string(goModData)
	frameworkChecks := map[string]string{
		"gin":         "github.com/gin-gonic/gin",
		"echo":        "github.com/labstack/echo",
		"fiber":       "github.com/gofiber/fiber",
		"chi":         "github.com/go-chi/chi",
		"gorilla/mux": "github.com/gorilla/mux",
		"httprouter":  "github.com/julienschmidt/httprouter",
	}

	for framework, importPath := range frameworkChecks {
		if strings.Contains(content, importPath) {
			return framework, nil
		}
	}

	return "net/http", nil
}

// detectPythonFramework detects Python framework
func (a *AgentAnalyzer) detectPythonFramework() (string, error) {
	// Check requirements.txt or pyproject.toml
	requirementsPath := filepath.Join(a.agentRootFolder, "requirements.txt")
	pyprojectPath := filepath.Join(a.agentRootFolder, "pyproject.toml")

	var content string
	if _, err := os.Stat(requirementsPath); err == nil {
		data, err := os.ReadFile(requirementsPath)
		if err == nil {
			content = string(data)
		}
	} else if _, err := os.Stat(pyprojectPath); err == nil {
		data, err := os.ReadFile(pyprojectPath)
		if err == nil {
			content = string(data)
		}
	}

	frameworkChecks := map[string][]string{
		"fastapi":   {"fastapi", "uvicorn"},
		"flask":     {"flask"},
		"django":    {"django"},
		"tornado":   {"tornado"},
		"aiohttp":   {"aiohttp"},
		"starlette": {"starlette"},
		"quart":     {"quart"},
	}

	for framework, packages := range frameworkChecks {
		for _, pkg := range packages {
			if strings.Contains(strings.ToLower(content), pkg) {
				return framework, nil
			}
		}
	}

	return "wsgi", nil
}

// findMainFiles finds the main entry point files
func (a *AgentAnalyzer) findMainFiles(language string) ([]string, error) {
	var mainFiles []string

	switch language {
	case "javascript":
		// Look for common entry points
		entryPoints := []string{"index.js", "app.js", "server.js", "main.js", "src/index.js", "src/app.js"}
		for _, entry := range entryPoints {
			filePath := filepath.Join(a.agentRootFolder, entry)
			if _, err := os.Stat(filePath); err == nil {
				mainFiles = append(mainFiles, entry)
			}
		}
	case "go":
		// Look for main.go files
		cmd := exec.Command("find", a.agentRootFolder, "-name", "main.go", "-type", "f")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					relPath, _ := filepath.Rel(a.agentRootFolder, line)
					mainFiles = append(mainFiles, relPath)
				}
			}
		}
	case "python":
		// Look for common entry points
		entryPoints := []string{"main.py", "app.py", "server.py", "run.py", "wsgi.py", "asgi.py"}
		for _, entry := range entryPoints {
			filePath := filepath.Join(a.agentRootFolder, entry)
			if _, err := os.Stat(filePath); err == nil {
				mainFiles = append(mainFiles, entry)
			}
		}
	}

	return mainFiles, nil
}

// extractPackageInfo extracts package information
func (a *AgentAnalyzer) extractPackageInfo(language string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	switch language {
	case "javascript":
		packageJsonPath := filepath.Join(a.agentRootFolder, "package.json")
		if _, err := os.Stat(packageJsonPath); err == nil {
			data, err := os.ReadFile(packageJsonPath)
			if err == nil {
				var packageJson map[string]interface{}
				if err := json.Unmarshal(data, &packageJson); err == nil {
					info["name"] = packageJson["name"]
					info["version"] = packageJson["version"]
					info["description"] = packageJson["description"]
				}
			}
		}
	case "go":
		goModPath := filepath.Join(a.agentRootFolder, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			data, err := os.ReadFile(goModPath)
			if err == nil {
				content := string(data)
				lines := strings.Split(content, "\n")
				if len(lines) > 0 {
					// Extract module name from first line
					moduleLine := strings.TrimSpace(lines[0])
					if strings.HasPrefix(moduleLine, "module ") {
						info["module"] = strings.TrimSpace(strings.TrimPrefix(moduleLine, "module"))
					}
				}
			}
		}
	case "python":
		// Try pyproject.toml first, then setup.py
		pyprojectPath := filepath.Join(a.agentRootFolder, "pyproject.toml")
		if _, err := os.Stat(pyprojectPath); err == nil {
			data, err := os.ReadFile(pyprojectPath)
			if err == nil {
				// Simple TOML parsing for name and version
				content := string(data)
				if nameMatch := regexp.MustCompile(`name\s*=\s*["']([^"']+)["']`).FindStringSubmatch(content); len(nameMatch) > 1 {
					info["name"] = nameMatch[1]
				}
				if versionMatch := regexp.MustCompile(`version\s*=\s*["']([^"']+)["']`).FindStringSubmatch(content); len(versionMatch) > 1 {
					info["version"] = versionMatch[1]
				}
			}
		}
	}

	return info, nil
}

// findEndpoints finds API endpoints using static analysis
func (a *AgentAnalyzer) findEndpoints(language, framework string) ([]EndpointInfo, error) {
	var endpoints []EndpointInfo

	// Always include the main endpoint from config
	endpoints = append(endpoints, EndpointInfo{
		Path:        a.endpoint,
		Method:      "POST", // Default to POST for AI agents
		Description: "Main AI agent endpoint",
		Handler:     "main",
	})

	// Use grep to find additional endpoints
	switch language {
	case "javascript":
		endpoints = append(endpoints, a.findJavaScriptEndpoints(framework)...)
	case "go":
		endpoints = append(endpoints, a.findGoEndpoints(framework)...)
	case "python":
		endpoints = append(endpoints, a.findPythonEndpoints(framework)...)
	}

	return endpoints, nil
}

// findJavaScriptEndpoints finds JavaScript endpoints
func (a *AgentAnalyzer) findJavaScriptEndpoints(framework string) []EndpointInfo {
	var endpoints []EndpointInfo

	// Common patterns for different frameworks
	patterns := map[string][]string{
		"express": {
			`app\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
			`router\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
		},
		"fastify": {
			`fastify\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
		},
		"koa": {
			`router\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
		},
	}

	frameworkPatterns, exists := patterns[framework]
	if !exists {
		frameworkPatterns = patterns["express"] // Default to express patterns
	}

	for _, pattern := range frameworkPatterns {
		cmd := exec.Command("grep", "-r", "-E", pattern, a.agentRootFolder)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					// Parse the grep output to extract endpoint info
					if endpoint := a.parseJavaScriptEndpoint(line, pattern); endpoint != nil {
						endpoints = append(endpoints, *endpoint)
					}
				}
			}
		}
	}

	return endpoints
}

// findGoEndpoints finds Go endpoints
func (a *AgentAnalyzer) findGoEndpoints(framework string) []EndpointInfo {
	var endpoints []EndpointInfo

	// Common patterns for different frameworks
	patterns := map[string][]string{
		"gin": {
			`router\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["']`,
			`r\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["']`,
		},
		"echo": {
			`e\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["']`,
		},
		"fiber": {
			`app\.(Get|Post|Put|Delete|Patch)\s*\(\s*["']([^"']+)["']`,
		},
		"net/http": {
			`http\.(HandleFunc|Handle)\s*\(\s*["']([^"']+)["']`,
		},
	}

	frameworkPatterns, exists := patterns[framework]
	if !exists {
		frameworkPatterns = patterns["gin"] // Default to gin patterns
	}

	for _, pattern := range frameworkPatterns {
		cmd := exec.Command("grep", "-r", "-E", pattern, a.agentRootFolder)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					if endpoint := a.parseGoEndpoint(line, pattern); endpoint != nil {
						endpoints = append(endpoints, *endpoint)
					}
				}
			}
		}
	}

	return endpoints
}

// findPythonEndpoints finds Python endpoints
func (a *AgentAnalyzer) findPythonEndpoints(framework string) []EndpointInfo {
	var endpoints []EndpointInfo

	// Common patterns for different frameworks
	patterns := map[string][]string{
		"fastapi": {
			`@app\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
			`@router\.(get|post|put|delete|patch)\s*\(\s*["']([^"']+)["']`,
		},
		"flask": {
			`@app\.route\s*\(\s*["']([^"']+)["']`,
		},
		"django": {
			`path\s*\(\s*["']([^"']+)["']`,
		},
	}

	frameworkPatterns, exists := patterns[framework]
	if !exists {
		frameworkPatterns = patterns["fastapi"] // Default to fastapi patterns
	}

	for _, pattern := range frameworkPatterns {
		cmd := exec.Command("grep", "-r", "-E", pattern, a.agentRootFolder)
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if line != "" {
					if endpoint := a.parsePythonEndpoint(line, pattern); endpoint != nil {
						endpoints = append(endpoints, *endpoint)
					}
				}
			}
		}
	}

	return endpoints
}

// parseJavaScriptEndpoint parses JavaScript endpoint from grep output
func (a *AgentAnalyzer) parseJavaScriptEndpoint(line, pattern string) *EndpointInfo {
	// Extract method and path from the line
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		method := strings.ToUpper(matches[1])
		path := matches[2]
		return &EndpointInfo{
			Path:        path,
			Method:      method,
			Description: fmt.Sprintf("Auto-detected %s endpoint", method),
			Handler:     "auto-detected",
		}
	}
	return nil
}

// parseGoEndpoint parses Go endpoint from grep output
func (a *AgentAnalyzer) parseGoEndpoint(line, pattern string) *EndpointInfo {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		method := strings.ToUpper(matches[1])
		path := matches[2]
		return &EndpointInfo{
			Path:        path,
			Method:      method,
			Description: fmt.Sprintf("Auto-detected %s endpoint", method),
			Handler:     "auto-detected",
		}
	}
	return nil
}

// parsePythonEndpoint parses Python endpoint from grep output
func (a *AgentAnalyzer) parsePythonEndpoint(line, pattern string) *EndpointInfo {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		method := strings.ToUpper(matches[1])
		path := matches[2]
		return &EndpointInfo{
			Path:        path,
			Method:      method,
			Description: fmt.Sprintf("Auto-detected %s endpoint", method),
			Handler:     "auto-detected",
		}
	} else if len(matches) == 2 {
		// For patterns that only capture path (like Flask)
		path := matches[1]
		return &EndpointInfo{
			Path:        path,
			Method:      "GET", // Default for Flask routes
			Description: "Auto-detected endpoint",
			Handler:     "auto-detected",
		}
	}
	return nil
}

// findHealthEndpoint finds health check endpoint
func (a *AgentAnalyzer) findHealthEndpoint(language, framework string) (string, error) {
	// Common health endpoint patterns
	healthPatterns := []string{
		"/health",
		"/healthz",
		"/ping",
		"/status",
		"/api/health",
		"/api/status",
	}

	// Use grep to find health endpoints
	pattern := strings.Join(healthPatterns, "|")
	cmd := exec.Command("grep", "-r", "-E", pattern, a.agentRootFolder)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line != "" {
				for _, healthPath := range healthPatterns {
					if strings.Contains(line, healthPath) {
						return healthPath, nil
					}
				}
			}
		}
	}

	return "/health", nil // Default health endpoint
}

// generateSchemasWithAI uses AI to analyze the code and generate request/response schemas
func (a *AgentAnalyzer) generateSchemasWithAI(analysis *AgentAnalysis) error {
	// First, perform static analysis to extract schema hints
	staticHints := a.extractSchemaHintsFromCode(analysis)

	// Read comprehensive code context for AI analysis
	codeContext := a.buildComprehensiveCodeContext(analysis, staticHints)

	// Generate request schema using AI with enhanced context
	requestSchema, err := a.generateRequestSchemaWithAI(codeContext, staticHints)
	if err != nil {
		log.Printf("Warning: Failed to generate request schema with AI: %v", err)
		requestSchema = a.generateFallbackRequestSchema(analysis, staticHints)
	}
	analysis.RequestSchemas["main"] = requestSchema

	// Generate response schema using AI with enhanced context
	responseSchema, err := a.generateResponseSchemaWithAI(codeContext, staticHints)
	if err != nil {
		log.Printf("Warning: Failed to generate response schema with AI: %v", err)
		responseSchema = a.generateFallbackResponseSchema(analysis, staticHints)
	}
	analysis.ResponseSchemas["main"] = responseSchema

	// Perform runtime schema discovery if endpoint is accessible
	if runtimeHints, err := a.performRuntimeSchemaDiscovery(); err == nil {
		a.mergeRuntimeHints(analysis, runtimeHints)
	}

	return nil
}

// GenerateEndpointConfig generates the final endpoint configuration YAML
func (a *AgentAnalyzer) GenerateEndpointConfig(analysis *AgentAnalysis) (map[string]interface{}, error) {
	// Get service name from package info or use default
	serviceName := "AI Agent Service"
	if name, ok := analysis.PackageInfo["name"]; ok {
		serviceName = fmt.Sprintf("%v", name)
	}

	// Get version from package info or use default
	version := "1.0.0"
	if ver, ok := analysis.PackageInfo["version"]; ok {
		version = fmt.Sprintf("%v", ver)
	}

	// Build the endpoint configuration
	config := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        serviceName,
			"version":     version,
			"description": a.agentPurpose,
			"base_url":    a.baseURL,
		},
		"endpoints": make(map[string]interface{}),
	}

	endpoints := config["endpoints"].(map[string]interface{})

	// Add health endpoint if found
	if analysis.HealthEndpoint != "" {
		endpoints["health"] = map[string]interface{}{
			"path":        analysis.HealthEndpoint,
			"method":      "GET",
			"description": "Health check endpoint",
		}
	}

	// Add main endpoint with schemas
	mainEndpoint := map[string]interface{}{
		"path":        a.endpoint,
		"method":      "POST",
		"description": "Main AI agent endpoint",
	}

	// Add request schema if available
	if requestSchema, ok := analysis.RequestSchemas["main"]; ok {
		mainEndpoint["request_schema"] = requestSchema
	}

	// Add response schema if available
	if responseSchema, ok := analysis.ResponseSchemas["main"]; ok {
		mainEndpoint["response_schema"] = responseSchema
	}

	endpoints["main"] = mainEndpoint

	// Add other discovered endpoints
	for i, endpoint := range analysis.Endpoints {
		if endpoint.Path != a.endpoint { // Skip main endpoint as it's already added
			endpointKey := fmt.Sprintf("endpoint_%d", i)
			endpoints[endpointKey] = map[string]interface{}{
				"path":        endpoint.Path,
				"method":      endpoint.Method,
				"description": endpoint.Description,
			}
		}
	}

	return config, nil
}

// extractSchemaHintsFromCode performs static analysis to extract schema information
func (a *AgentAnalyzer) extractSchemaHintsFromCode(analysis *AgentAnalysis) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	// Analyze main files for schema patterns
	for _, mainFile := range analysis.MainFiles {
		filePath := filepath.Join(a.agentRootFolder, mainFile)
		fileHints := a.analyzeFileForSchemaHints(filePath, analysis.Language, analysis.Framework)
		for k, v := range fileHints {
			hints[k] = v
		}
	}

	// Look for additional schema files
	schemaFiles := a.findSchemaFiles(analysis.Language)
	for _, schemaFile := range schemaFiles {
		filePath := filepath.Join(a.agentRootFolder, schemaFile)
		fileHints := a.analyzeSchemaFile(filePath, analysis.Language)
		for k, v := range fileHints {
			hints[k] = v
		}
	}

	return hints
}

// analyzeFileForSchemaHints analyzes a single file for schema patterns
func (a *AgentAnalyzer) analyzeFileForSchemaHints(filePath, language, framework string) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return hints
	}
	content := string(data)

	switch language {
	case "javascript":
		hints = a.extractJavaScriptSchemaHints(content, framework)
	case "go":
		hints = a.extractGoSchemaHints(content, framework)
	case "python":
		hints = a.extractPythonSchemaHints(content, framework)
	}

	return hints
}

// extractJavaScriptSchemaHints extracts schema hints from JavaScript code
func (a *AgentAnalyzer) extractJavaScriptSchemaHints(content, framework string) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	// Common patterns for different frameworks
	patterns := map[string][]string{
		"express": {
			`req\.body\.(\w+)`,                     // Express body parsing
			`req\.params\.(\w+)`,                   // URL parameters
			`req\.query\.(\w+)`,                    // Query parameters
			`const\s+(\w+)\s*=\s*req\.body\.(\w+)`, // Destructuring
		},
		"fastify": {
			`request\.body\.(\w+)`,
			`request\.params\.(\w+)`,
			`request\.query\.(\w+)`,
		},
		"koa": {
			`ctx\.request\.body\.(\w+)`,
			`ctx\.params\.(\w+)`,
			`ctx\.query\.(\w+)`,
		},
	}

	frameworkPatterns := patterns[framework]
	if frameworkPatterns == nil {
		frameworkPatterns = patterns["express"] // Default
	}

	for _, pattern := range frameworkPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				fieldName := match[1]
				if len(match) > 2 {
					fieldName = match[2] // Use the second capture group if available
				}

				hint := SchemaHint{
					FieldName:   fieldName,
					FieldType:   a.inferJavaScriptType(fieldName, content),
					IsRequired:  a.isFieldRequired(fieldName, content),
					Description: a.generateFieldDescription(fieldName),
					Example:     a.generateFieldExample(fieldName),
				}
				hints[fieldName] = hint
			}
		}
	}

	// Look for validation schemas (Joi, Yup, etc.)
	hints = a.mergeValidationSchemaHints(content, hints, "javascript")

	return hints
}

// extractGoSchemaHints extracts schema hints from Go code
func (a *AgentAnalyzer) extractGoSchemaHints(content, framework string) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	// Look for struct definitions
	structPattern := `type\s+(\w+)\s+struct\s*\{([^}]+)\}`
	re := regexp.MustCompile(structPattern)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			_ = match[1] // structName - not used in current implementation
			structBody := match[2]

			// Parse struct fields
			fieldPattern := `(\w+)\s+(\w+)\s+` + "`" + `json:"([^"]*)"` + "`"
			fieldRe := regexp.MustCompile(fieldPattern)
			fieldMatches := fieldRe.FindAllStringSubmatch(structBody, -1)

			for _, fieldMatch := range fieldMatches {
				if len(fieldMatch) > 3 {
					fieldName := fieldMatch[1]
					fieldType := fieldMatch[2]
					jsonTag := fieldMatch[3]

					// Parse JSON tag for additional info
					jsonName, isOmitted := a.parseGoJSONTag(jsonTag)
					if isOmitted {
						continue
					}
					if jsonName == "" {
						jsonName = fieldName
					}

					hint := SchemaHint{
						FieldName:   jsonName,
						FieldType:   a.mapGoTypeToJSONType(fieldType),
						IsRequired:  !strings.Contains(jsonTag, "omitempty"),
						Description: a.generateFieldDescription(jsonName),
						Example:     a.generateFieldExample(jsonName),
					}
					hints[jsonName] = hint
				}
			}
		}
	}

	return hints
}

// extractPythonSchemaHints extracts schema hints from Python code
func (a *AgentAnalyzer) extractPythonSchemaHints(content, framework string) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	// Look for Pydantic models
	pydanticPattern := `class\s+(\w+)\(BaseModel\):\s*([^}]+)`
	re := regexp.MustCompile(pydanticPattern)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			modelBody := match[2]

			// Parse Pydantic fields
			fieldPattern := `(\w+):\s*(\w+)(?:\s*=\s*Field\([^)]*\))?`
			fieldRe := regexp.MustCompile(fieldPattern)
			fieldMatches := fieldRe.FindAllStringSubmatch(modelBody, -1)

			for _, fieldMatch := range fieldMatches {
				if len(fieldMatch) > 2 {
					fieldName := fieldMatch[1]
					fieldType := fieldMatch[2]

					hint := SchemaHint{
						FieldName:   fieldName,
						FieldType:   a.mapPythonTypeToJSONType(fieldType),
						IsRequired:  !strings.Contains(modelBody, fieldName+"=Field(default="),
						Description: a.generateFieldDescription(fieldName),
						Example:     a.generateFieldExample(fieldName),
					}
					hints[fieldName] = hint
				}
			}
		}
	}

	// Look for FastAPI request models
	fastapiPattern := `@app\.(post|get|put|delete|patch)\([^)]*\)\s*\ndef\s+\w+\(([^)]+)\):`
	fastapiRe := regexp.MustCompile(fastapiPattern)
	fastapiMatches := fastapiRe.FindAllStringSubmatch(content, -1)

	for _, match := range fastapiMatches {
		if len(match) > 1 {
			params := match[1]
			// Parse function parameters
			paramPattern := `(\w+):\s*(\w+)`
			paramRe := regexp.MustCompile(paramPattern)
			paramMatches := paramRe.FindAllStringSubmatch(params, -1)

			for _, paramMatch := range paramMatches {
				if len(paramMatch) > 2 {
					paramName := paramMatch[1]
					paramType := paramMatch[2]

					hint := SchemaHint{
						FieldName:   paramName,
						FieldType:   a.mapPythonTypeToJSONType(paramType),
						IsRequired:  true,
						Description: a.generateFieldDescription(paramName),
						Example:     a.generateFieldExample(paramName),
					}
					hints[paramName] = hint
				}
			}
		}
	}

	return hints
}
