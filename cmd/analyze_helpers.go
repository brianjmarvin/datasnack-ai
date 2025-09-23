package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Helper methods for type inference and mapping
func (a *AgentAnalyzer) inferJavaScriptType(fieldName, content string) string {
	// Look for type hints or validation
	if strings.Contains(content, fieldName+": string") {
		return "string"
	}
	if strings.Contains(content, fieldName+": number") {
		return "number"
	}
	if strings.Contains(content, fieldName+": boolean") {
		return "boolean"
	}
	if strings.Contains(content, fieldName+": array") || strings.Contains(content, fieldName+"[]") {
		return "array"
	}

	// Infer from field name
	fieldNameLower := strings.ToLower(fieldName)
	if strings.Contains(fieldNameLower, "id") || strings.Contains(fieldNameLower, "count") || strings.Contains(fieldNameLower, "number") {
		return "number"
	}
	if strings.Contains(fieldNameLower, "is") || strings.Contains(fieldNameLower, "has") || strings.Contains(fieldNameLower, "can") {
		return "boolean"
	}
	if strings.Contains(fieldNameLower, "list") || strings.Contains(fieldNameLower, "items") {
		return "array"
	}

	return "string" // Default
}

func (a *AgentAnalyzer) mapGoTypeToJSONType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "[]string", "[]int", "[]float64", "[]interface{}":
		return "array"
	default:
		return "string"
	}
}

func (a *AgentAnalyzer) mapPythonTypeToJSONType(pythonType string) string {
	switch pythonType {
	case "str", "string":
		return "string"
	case "int", "float", "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "list", "List", "array":
		return "array"
	default:
		return "string"
	}
}

func (a *AgentAnalyzer) isFieldRequired(fieldName, content string) bool {
	// Look for validation patterns that indicate required fields
	requiredPatterns := []string{
		fieldName + ".*required.*true",
		fieldName + ".*required",
		"required.*" + fieldName,
		fieldName + ".*not.*null",
		fieldName + ".*not.*empty",
	}

	for _, pattern := range requiredPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if re.MatchString(content) {
			return true
		}
	}

	// Default to required for common AI agent fields
	requiredFields := []string{"query", "prompt", "input", "message", "text", "question"}
	for _, reqField := range requiredFields {
		if strings.EqualFold(fieldName, reqField) {
			return true
		}
	}

	return false
}

func (a *AgentAnalyzer) generateFieldDescription(fieldName string) string {
	descriptions := map[string]string{
		"query":     "The input query or prompt for the AI agent",
		"prompt":    "The prompt or instruction for the AI agent",
		"input":     "The input data for processing",
		"message":   "The message content to be processed",
		"text":      "The text content to be analyzed",
		"question":  "The question to be answered",
		"context":   "Additional context for the request",
		"options":   "Configuration options for the request",
		"settings":  "Settings and parameters for the request",
		"metadata":  "Metadata associated with the request",
		"userId":    "Unique identifier for the user",
		"sessionId": "Unique identifier for the session",
		"timestamp": "Timestamp of the request",
	}

	if desc, exists := descriptions[strings.ToLower(fieldName)]; exists {
		return desc
	}

	return fmt.Sprintf("The %s field", fieldName)
}

func (a *AgentAnalyzer) generateFieldExample(fieldName string) string {
	examples := map[string]string{
		"query":     "What is the weather like today?",
		"prompt":    "Analyze the following data and provide insights",
		"input":     "Sample input data",
		"message":   "Hello, how can I help you?",
		"text":      "This is sample text to be processed",
		"question":  "What are the latest developments in AI?",
		"context":   "Additional context information",
		"userId":    "user123",
		"sessionId": "session456",
		"timestamp": "2024-01-01T00:00:00Z",
	}

	if example, exists := examples[strings.ToLower(fieldName)]; exists {
		return example
	}

	return fmt.Sprintf("example_%s", fieldName)
}

