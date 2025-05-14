package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/asgard-ai-platform/asgard-mcp-server/pkg/mcp"
)

func main() {
	// Define flags for endpoint URL and API key
	endpointURL := flag.String("endpoint", "", "The endpoint URL for the MCP server")
	apiKey := flag.String("api-key", "", "The API key for authentication")

	// Parse flags
	flag.Parse()

	// Validate mandatory parameters
	if *endpointURL == "" || *apiKey == "" {
		fmt.Println("Error: Both endpoint URL and API key are required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize MCP server
	server, err := mcp.NewServer(*endpointURL, *apiKey)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}
