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
	endpointURL := flag.String("endpoint", "", "The endpoint URL for the MCP asgard-mcp-server")
	apiKey := flag.String("api-key", "", "The API key for authentication")

	// Parse flags
	flag.Parse()

	// Validate mandatory parameters
	if *endpointURL == "" || *apiKey == "" {
		fmt.Println("Error: Both endpoint URL and API key are required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize MCP asgard-mcp-server
	server, err := mcp.NewServer(*endpointURL, *apiKey)
	if err != nil {
		log.Fatalf("Failed to create MCP asgard-mcp-server: %v", err)
	}

	// Start the asgard-mcp-server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start MCP asgard-mcp-server: %v", err)
	}
}
