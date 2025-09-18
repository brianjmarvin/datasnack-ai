package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [workflow-file]",
	Short: "Convert n8n workflow to include webhook for evaluation",
	Long: `Convert an n8n workflow JSON file to include a webhook node that allows
the workflow to be executed programmatically and return results.

The converted workflow will be saved with "_eval" appended to the filename.

Example:
  ai-evaluator convert n8n/gmail-ai.json
  ai-evaluator convert /path/to/workflow.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowFile := args[0]

		// Check if file exists
		if _, err := os.Stat(workflowFile); os.IsNotExist(err) {
			log.Fatalf("Workflow file does not exist: %s", workflowFile)
		}

		// Read the workflow file
		log.Printf("Reading workflow from: %s", workflowFile)
		workflowData, err := os.ReadFile(workflowFile)
		if err != nil {
			log.Fatalf("Failed to read workflow file: %v", err)
		}

		// Parse the workflow JSON
		var workflow map[string]interface{}
		if err := json.Unmarshal(workflowData, &workflow); err != nil {
			log.Fatalf("Failed to parse workflow JSON: %v", err)
		}

		// Convert the workflow to include webhook
		convertedWorkflow, err := convertWorkflowToWebhook(workflow)
		if err != nil {
			log.Fatalf("Failed to convert workflow: %v", err)
		}

		// Generate output filename
		dir := filepath.Dir(workflowFile)
		filename := filepath.Base(workflowFile)
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		outputFile := filepath.Join(dir, name+"_eval"+ext)

		// Write the converted workflow
		convertedData, err := json.MarshalIndent(convertedWorkflow, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal converted workflow: %v", err)
		}

		if err := os.WriteFile(outputFile, convertedData, 0644); err != nil {
			log.Fatalf("Failed to write converted workflow: %v", err)
		}

		log.Printf("Converted workflow saved to: %s", outputFile)
		log.Println("The workflow now includes a webhook node for programmatic execution.")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}

// convertWorkflowToWebhook converts an n8n workflow to include a webhook node using AI
func convertWorkflowToWebhook(workflow map[string]interface{}) (map[string]interface{}, error) {
	// Initialize AI client for intelligent webhook conversion
	ai, err := initializeAIClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI client: %w", err)
	}

	// Convert workflow to JSON string for AI analysis
	workflowJSON, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow for AI analysis: %w", err)
	}

	// Use AI to analyze the workflow and generate webhook integration
	aiPrompt := fmt.Sprintf(`You are an expert n8n workflow engineer. Analyze this n8n workflow and add webhook nodes for programmatic execution.

Workflow JSON:
%s

CRITICAL REQUIREMENTS based on previous learnings:

1. WEBHOOK TRIGGER NODE:
   - Type: n8n-nodes-base.webhook
   - Method: POST
   - Path: "evaluate"
   - responseMode: "responseNode"
   - options.rawBody: true

2. WEBHOOK RESPONSE NODE:
   - Type: n8n-nodes-base.respondToWebhook
   - respondWith: "json"
   - responseBody: "={{ $json }}" (CRITICAL: Use actual workflow data, NOT hardcoded values)
   - responseHeaders: Content-Type: application/json

3. CONNECTION ANALYSIS:
   - Identify the ACTUAL final node in the workflow execution chain
   - Connect the final node to Webhook Response (NOT Webhook Response to itself)
   - Ensure Webhook Response has NO outgoing connections (empty main array)
   - Replace or connect to the original trigger node with Webhook Trigger

4. COMMON ISSUES TO AVOID:
   - Never hardcode responseBody like "myField": "value"
   - Never create circular connections (Webhook Response → Webhook Response)
   - Always use "={{ $json }}" for responseBody to return actual workflow output
   - Ensure the final workflow node connects to Webhook Response

5. WORKFLOW FLOW:
   Webhook Trigger → [existing workflow nodes] → Final Node → Webhook Response

