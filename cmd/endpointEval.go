package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"datasnack/cloneAttack"

	"github.com/spf13/cobra"
)

// endpointEvalCmd represents the endpointEval command
var endpointEvalCmd = &cobra.Command{
	Use:   "endpointEval [yaml-config-file]",
	Short: "Evaluate any HTTP endpoint with YAML-defined schema",
	Long: `Evaluate any HTTP endpoint for security vulnerabilities using AI-powered analysis.
This command works like the 'evaluate' command but takes a YAML file that defines
the full endpoint URL and the schema for the payload.

The YAML file should contain:
- service: base URL and service information
- endpoints: endpoint definitions with paths, methods, and request schemas
- The same test configuration as other evaluation commands

Example:
  ai-evaluator endpointEval config/my-endpoint-config.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		yamlConfigFile := args[0]

		// Check if file exists
		if _, err := os.Stat(yamlConfigFile); os.IsNotExist(err) {
			log.Fatalf("YAML config file does not exist: %s", yamlConfigFile)
		}

		// Read agent details configuration (same as evaluaten8n)
		log.Println("Reading agent details from: config/agentDetails.json")
		agentDetailsData, err := os.ReadFile("config/agentDetails.json")
		if err != nil {
			log.Fatalf("Failed to read agent details: %v", err)
		}

		var agentDetails struct {
			AgentPurpose      string `json:"agentPurpose"`
			TestConfiguration struct {
				DataLeakageTests     int `json:"dataLeakageTests"`
				PromptInjectionTests int `json:"promptInjectionTests"`
				ConsistencyTests     int `json:"consistencyTests"`
				IterationsPerTest    int `json:"iterationsPerTest"`
			} `json:"testConfiguration"`
		}

		if err := json.Unmarshal(agentDetailsData, &agentDetails); err != nil {
			log.Fatalf("Failed to parse agent details: %v", err)
		}

		// Initialize AI client (same logic as serve.go)
		ai, err := initializeAIClient()
		if err != nil {
			log.Fatalf("Failed to initialize AI client: %v", err)
		}

		// Initialize endpoint evaluator
		evaluator, err := cloneAttack.NewEndpointEvaluator(
			ai,
			yamlConfigFile,
			agentDetails.AgentPurpose,
			cloneAttack.TestConfiguration{
				DataLeakageTests:     agentDetails.TestConfiguration.DataLeakageTests,
				PromptInjectionTests: agentDetails.TestConfiguration.PromptInjectionTests,
				ConsistencyTests:     agentDetails.TestConfiguration.ConsistencyTests,
				IterationsPerTest:    agentDetails.TestConfiguration.IterationsPerTest,
			},
		)
		if err != nil {
			log.Fatalf("Failed to initialize endpoint evaluator: %v", err)
		}

		// Run comprehensive vulnerability test
		results, err := evaluator.RunComprehensiveVulnerabilityTest()
		if err != nil {
			log.Fatalf("Evaluation failed: %v", err)
		}

		// Save results
		timestamp := time.Now().Format("20060102_150405")
		resultsFile := fmt.Sprintf("results/endpoint_evaluation_results_%s.json", timestamp)

		resultsJSON, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal results: %v", err)
		} else {
			if err := os.WriteFile(resultsFile, resultsJSON, 0644); err != nil {
				log.Printf("Failed to write results: %v", err)
			} else {
				log.Printf("Results saved to: %s", resultsFile)
			}
		}

		log.Printf("Endpoint evaluation completed: %d total calls, %d successful, %d failed",
			results.TotalCalls,
			results.SuccessfulCalls,
			results.FailedCalls)
	},
}

func init() {
	rootCmd.AddCommand(endpointEvalCmd)
}
