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
	log.SetFlags(log.Lshortfile)
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Error loading .env file")
	}
}
