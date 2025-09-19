/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"datasnack/cloneAttack"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// PromptConfig represents the structure of the prompt_config.yaml file
type PromptConfig struct {
	Version          string                 `yaml:"version"`
	LastUpdated      string                 `yaml:"last_updated"`
	Description      string                 `yaml:"description"`
	OriginalPrompts  map[string]PromptInfo  `yaml:"original_prompts"`
	ModifiedPrompts  map[string]interface{} `yaml:"modified_prompts"`
	UsageStats       map[string]interface{} `yaml:"usage_stats"`
	PromptCategories map[string]Category    `yaml:"prompt_categories"`
}

type PromptInfo struct {
	Prompt      string `yaml:"prompt"`
	Location    string `yaml:"location"`
	AgentType   string `yaml:"agent_type"`
	Description string `yaml:"description"`
}

type Category struct {
	Description string   `yaml:"description"`
	Prompts     []string `yaml:"prompts"`
}

// EvaluationResults represents the structure of evaluation results
type EvaluationResults struct {
	TotalCalls          int                    `json:"totalCalls"`
	SuccessfulCalls     int                    `json:"successfulCalls"`
	FailedCalls         int                    `json:"failedCalls"`
	AverageResponseTime float64                `json:"averageResponseTime"`
	Vulnerabilities     []Vulnerability        `json:"vulnerabilities"`
	PromptOptimizations []interface{}          `json:"promptOptimizations"`
	PerformanceMetrics  map[string]interface{} `json:"performanceMetrics"`
	Recommendations     []string               `json:"recommendations"`
	StartTime           string                 `json:"startTime"`
	EndTime             string                 `json:"endTime"`
}

type Vulnerability struct {
	Type        string `json:"Type"`
	Severity    string `json:"Severity"`
	Description string `json:"Description"`
	CallID      string `json:"CallID"`
	Score       int    `json:"Score"`
	Response    string `json:"Response"`
	Prompt      string `json:"Prompt"`
}

// PromptSuggestion represents a suggestion for improving a prompt
type PromptSuggestion struct {
	PromptName         string   `json:"prompt_name"`
	CurrentPrompt      string   `json:"current_prompt"`
	SuggestedPrompt    string   `json:"suggested_prompt"`
	Reasoning          string   `json:"reasoning"`
	VulnerabilityTypes []string `json:"vulnerability_types"`
	Severity           string   `json:"severity"`
	Confidence         float64  `json:"confidence"`
	Impact             string   `json:"impact"`
}

// SuggestionsReport represents the complete suggestions report
type SuggestionsReport struct {
	GeneratedAt            string             `json:"generated_at"`
	EvaluationFile         string             `json:"evaluation_file"`
	PromptConfigFile       string             `json:"prompt_config_file"`
	TotalVulnerabilities   int                `json:"total_vulnerabilities"`
	VulnerabilitySummary   map[string]int     `json:"vulnerability_summary"`
	Suggestions            []PromptSuggestion `json:"suggestions"`
	OverallRecommendations []string           `json:"overall_recommendations"`
	AnalysisSummary        string             `json:"analysis_summary"`
}

// go run . suggestions
var suggestionsCmd = &cobra.Command{
	Use:   "suggestions",
	Short: "Generate prompt improvement suggestions based on evaluation results",
	Long: `Analyzes the most recent evaluation results and generates suggestions for improving
the AI agent's prompts based on identified vulnerabilities and performance issues.

The command:
1. Finds the most recent evaluation_results_*.json file
2. Loads the prompt_config.yaml from the agent's root folder
3. Analyzes vulnerabilities and performance issues
4. Generates specific suggestions for prompt improvements
5. Saves suggestions to a prompt_suggestions_*.json file`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load agent configuration to get the root folder
		configPath := os.Getenv("AGENT_CONFIG")
		if configPath == "" {
			configPath = "config/agentConfig.json"
		}

		log.Println("Reading agent configuration from:", configPath)
		configData, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatalln("Failed to read agent config file:", err)
		}

		var agentConfig PythonAgentConfig
		if err := json.Unmarshal(configData, &agentConfig); err != nil {
			log.Fatalln("Failed to unmarshal agent config:", err)
		}

		// Find the most recent evaluation results file
		resultsDir := "results"
		evaluationFile, err := findMostRecentEvaluationFile(resultsDir)
		if err != nil {
			log.Fatalln("Failed to find evaluation results file:", err)
		}

		log.Printf("Using evaluation results from: %s", evaluationFile)

		// Load evaluation results
		evaluationResults, err := loadEvaluationResults(evaluationFile)
		if err != nil {
			log.Fatalln("Failed to load evaluation results:", err)
		}

		// Load prompt configuration
		promptConfigPath := filepath.Join(agentConfig.AgentRootFolder, "backend", "evaluation", "config", "prompt_config.yaml")
		log.Printf("Loading prompt config from: %s", promptConfigPath)

		promptConfig, err := loadPromptConfig(promptConfigPath)
		if err != nil {
			log.Fatalln("Failed to load prompt config:", err)
		}

		// Initialize AI client for generating suggestions
		ai, err := initializeAIClient()
		if err != nil {
			log.Fatalln("Failed to initialize AI client:", err)
		}

		// Generate suggestions
		log.Println("Analyzing evaluation results and generating suggestions...")
		suggestions, err := generatePromptSuggestions(evaluationResults, promptConfig, ai)
		if err != nil {
			log.Fatalln("Failed to generate suggestions:", err)
		}

		// Create suggestions report
		report := createSuggestionsReport(evaluationFile, promptConfigPath, evaluationResults, suggestions)

		// Save suggestions to file
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("results/prompt_suggestions_%s.json", timestamp)

		reportJSON, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatalln("Failed to marshal suggestions report:", err)
		}

		if err := os.WriteFile(filename, reportJSON, 0644); err != nil {
			log.Fatalln("Failed to write suggestions file:", err)
		}

		log.Printf("Suggestions saved to: %s", filename)
		log.Printf("Generated %d prompt suggestions based on %d vulnerabilities", len(suggestions), len(evaluationResults.Vulnerabilities))
	},
}

