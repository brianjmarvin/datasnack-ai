/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

// rootCmd represents the main command for the DataSnack CLI application
var rootCmd = &cobra.Command{
	Use:   "datasnack",
	Short: "Analyze code for security vulnerabilities",
	Long: `DataSnack is a security analysis tool that performs comprehensive 
vulnerability testing on your codebase using AI-powered analysis.

It reads configuration from environment variables and outputs results 
to attackResults.json.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Load required environment variables
	// 	endpoint := os.Getenv("TEST_ENDPOINT")
	// 	apiKey := os.Getenv("TEST_API_KEY")
	// 	headerKey := os.Getenv("HEADER_KEY")
	// 	schemaPath := os.Getenv("SCHEMA_FILE")

	// 	// Initialize AWS Bedrock AI client
	// 	ai := awsbedrock.New()

	// 	// Read schema file
	// 	log.Println("Reading schema from:", schemaPath)
	// 	schema, err := os.ReadFile(schemaPath)
	// 	if err != nil {
	// 		log.Fatalln("Failed to read schema file:", err)
	// 	}

	// 	// Initialize and run security analysis
	// 	analyzer := cloneAttack.NewCloneAttack(ai, endpoint, headerKey, apiKey, string(schema))
	// 	results, err := analyzer.RunComprehensiveVulnerabilityTest()
	// 	if err != nil {
	// 		log.Fatalln("Analysis failed:", err)
	// 	}

	// 	// Save results to JSON file
	// 	resultsJSON, err := json.Marshal(results)
	// 	if err != nil {
	// 		log.Println("Failed to marshal results:", err)
	// 		return
	// 	}

	// 	if err := os.WriteFile("attackResults.json", resultsJSON, 0644); err != nil {
	// 		log.Println("Failed to write results:", err)
	// 	}
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().String("path", "", "used to create emails for DataSnack")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.datasnack.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	log.SetFlags(log.Lshortfile)
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading - in AWS servermode")
	}
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
