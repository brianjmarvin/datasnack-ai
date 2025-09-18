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

	// Use AI to analyze the workflow and generate comprehensive webhook integration
	aiPrompt := fmt.Sprintf(`You are an expert n8n workflow engineer specializing in CLI evaluation instrumentation. Analyze this n8n workflow and add comprehensive webhook-based evaluation capabilities.

Workflow JSON:
%s

## Core Requirements for CLI Evaluation Instrumentation

### 1. Endpoint Design Principles
- **Minimal Response Payload**: Return only essential metrics and response data
- **Simplified Integration**: Easy for CLI tools to consume
- **Comprehensive Error Handling**: Graceful failure modes
- **Provider Agnostic**: Support multiple AI models and services
- **No Evaluation Logic**: Pure instrumentation - evaluation handled by calling CLI

### 2. Essential Response Payload Structure
The webhook response must follow this standardized format:
`+"```"+`json
{
  "success": boolean,
  "query": string,
  "response": string,
  "metrics": {
    "response_time": float,
    "total_time": float,
    "response_length": int,
    "word_count": int,
    "character_count": int,
    "has_content": boolean,
    "timestamp": string
  },
  "provider_info": {
    "provider": string,
    "model": string,
    "temperature": string,
    "reasoning_effort": string
  },
  "timing": {
    "response_time": float,
    "total_time": float
  },
  "error": string | null,
  "workflow_metrics": {
    "workflow_name": string,
    "nodes_executed": int,
    "custom_metrics": object
  }
}
`+"```"+`

### 3. Required Node Structure

#### A. Webhook Trigger Node:
`+"```"+`json
{
  "id": "webhook-trigger-cli",
  "name": "Webhook Trigger (CLI Evaluation)",
  "type": "n8n-nodes-base.webhook",
  "parameters": {
    "httpMethod": "POST",
    "path": "evaluate",
    "responseMode": "responseNode",
    "options": {}
  }
}
`+"```"+`

#### B. Request Preparation Node (Code):
`+"```"+`javascript
// Prepare evaluation request data from webhook
const webhookData = $input.first().json;
const startTime = Date.now();

// Extract request parameters
const requestData = {
  query: webhookData.query || webhookData.body?.query || 'Test workflow execution',
  provider: webhookData.provider || 'openai',
  model: webhookData.model || 'gpt-3.5-turbo',
  temperature: webhookData.temperature || 0.0,
  workflow_type: webhookData.workflow_type || 'general',
  custom_params: webhookData.custom_params || {},
  start_time: startTime,
  request_id: webhookData.request_id || `+"`"+`eval_${startTime}`+"`"+`
};

// Create test data for your specific workflow
const testData = {
  input_text: requestData.query,
  input_data: requestData.custom_params,
  evaluation_request: requestData
};

return { json: testData };
`+"```"+`

#### C. Metrics Calculation Node (Code):
`+"```"+`javascript
// Calculate evaluation metrics for CLI response
const inputData = $input.first().json;
const endTime = Date.now();

// Get original request data
const originalData = $('Prepare Evaluation Request').item.json;
const requestData = originalData.evaluation_request;
const startTime = requestData.start_time;

// Calculate timing metrics
const totalTime = (endTime - startTime) / 1000;
const responseTime = totalTime;

// Extract response content - customize based on workflow output
let responseContent = '';
if (inputData.choices && inputData.choices[0] && inputData.choices[0].message) {
  responseContent = inputData.choices[0].message.content || '';
} else if (inputData.body) {
  responseContent = inputData.body || '';
} else if (inputData.result) {
  responseContent = inputData.result || '';
} else if (inputData.output) {
  responseContent = inputData.output || '';
} else if (typeof inputData === 'string') {
  responseContent = inputData;
} else {
  responseContent = JSON.stringify(inputData);
}

// Calculate content metrics
const responseLength = responseContent.length;
const wordCount = responseContent.split(/\s+/).filter(word => word.length > 0).length;
const characterCount = responseContent.length;
const hasContent = responseLength > 0;

// Count nodes executed
const nodesExecuted = Object.keys($).length;

// Create standardized evaluation response
const evaluationResponse = {
  success: true,
  query: requestData.query,
  response: responseContent,
  metrics: {
    response_time: responseTime,
    total_time: totalTime,
    response_length: responseLength,
    word_count: wordCount,
    character_count: characterCount,
    has_content: hasContent,
    timestamp: new Date().toISOString()
  },
  provider_info: {
    provider: requestData.provider,
    model: requestData.model,
    temperature: requestData.temperature.toString(),
    reasoning_effort: 'medium'
  },
  timing: {
    response_time: responseTime,
    total_time: totalTime
  },
  error: null,
  workflow_metrics: {
    workflow_name: 'Your Workflow Name',
    nodes_executed: nodesExecuted,
    custom_metrics: {}
  }
};

return { json: evaluationResponse };
`+"```"+`

#### D. Webhook Response Node:
`+"```"+`json
{
  "id": "webhook-response-cli",
  "name": "Webhook Response (CLI)",
  "type": "n8n-nodes-base.respondToWebhook",
  "parameters": {
    "respondWith": "json",
    "responseBody": "={{ $json }}",
    "options": {
      "responseHeaders": {
        "entries": [
          {
            "name": "Content-Type",
            "value": "application/json"
          },
          {
            "name": "Access-Control-Allow-Origin",
            "value": "*"
          }
        ]
      }
    }
  }
}
`+"```"+`

### 4. Connection Requirements
- Webhook Trigger → Request Preparation → [existing workflow nodes] → Metrics Calculation → Webhook Response
- Ensure Webhook Response has NO outgoing connections
- Connect the final workflow output to Metrics Calculation node
- Handle errors gracefully with proper error responses

### 5. Workflow-Specific Customizations
Based on the workflow type, customize the request preparation and metrics calculation:
- **AI/LLM Workflows**: Handle model parameters, system prompts, max_tokens
- **Data Processing**: Handle input_data, processing_type, output_format
- **Email/Communication**: Handle email_type, recipient, subject

### 6. Error Handling
Include error handling that returns structured error responses following the same format but with success: false and appropriate error messages.

### 7. Success Criteria
The instrumentation is successful when:
- Simple queries return real responses (not mock responses)
- Complex queries return real workflow outputs
- All metrics are collected accurately
- Error handling works gracefully
- CLI integration is straightforward

## Implementation Instructions

1. **Analyze the existing workflow** to understand its structure and purpose
2. **Add the required nodes** in the correct order
3. **Create proper connections** between all nodes
4. **Customize the Code nodes** based on the workflow's specific functionality
5. **Ensure the response format** matches the standardized structure exactly
6. **Test error scenarios** to ensure graceful failure handling

Return ONLY the complete modified workflow JSON with proper webhook integration that follows the CLI evaluation instrumentation standards.`, string(workflowJSON))

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