// findMostRecentEvaluationFile finds the most recent evaluation_results_*.json file
func findMostRecentEvaluationFile(resultsDir string) (string, error) {
	var files []string

	err := filepath.WalkDir(resultsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), "evaluation_results_") && strings.HasSuffix(d.Name(), ".json") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no evaluation results files found in %s", resultsDir)
	}

	// Sort by modification time (most recent first)
	sort.Slice(files, func(i, j int) bool {
		info1, _ := os.Stat(files[i])
		info2, _ := os.Stat(files[j])
		return info1.ModTime().After(info2.ModTime())
	})

	return files[0], nil
}

// loadEvaluationResults loads evaluation results from a JSON file
func loadEvaluationResults(filename string) (*EvaluationResults, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var results EvaluationResults
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// loadPromptConfig loads prompt configuration from a YAML file
func loadPromptConfig(filename string) (*PromptConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config PromptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// generatePromptSuggestions generates suggestions based on evaluation results and prompt config
func generatePromptSuggestions(results *EvaluationResults, config *PromptConfig, ai cloneAttack.AIClient) ([]PromptSuggestion, error) {
	var suggestions []PromptSuggestion

	// Analyze vulnerabilities by type
	vulnerabilityTypes := make(map[string]int)
	for _, vuln := range results.Vulnerabilities {
		vulnerabilityTypes[vuln.Type]++
	}

	// Generate suggestions for each prompt category
	for categoryName, category := range config.PromptCategories {
		log.Printf("Analyzing category: %s", categoryName)

		for _, promptName := range category.Prompts {
			if promptInfo, exists := config.OriginalPrompts[promptName]; exists {
				suggestion, err := generateSuggestionForPrompt(promptName, promptInfo, results, ai)
				if err != nil {
					log.Printf("Failed to generate suggestion for %s: %v", promptName, err)
					continue
				}

				if suggestion != nil {
					suggestions = append(suggestions, *suggestion)
				}
			}
		}
	}

	return suggestions, nil
}

// generateSuggestionForPrompt generates a suggestion for a specific prompt
func generateSuggestionForPrompt(promptName string, promptInfo PromptInfo, results *EvaluationResults, ai cloneAttack.AIClient) (*PromptSuggestion, error) {
	// Find relevant vulnerabilities for this prompt type
	relevantVulns := findRelevantVulnerabilities(promptName, promptInfo, results)

	if len(relevantVulns) == 0 {
		return nil, nil // No relevant vulnerabilities found
	}

	// Create analysis prompt for AI
	analysisPrompt := createAnalysisPrompt(promptName, promptInfo, relevantVulns)

	// Get AI suggestion
	response, err := ai.GenerateAI(analysisPrompt, "You are an expert in AI prompt engineering and security. Analyze the prompt and vulnerabilities to provide specific, actionable suggestions for improvement.", nil)
	if err != nil {
		return nil, err
	}

	// Parse AI response to extract suggestion
	suggestion := parseAISuggestion(promptName, promptInfo.Prompt, response, relevantVulns)

	return suggestion, nil
}

// findRelevantVulnerabilities finds vulnerabilities relevant to a specific prompt
func findRelevantVulnerabilities(promptName string, promptInfo PromptInfo, results *EvaluationResults) []Vulnerability {
	var relevant []Vulnerability

	// Simple heuristic: look for vulnerabilities that might be related to this prompt type
	for _, vuln := range results.Vulnerabilities {
		// Check if vulnerability type matches prompt agent type or category
		if strings.Contains(strings.ToLower(vuln.Type), strings.ToLower(promptInfo.AgentType)) ||
			strings.Contains(strings.ToLower(vuln.Description), strings.ToLower(promptInfo.AgentType)) {
			relevant = append(relevant, vuln)
		}
	}

	// If no specific matches, include high-severity vulnerabilities
	if len(relevant) == 0 {
		for _, vuln := range results.Vulnerabilities {
			if vuln.Severity == "high" || vuln.Severity == "medium" {
				relevant = append(relevant, vuln)
			}
		}
	}

	return relevant
}

// createAnalysisPrompt creates a prompt for AI analysis
func createAnalysisPrompt(promptName string, promptInfo PromptInfo, vulnerabilities []Vulnerability) string {
	vulnSummary := ""
	for i, vuln := range vulnerabilities {
		if i < 5 { // Limit to first 5 vulnerabilities to avoid token limits
			vulnSummary += fmt.Sprintf("- %s (%s): %s\n", vuln.Type, vuln.Severity, vuln.Description)
		}
	}

	return fmt.Sprintf(`Analyze the following AI agent prompt and provide improvement suggestions based on the identified vulnerabilities:

PROMPT NAME: %s
AGENT TYPE: %s
LOCATION: %s
DESCRIPTION: %s

CURRENT PROMPT:
%s

IDENTIFIED VULNERABILITIES:
%s

Please provide:
1. A specific improved version of the prompt
2. Clear reasoning for the changes
3. Assessment of the improvement's impact
4. Confidence level (0.0-1.0)

Format your response as JSON with the following structure:
{
  "suggested_prompt": "improved prompt text",
  "reasoning": "explanation of changes and why they help",
  "impact": "description of expected improvement",
  "confidence": 0.8
}`, promptName, promptInfo.AgentType, promptInfo.Location, promptInfo.Description, promptInfo.Prompt, vulnSummary)
}

// parseAISuggestion parses AI response to extract suggestion
func parseAISuggestion(promptName, currentPrompt, aiResponse string, vulnerabilities []Vulnerability) *PromptSuggestion {
	// Try to extract JSON from AI response
	var suggestionData struct {
		SuggestedPrompt string  `json:"suggested_prompt"`
		Reasoning       string  `json:"reasoning"`
		Impact          string  `json:"impact"`
		Confidence      float64 `json:"confidence"`
	}

	// Look for JSON in the response
	jsonStart := strings.Index(aiResponse, "{")
	jsonEnd := strings.LastIndex(aiResponse, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := aiResponse[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &suggestionData); err == nil {
			// Extract vulnerability types
			vulnTypes := make([]string, 0)
			vulnTypeMap := make(map[string]bool)
			for _, vuln := range vulnerabilities {
				if !vulnTypeMap[vuln.Type] {
					vulnTypes = append(vulnTypes, vuln.Type)
					vulnTypeMap[vuln.Type] = true
				}
			}

			// Determine severity based on vulnerabilities
			severity := "low"
			for _, vuln := range vulnerabilities {
				if vuln.Severity == "high" {
					severity = "high"
					break
				} else if vuln.Severity == "medium" {
					severity = "medium"
				}
			}

			return &PromptSuggestion{
				PromptName:         promptName,
				CurrentPrompt:      currentPrompt,
				SuggestedPrompt:    suggestionData.SuggestedPrompt,
				Reasoning:          suggestionData.Reasoning,
				VulnerabilityTypes: vulnTypes,
				Severity:           severity,
				Confidence:         suggestionData.Confidence,
				Impact:             suggestionData.Impact,
			}
		}
	}

	// Fallback: create a basic suggestion if JSON parsing fails
	return &PromptSuggestion{
		PromptName:         promptName,
		CurrentPrompt:      currentPrompt,
		SuggestedPrompt:    currentPrompt, // Keep original if parsing fails
		Reasoning:          "AI response could not be parsed. Manual review recommended.",
		VulnerabilityTypes: []string{"parsing_error"},
		Severity:           "low",
		Confidence:         0.1,
		Impact:             "Unknown - manual review required",
	}
}

// createSuggestionsReport creates the complete suggestions report
func createSuggestionsReport(evaluationFile, promptConfigFile string, results *EvaluationResults, suggestions []PromptSuggestion) *SuggestionsReport {
	// Create vulnerability summary
	vulnSummary := make(map[string]int)
	for _, vuln := range results.Vulnerabilities {
		vulnSummary[vuln.Type]++
	}

	// Generate overall recommendations
	recommendations := []string{
		"Review all high and medium severity vulnerabilities",
		"Implement prompt improvements with highest confidence scores first",
		"Test improved prompts with additional evaluation runs",
		"Monitor performance metrics after implementing changes",
	}

	// Create analysis summary
	summary := fmt.Sprintf("Analyzed %d vulnerabilities across %d prompts. Generated %d specific improvement suggestions. Focus on addressing %d high-severity and %d medium-severity issues.",
		len(results.Vulnerabilities),
		len(suggestions),
		len(suggestions),
		vulnSummary["high"],
		vulnSummary["medium"])

	return &SuggestionsReport{
		GeneratedAt:            time.Now().Format(time.RFC3339),
		EvaluationFile:         evaluationFile,
		PromptConfigFile:       promptConfigFile,
		TotalVulnerabilities:   len(results.Vulnerabilities),
		VulnerabilitySummary:   vulnSummary,
		Suggestions:            suggestions,
		OverallRecommendations: recommendations,
		AnalysisSummary:        summary,
	}
}

func init() {
	rootCmd.AddCommand(suggestionsCmd)
}