func (a *AgentAnalyzer) parseGoJSONTag(jsonTag string) (string, bool) {
	// Parse JSON tag like "name,omitempty" or "name"
	parts := strings.Split(jsonTag, ",")
	if len(parts) == 0 {
		return "", true
	}

	name := parts[0]
	isOmitted := false
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) == "omitempty" {
			isOmitted = true
			break
		}
	}

	return name, isOmitted
}

func (a *AgentAnalyzer) findSchemaFiles(language string) []string {
	var schemaFiles []string

	switch language {
	case "javascript":
		// Look for schema definition files
		patterns := []string{"schema.js", "schemas.js", "validation.js", "models.js", "types.js", "*.schema.js"}
		for _, pattern := range patterns {
			cmd := exec.Command("find", a.agentRootFolder, "-name", pattern, "-type", "f")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, line := range lines {
					if line != "" {
						relPath, _ := filepath.Rel(a.agentRootFolder, line)
						schemaFiles = append(schemaFiles, relPath)
					}
				}
			}
		}
	case "go":
		// Look for model files
		patterns := []string{"*model*.go", "*schema*.go", "*types*.go"}
		for _, pattern := range patterns {
			cmd := exec.Command("find", a.agentRootFolder, "-name", pattern, "-type", "f")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, line := range lines {
					if line != "" {
						relPath, _ := filepath.Rel(a.agentRootFolder, line)
						schemaFiles = append(schemaFiles, relPath)
					}
				}
			}
		}
	case "python":
		// Look for model files
		patterns := []string{"*model*.py", "*schema*.py", "*types*.py", "*pydantic*.py"}
		for _, pattern := range patterns {
			cmd := exec.Command("find", a.agentRootFolder, "-name", pattern, "-type", "f")
			output, err := cmd.Output()
			if err == nil {
				lines := strings.Split(strings.TrimSpace(string(output)), "\n")
				for _, line := range lines {
					if line != "" {
						relPath, _ := filepath.Rel(a.agentRootFolder, line)
						schemaFiles = append(schemaFiles, relPath)
					}
				}
			}
		}
	}

	return schemaFiles
}

func (a *AgentAnalyzer) analyzeSchemaFile(filePath, language string) map[string]SchemaHint {
	hints := make(map[string]SchemaHint)

	_, err := os.ReadFile(filePath)
	if err != nil {
		return hints
	}

	// This would contain language-specific schema file analysis
	// For now, delegate to the main file analysis
	return a.analyzeFileForSchemaHints(filePath, language, "unknown")
}

func (a *AgentAnalyzer) mergeValidationSchemaHints(content string, hints map[string]SchemaHint, language string) map[string]SchemaHint {
	// Look for validation library patterns (Joi, Yup, etc.)
	validationPatterns := map[string][]string{
		"joi": {
			`(\w+):\s*Joi\.(\w+)\(\)`,
			`(\w+):\s*Joi\.string\(\)`,
			`(\w+):\s*Joi\.number\(\)`,
			`(\w+):\s*Joi\.boolean\(\)`,
		},
		"yup": {
			`(\w+):\s*yup\.(\w+)\(\)`,
			`(\w+):\s*yup\.string\(\)`,
			`(\w+):\s*yup\.number\(\)`,
			`(\w+):\s*yup\.boolean\(\)`,
		},
	}

	for _, patterns := range validationPatterns {
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 2 {
					fieldName := match[1]
					fieldType := match[2]

					// Update existing hint or create new one
					if hint, exists := hints[fieldName]; exists {
						hint.FieldType = a.mapValidationTypeToJSONType(fieldType)
						hints[fieldName] = hint
					} else {
						hints[fieldName] = SchemaHint{
							FieldName:   fieldName,
							FieldType:   a.mapValidationTypeToJSONType(fieldType),
							IsRequired:  true,
							Description: a.generateFieldDescription(fieldName),
							Example:     a.generateFieldExample(fieldName),
						}
					}
				}
			}
		}
	}

	return hints
}

func (a *AgentAnalyzer) mapValidationTypeToJSONType(validationType string) string {
	switch strings.ToLower(validationType) {
	case "string", "str":
		return "string"
	case "number", "num", "int", "integer", "float":
		return "number"
	case "boolean", "bool":
		return "boolean"
	case "array", "arr":
		return "array"
	default:
		return "string"
	}
}

