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

// evaluaten8nCmd represents the evaluaten8n command
var evaluaten8nCmd = &cobra.Command{
	Use:   "evaluaten8n [workflow-file]",
	Short: "Evaluate n8n workflow for security vulnerabilities",
	Long: `Evaluate an n8n workflow for security vulnerabilities using AI-powered analysis.
This command works similar to the 'evaluate' command but tests n8n workflows via webhooks.

The workflow file should be a converted n8n workflow (with _eval suffix) that includes webhook nodes.

Example:
  ai-evaluator evaluaten8n n8n/gmail-ai_eval.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workflowFile := args[0]

		// Check if file exists
		if _, err := os.Stat(workflowFile); os.IsNotExist(err) {
			log.Fatalf("Workflow file does not exist: %s", workflowFile)
		}

		// Read agent details configuration
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

		// Initialize n8n workflow evaluator
		evaluator := cloneAttack.NewN8nWorkflowEvaluator(
			ai,
			workflowFile,
			agentDetails.AgentPurpose,
			cloneAttack.TestConfiguration{
				DataLeakageTests:     agentDetails.TestConfiguration.DataLeakageTests,
				PromptInjectionTests: agentDetails.TestConfiguration.PromptInjectionTests,
				ConsistencyTests:     agentDetails.TestConfiguration.ConsistencyTests,
				IterationsPerTest:    agentDetails.TestConfiguration.IterationsPerTest,
			},
		)

		// Run comprehensive vulnerability test
		results, err := evaluator.RunComprehensiveVulnerabilityTest()
		if err != nil {
			log.Fatalf("Evaluation failed: %v", err)
		}

		// Save results
		timestamp := time.Now().Format("20060102_150405")
		resultsFile := fmt.Sprintf("results/n8n_evaluation_results_%s.json", timestamp)

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

		log.Printf("N8N Workflow evaluation completed: %d total calls, %d successful, %d failed",
			results.TotalCalls,
			results.SuccessfulCalls,
			results.FailedCalls)
	},
}

func init() {
	rootCmd.AddCommand(evaluaten8nCmd)
}
