package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the local MCP asgard-mcp-server
type Server struct {
	endpointURL string
	apiKey      string
	tools       []Tool
	mutex       sync.RWMutex
	apiClient   *APIClient
	mcpServer   *server.MCPServer
}

// NewServer creates a new MCP asgard-mcp-server with the provided endpoint URL and API key
func NewServer(endpointURL, apiKey string) (*Server, error) {
	// Create the asgard-mcp-server
	s := &Server{
		endpointURL: endpointURL,
		apiKey:      apiKey,
	}

	// Create API client
	s.apiClient = NewAPIClient(endpointURL, apiKey)

	// Fetch the toolset manifest
	manifest, err := s.apiClient.FetchToolsetManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch toolset manifest: %w", err)
	}

	// Store tools from manifest
	s.mutex.Lock()
	s.tools = manifest.Tools
	s.mutex.Unlock()

	// Create hooks for logging
	hooks := &server.Hooks{}

	// Add hook to log incoming requests
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("[RPC] Received method: %s", method)
	})

	// Add hook to log successful responses
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Printf("[RPC] Response for method %s (failed to marshal)", method)
			return
		}
		log.Printf("[RPC] Response for method %s: %s", method, string(resultJSON))
	})

	// Add hook to log errors
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("[RPC] Error for method %s: %v", method, err)
	})

	// Add detailed logging for tool call requests
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		// Marshal tool arguments for detailed logging
		argsJSON, err := json.MarshalIndent(message.Params.Arguments, "", "  ")
		if err != nil {
			log.Printf("[RPC-TOOL] Call to tool '%s' with arguments (failed to marshal)", message.Params.Name)
			return
		}
		log.Printf("[RPC-TOOL] Call to tool '%s' with arguments: %s", message.Params.Name, string(argsJSON))
	})

	// Add detailed logging for tool call responses
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		switch {
		case result.IsError:
			log.Printf("[RPC-TOOL] Tool '%s' response (error)", message.Params.Name)
		case len(result.Content) > 0:
			// Log first content item type
			switch content := result.Content[0].(type) {
			case mcp.TextContent:
				log.Printf("[RPC-TOOL] Tool '%s' response (text): %s", message.Params.Name, content.Text)
			case mcp.ImageContent:
				log.Printf("[RPC-TOOL] Tool '%s' response (image): %s", message.Params.Name, content.MIMEType)
			case mcp.AudioContent:
				log.Printf("[RPC-TOOL] Tool '%s' response (audio): %s", message.Params.Name, content.MIMEType)
			case mcp.EmbeddedResource:
				log.Printf("[RPC-TOOL] Tool '%s' response (resource)", message.Params.Name)
			default:
				log.Printf("[RPC-TOOL] Tool '%s' response (unknown content type)", message.Params.Name)
			}
		default:
			log.Printf("[RPC-TOOL] Tool '%s' response (empty)", message.Params.Name)
		}
	})

	// Create MCP asgard-mcp-server with options
	s.mcpServer = server.NewMCPServer(
		"asgard-mcp-asgard-mcp-server",
		"0.0.1",
		server.WithToolCapabilities(true),
		server.WithHooks(hooks),
		server.WithLogging(),
	)

	// Register tool handlers
	if err := s.registerToolHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register tool handlers: %w", err)
	}

	return s, nil
}

// Start starts the MCP asgard-mcp-server, handling stdin/stdout communication
func (s *Server) Start() error {
	log.Println("Starting MCP asgard-mcp-server...")
	log.Printf("Endpoint: %s", s.endpointURL)

	s.mutex.RLock()
	log.Printf("Available tools: %d", len(s.tools))
	for _, tool := range s.tools {
		log.Printf("  - %s: %s", tool.Name, tool.Description)
	}
	s.mutex.RUnlock()

	// Create the stdio asgard-mcp-server
	stdioServer := server.NewStdioServer(s.mcpServer)

	// Set up error logging
	stdioServer.SetErrorLogger(log.New(os.Stderr, "[ERROR] ", log.LstdFlags))

	// Start the asgard-mcp-server
	return server.ServeStdio(s.mcpServer)
}

// registerToolHandlers registers all tools from the manifest with the MCP asgard-mcp-server
func (s *Server) registerToolHandlers() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Register handlers for each tool
	for _, tool := range s.tools {
		// Create a local copy of the tool to avoid closure issues
		localTool := tool

		// Define a handler for the tool
		handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Create the arguments JSON
			argsJSON, err := json.Marshal(req.Params.Arguments)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal arguments: %v", err)), nil
			}

			// Log API call
			log.Printf("[API-CALL] Executing tool '%s'", localTool.Name)

			// Execute the tool request
			// The APIClient.ExecuteToolRequest method now handles the Asgard response format
			// and returns the "data" field content when applicable
			responseJSON, err := s.apiClient.ExecuteToolRequest(&localTool, argsJSON)
			if err != nil {
				log.Printf("[API-CALL] Tool '%s' execution failed: %v", localTool.Name, err)
				return mcp.NewToolResultError(fmt.Sprintf("Tool execution failed: %v", err)), nil
			}

			log.Printf("[API-CALL] Tool '%s' response received: %d bytes", localTool.Name, len(responseJSON))

			// Parse the response
			var responseObj interface{}
			if err := json.Unmarshal(responseJSON, &responseObj); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to parse tool response: %v", err)), nil
			}

			// Format the response as indented JSON for readability
			responseText, err := json.MarshalIndent(responseObj, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to format tool response: %v", err)), nil
			}

			return mcp.NewToolResultText(string(responseText)), nil
		}

		// Create an MCP Tool definition
		mcpTool := mcp.Tool{
			Name:        localTool.Name,
			Description: localTool.Description,
		}

		// Convert input schema from JSON to ToolInputSchema
		var schema map[string]interface{}
		if err := json.Unmarshal(localTool.InputSchema, &schema); err != nil {
			return fmt.Errorf("failed to parse input schema for tool %s: %w", localTool.Name, err)
		}

		if schema != nil {
			// Ensure the schema has the required 'type' field set to 'object'
			if _, ok := schema["type"]; !ok {
				schema["type"] = "object"
			}
			if tool.AllowUploadFiles {
				// Ensue the schema has the required 'properties' field
				if _, ok := schema["properties"]; !ok {
					schema["properties"] = make(map[string]interface{})
				}
				// Append the UploadedFilePaths field if the tool allows file uploads
				if props, ok := schema["properties"].(map[string]interface{}); ok {
					props[UploadedFilePathsFieldName] = UploadedFilePathsSchema
				}
			}
		}

		// Convert schema back to JSON for the tool definition
		updatedSchema, err := json.Marshal(schema)
		if err != nil {
			return fmt.Errorf("failed to marshal updated input schema for tool %s: %w", localTool.Name, err)
		}

		// Set the RawInputSchema to the modified schema
		mcpTool.RawInputSchema = updatedSchema

		// Register the tool with the asgard-mcp-server
		s.mcpServer.AddTool(mcpTool, handler)
	}

	return nil
}