// buildComprehensiveCodeContext builds a comprehensive context for AI analysis
func (a *AgentAnalyzer) buildComprehensiveCodeContext(analysis *AgentAnalysis, staticHints map[string]SchemaHint) string {
	var context strings.Builder

	// Basic information
	context.WriteString(fmt.Sprintf("Agent Purpose: %s\n", a.agentPurpose))
	context.WriteString(fmt.Sprintf("Language: %s, Framework: %s\n", analysis.Language, analysis.Framework))
	context.WriteString(fmt.Sprintf("Base URL: %s\n", a.baseURL))
	context.WriteString(fmt.Sprintf("Main Endpoint: %s\n", a.endpoint))

	// Static analysis results
	context.WriteString("\n=== STATIC ANALYSIS RESULTS ===\n")
	context.WriteString(fmt.Sprintf("Discovered %d potential fields:\n", len(staticHints)))
	for fieldName, hint := range staticHints {
		context.WriteString(fmt.Sprintf("- %s (%s, required: %v): %s\n",
			fieldName, hint.FieldType, hint.IsRequired, hint.Description))
	}

	// Code context from main files
	context.WriteString("\n=== CODE CONTEXT ===\n")
	for _, mainFile := range analysis.MainFiles {
		filePath := filepath.Join(a.agentRootFolder, mainFile)
		if data, err := os.ReadFile(filePath); err == nil {
			context.WriteString(fmt.Sprintf("\n--- %s ---\n", mainFile))
			content := string(data)
			// Limit content but preserve important parts
			if len(content) > 3000 {
				// Try to keep the beginning and end
				context.WriteString(content[:1500])
				context.WriteString("\n... [truncated] ...\n")
				context.WriteString(content[len(content)-1500:])
			} else {
				context.WriteString(content)
			}
		}
	}

	// Package information
	if len(analysis.PackageInfo) > 0 {
		context.WriteString("\n=== PACKAGE INFORMATION ===\n")
		for k, v := range analysis.PackageInfo {
			context.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}

	return context.String()
}

// generateRequestSchemaWithAI generates request schema using AI with enhanced context
func (a *AgentAnalyzer) generateRequestSchemaWithAI(codeContext string, staticHints map[string]SchemaHint) (map[string]interface{}, error) {
	// Convert static hints to JSON for AI context
	hintsJSON, _ := json.Marshal(staticHints)

	prompt := fmt.Sprintf(`Analyze this AI agent code and generate a comprehensive JSON schema for the request payload of the main endpoint.

%s

Static Analysis Hints (JSON):
%s

Based on the comprehensive analysis above, generate a detailed JSON schema that includes:
1. All discovered fields from static analysis
2. Additional fields that might be expected based on the agent's purpose
3. Proper field types, descriptions, and realistic examples
4. Required vs optional field classification
5. Field constraints, enums, and validation rules
6. Nested objects and arrays where appropriate
7. Common AI agent patterns (context, options, metadata, etc.)

The schema should be production-ready and comprehensive. Return only a valid JSON schema object.`, codeContext, string(hintsJSON))

	systemPrompt := `You are an expert API schema analyst specializing in AI agents. You excel at:
- Analyzing code to understand API structure
- Generating comprehensive JSON schemas
- Identifying common AI agent patterns
- Creating realistic examples and descriptions
- Ensuring schemas are production-ready

Generate a detailed, accurate JSON schema based on the provided analysis.`

	response, err := a.ai.GenerateAI(prompt, systemPrompt, nil)
	if err != nil {
		return nil, err
	}

	// Parse and validate the response
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(response), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}

	// Validate that it's a proper JSON schema
	if schema["type"] != "object" {
		return nil, fmt.Errorf("AI response is not a valid object schema")
	}

	return schema, nil
}

// generateResponseSchemaWithAI generates response schema using AI with enhanced context
func (a *AgentAnalyzer) generateResponseSchemaWithAI(codeContext string, staticHints map[string]SchemaHint) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this AI agent code and generate a comprehensive JSON schema for the response payload of the main endpoint.

