/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"
	"os"

	"datasnack/cmd"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Check if we should automatically launch the server
	launchServer := os.Getenv("LAUNCH_SERVER")
	if launchServer == "true" {
		log.Println("LAUNCH_SERVER=true detected, automatically starting server...")

		// Get server port from environment or use default
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}

		// Set up arguments for server command
		os.Args = []string{os.Args[0], "server", port}
		log.Printf("Starting server on port %s", port)
	}

	cmd.Execute()
}