Return ONLY the complete modified workflow JSON with proper webhook integration.`, string(workflowJSON))

	// Get AI-generated webhook integration
	aiResponse, err := ai.GenerateAI(aiPrompt, "", []map[string]string{})
	if err != nil {
		log.Printf("AI webhook integration failed, falling back to manual method: %v", err)
		return convertWorkflowToWebhookManual(workflow)
	}

	// Parse AI response to extract JSON
	convertedWorkflowJSON, err := extractJSONFromAIResponse(aiResponse)
	if err != nil {
		log.Printf("Failed to extract JSON from AI response, falling back to manual method: %v", err)
		return convertWorkflowToWebhookManual(workflow)
	}

	// Parse the AI-generated workflow
	var convertedWorkflow map[string]interface{}
	if err := json.Unmarshal([]byte(convertedWorkflowJSON), &convertedWorkflow); err != nil {
		log.Printf("Failed to parse AI-generated workflow JSON, falling back to manual method: %v", err)
		return convertWorkflowToWebhookManual(workflow)
	}

	// Validate that the converted workflow has the required webhook nodes
	if err := validateWebhookIntegration(convertedWorkflow); err != nil {
		log.Printf("AI-generated webhook integration validation failed, falling back to manual method: %v", err)
		return convertWorkflowToWebhookManual(workflow)
	}

	log.Println("Successfully used AI to convert workflow with webhook integration")
	return convertedWorkflow, nil
}

// convertWorkflowToWebhookManual provides a fallback manual webhook conversion method
func convertWorkflowToWebhookManual(workflow map[string]interface{}) (map[string]interface{}, error) {
	// Create a deep copy of the workflow
	workflowBytes, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow: %w", err)
	}

	var convertedWorkflow map[string]interface{}
	if err := json.Unmarshal(workflowBytes, &convertedWorkflow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}

	// Get nodes array
	nodes, ok := convertedWorkflow["nodes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("workflow does not contain nodes array")
	}

	// Find the first trigger node (usually manual trigger)
	var firstTriggerNode map[string]interface{}

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		nodeType, ok := nodeMap["type"].(string)
		if !ok {
			continue
		}

		// Look for manual trigger or other trigger nodes
		if strings.Contains(nodeType, "manualTrigger") ||
			strings.Contains(nodeType, "trigger") ||
			strings.Contains(nodeType, "webhook") {
			firstTriggerNode = nodeMap
			break
		}
	}

	if firstTriggerNode == nil {
		return nil, fmt.Errorf("no trigger node found in workflow")
	}

	// Create webhook node
	webhookNode := map[string]interface{}{
		"id":       "webhook-eval-trigger",
		"name":     "Webhook Trigger",
		"type":     "n8n-nodes-base.webhook",
		"position": []interface{}{100, 300},
		"parameters": map[string]interface{}{
			"httpMethod":   "POST",
			"path":         "evaluate",
			"responseMode": "responseNode",
			"options": map[string]interface{}{
				"rawBody": true,
			},
		},
		"typeVersion": 1,
	}

	// Create response node with proper configuration based on learnings
	responseNode := map[string]interface{}{
		"id":       "webhook-eval-response",
		"name":     "Webhook Response",
		"type":     "n8n-nodes-base.respondToWebhook",
		"position": []interface{}{800, 300},
		"parameters": map[string]interface{}{
			"respondWith":  "json",
			"responseBody": "={{ $json }}", // CRITICAL: Use actual workflow data, not hardcoded values
			"options": map[string]interface{}{
				"responseHeaders": map[string]interface{}{
					"entries": []interface{}{
						map[string]interface{}{
							"name":  "Content-Type",
							"value": "application/json",
						},
					},
				},
			},
		},
		"typeVersion": 1,
	}

	// Add webhook and response nodes to the workflow
	nodes = append(nodes, webhookNode, responseNode)
	convertedWorkflow["nodes"] = nodes

	// Update connections to connect webhook to the original trigger's connections
	connections, ok := convertedWorkflow["connections"].(map[string]interface{})
	if !ok {
		connections = make(map[string]interface{})
		convertedWorkflow["connections"] = connections
	}

	// Get the original trigger node's connections
	originalTriggerName, ok := firstTriggerNode["name"].(string)
	if !ok {
		originalTriggerName = "Execute workflow"
	}

	// Create webhook connection to the original trigger's first connection
	if originalConnections, exists := connections[originalTriggerName]; exists {
		connections["Webhook Trigger"] = originalConnections
	}

	// Find the last node in the workflow to connect to response
	lastNodeName := findLastNode(nodes, connections)
	if lastNodeName != "" {
		// Connect last node to webhook response (avoiding circular connections)
		responseConnections := map[string]interface{}{
			"main": []interface{}{
				[]interface{}{
					map[string]interface{}{
						"node":  "Webhook Response",
						"type":  "main",
						"index": 0,
					},
				},
			},
		}
		connections[lastNodeName] = responseConnections
		log.Printf("Connected final node '%s' to Webhook Response", lastNodeName)
	} else {
		log.Printf("Warning: Could not identify final node to connect to Webhook Response")
	}

	// Ensure Webhook Response has no outgoing connections (prevents circular references)
	connections["Webhook Response"] = map[string]interface{}{
		"main": []interface{}{[]interface{}{}},
	}

	convertedWorkflow["connections"] = connections

	return convertedWorkflow, nil
}

// findLastNode finds the last node in the workflow execution chain
func findLastNode(nodes []interface{}, connections map[string]interface{}) string {
	// Simple heuristic: find a node that has no outgoing connections
	// or find the node that appears to be the final output node

	nodeNames := make(map[string]bool)
	connectedNodes := make(map[string]bool)

	// Collect all node names
	for _, node := range nodes {
		if nodeMap, ok := node.(map[string]interface{}); ok {
			if name, ok := nodeMap["name"].(string); ok {
				nodeNames[name] = true
			}
		}
	}

	// Find nodes that are connected to by others
	for _, connectionData := range connections {
		if connectionMap, ok := connectionData.(map[string]interface{}); ok {
			if mainConnections, ok := connectionMap["main"].([]interface{}); ok {
				for _, mainConn := range mainConnections {
					if mainConnArray, ok := mainConn.([]interface{}); ok {
						for _, conn := range mainConnArray {
							if connMap, ok := conn.(map[string]interface{}); ok {
								if nodeName, ok := connMap["node"].(string); ok {
									connectedNodes[nodeName] = true
								}
							}
						}
					}
				}
			}
		}
	}

	// Find nodes that are not connected to by others (potential end nodes)
	for nodeName := range nodeNames {
		if !connectedNodes[nodeName] &&
			nodeName != "Webhook Trigger" &&
			nodeName != "Webhook Response" {
			// Look for nodes that seem to be output/final nodes
			if strings.Contains(strings.ToLower(nodeName), "response") ||
				strings.Contains(strings.ToLower(nodeName), "output") ||
				strings.Contains(strings.ToLower(nodeName), "result") ||
				strings.Contains(strings.ToLower(nodeName), "insert") ||
				strings.Contains(strings.ToLower(nodeName), "save") ||
				strings.Contains(strings.ToLower(nodeName), "merge") ||
				strings.Contains(strings.ToLower(nodeName), "aggregate") {
				return nodeName
			}
		}
	}

	// If no obvious end node found, return the last node that's not a trigger or webhook
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodeMap, ok := nodes[i].(map[string]interface{}); ok {
			if name, ok := nodeMap["name"].(string); ok {
				lowerName := strings.ToLower(name)
				if !strings.Contains(lowerName, "trigger") &&
					!strings.Contains(lowerName, "execute") &&
					!strings.Contains(lowerName, "webhook") {
					return name
				}
			}
		}
	}

	return ""
}

// extractJSONFromAIResponse extracts JSON from AI response, handling cases where AI includes extra text
func extractJSONFromAIResponse(aiResponse string) (string, error) {
	// Clean the response first
	cleaned := strings.TrimSpace(aiResponse)

	// Try to find JSON boundaries
	jsonStart := strings.Index(cleaned, "{")
	jsonEnd := strings.LastIndex(cleaned, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return "", fmt.Errorf("no valid JSON found in AI response: %s", cleaned[:min(len(cleaned), 200)])
	}

	jsonStr := cleaned[jsonStart : jsonEnd+1]

	// Validate that it's valid JSON
	var testJSON map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &testJSON); err != nil {
		return "", fmt.Errorf("AI response contains invalid JSON: %w, response: %s", err, jsonStr[:min(len(jsonStr), 200)])
	}

	return jsonStr, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateWebhookIntegration validates that the converted workflow has proper webhook integration
func validateWebhookIntegration(workflow map[string]interface{}) error {
	nodes, ok := workflow["nodes"].([]interface{})
	if !ok {
		return fmt.Errorf("workflow does not contain nodes array")
	}

	// Check for webhook trigger node
	hasWebhookTrigger := false
	hasWebhookResponse := false

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		nodeType, ok := nodeMap["type"].(string)
		if !ok {
			continue
		}

		if nodeType == "n8n-nodes-base.webhook" {
			hasWebhookTrigger = true
		}

		if nodeType == "n8n-nodes-base.respondToWebhook" {
			hasWebhookResponse = true
		}
	}

	if !hasWebhookTrigger {
		return fmt.Errorf("workflow does not contain webhook trigger node")
	}

	if !hasWebhookResponse {
		return fmt.Errorf("workflow does not contain webhook response node")
	}

	return nil
}