%s

Based on the analysis above, generate a detailed JSON schema for the response that includes:
1. The main response content (text, data, results)
2. Metadata fields (processing time, status, etc.)
3. Error handling structure
4. Additional response fields based on the agent's purpose
5. Proper field types, descriptions, and examples
6. Nested objects and arrays where appropriate
7. Common AI agent response patterns

The schema should cover both success and error responses. Return only a valid JSON schema object.`, codeContext)

	systemPrompt := `You are an expert API response schema analyst specializing in AI agents. You excel at:
- Understanding AI agent response patterns
- Generating comprehensive response schemas
- Including proper error handling structures
- Creating realistic examples and descriptions
- Ensuring schemas cover all response scenarios

Generate a detailed, accurate JSON schema for the response payload.`

	response, err := a.ai.GenerateAI(prompt, systemPrompt, nil)
	if err != nil {
		return nil, err
	}

	// Parse and validate the response
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(response), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
	}

	return schema, nil
}

// generateFallbackRequestSchema creates a fallback schema when AI analysis fails
func (a *AgentAnalyzer) generateFallbackRequestSchema(analysis *AgentAnalysis, staticHints map[string]SchemaHint) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	// Add fields from static analysis
	for fieldName, hint := range staticHints {
		property := map[string]interface{}{
			"type":        hint.FieldType,
			"description": hint.Description,
			"example":     hint.Example,
		}

		// Add constraints if available
		for constraint, value := range hint.Constraints {
			property[constraint] = value
		}

		properties[fieldName] = property

		if hint.IsRequired {
			required = append(required, fieldName)
		}
	}

	// Add common AI agent fields if not already present
	commonFields := map[string]map[string]interface{}{
		"query": {
			"type":        "string",
			"description": "The input query or prompt for the AI agent",
			"example":     "What is the weather like today?",
		},
		"context": {
			"type":        "object",
			"description": "Additional context for the request",
			"example":     map[string]interface{}{"sessionId": "abc123"},
		},
		"options": {
			"type":        "object",
			"description": "Configuration options for the request",
			"example":     map[string]interface{}{"maxResults": 5},
		},
	}

	for fieldName, fieldDef := range commonFields {
		if _, exists := properties[fieldName]; !exists {
			properties[fieldName] = fieldDef
			if fieldName == "query" {
				required = append(required, fieldName)
			}
		}
	}

	return map[string]interface{}{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": properties,
		"required":   required,
		"title":      "AI Agent Request Schema",
	}
}

// generateFallbackResponseSchema creates a fallback response schema when AI analysis fails
func (a *AgentAnalyzer) generateFallbackResponseSchema(analysis *AgentAnalysis, staticHints map[string]SchemaHint) map[string]interface{} {
	return map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"response": map[string]interface{}{
				"type":        "string",
				"description": "The AI agent's response to the query",
				"example":     "The weather is sunny with a temperature of 75°F.",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Metadata about the response",
				"properties": map[string]interface{}{
					"processingTime": map[string]interface{}{
						"type":        "number",
						"description": "Time taken to process the query",
						"example":     2.5,
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"description": "Timestamp of the response",
						"example":     "2024-01-01T00:00:00Z",
					},
				},
			},
		},
		"required": []string{"response"},
		"title":    "AI Agent Response Schema",
	}
}

// performRuntimeSchemaDiscovery attempts to discover schema information at runtime
func (a *AgentAnalyzer) performRuntimeSchemaDiscovery() (map[string]interface{}, error) {
	// This would attempt to make test requests to the endpoint to discover schema
	// For now, return empty hints
	return make(map[string]interface{}), nil
}

// mergeRuntimeHints merges runtime discovery hints with existing analysis
func (a *AgentAnalyzer) mergeRuntimeHints(analysis *AgentAnalysis, runtimeHints map[string]interface{}) {
	// This would merge runtime discovery results with the existing schemas
	// For now, this is a placeholder
}
