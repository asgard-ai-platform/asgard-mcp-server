package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient handles API requests to the MCP asgard-mcp-server
type APIClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// Tool represents a tool from the API
type Tool struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	InputSchema    json.RawMessage `json:"input_schema"`
	InvokeEndpoint string          `json:"invoke_endpoint,omitempty"`
}

// ToolsetManifest represents the response from the toolset manifest endpoint
type ToolsetManifest struct {
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Generation int    `json:"generation"`
	Tools      []Tool `json:"tools"`
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchToolsetManifest fetches the toolset manifest from the endpoint
func (c *APIClient) FetchToolsetManifest() (*ToolsetManifest, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		IsSuccess bool `json:"isSuccess"`
		Data      struct {
			Namespace  string `json:"namespace"`
			Name       string `json:"name"`
			Generation int    `json:"generation"`
			Tools      []struct {
				Name           string          `json:"name"`
				Description    string          `json:"description"`
				InputSchema    json.RawMessage `json:"input_schema"`
				InvokeEndpoint string          `json:"invoke_endpoint"`
			} `json:"tools"`
		} `json:"data"`
		Error     *string `json:"error"`
		ErrorCode *string `json:"errorCode"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if !response.IsSuccess {
		errMsg := "unknown error"
		if response.Error != nil {
			errMsg = *response.Error
		}
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	// Create toolset manifest with converted tools
	manifest := &ToolsetManifest{
		Namespace:  response.Data.Namespace,
		Name:       response.Data.Name,
		Generation: response.Data.Generation,
		Tools:      make([]Tool, 0, len(response.Data.Tools)),
	}

	// Convert tools
	for _, t := range response.Data.Tools {
		tool := Tool{
			Name:           t.Name,
			Description:    t.Description,
			InputSchema:    t.InputSchema,
			InvokeEndpoint: t.InvokeEndpoint,
		}
		manifest.Tools = append(manifest.Tools, tool)
	}

	return manifest, nil
}

// ExecuteToolRequest executes a tool request by making an HTTP request to the invoke endpoint
func (c *APIClient) ExecuteToolRequest(tool *Tool, input json.RawMessage) (json.RawMessage, error) {
	// Use the invoke endpoint from the tool definition
	if tool.InvokeEndpoint == "" {
		return nil, fmt.Errorf("tool %s has no invoke endpoint", tool.Name)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", tool.InvokeEndpoint, bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse Asgard MCP response format
	var asgardResponse struct {
		IsSuccess bool            `json:"isSuccess"`
		Data      json.RawMessage `json:"data"`
		Error     *string         `json:"error"`
		ErrorCode *string         `json:"errorCode"`
	}

	if err := json.Unmarshal(body, &asgardResponse); err != nil {
		// If it's not in the Asgard format, return the raw response
		return body, nil
	}

	// Check for API errors
	if !asgardResponse.IsSuccess {
		errMsg := "unknown error"
		if asgardResponse.Error != nil {
			errMsg = *asgardResponse.Error
		}
		return nil, fmt.Errorf("API error: %s", errMsg)
	}

	// Return the data portion of the response
	if asgardResponse.Data != nil {
		return asgardResponse.Data, nil
	}

	// If no data but success is true, return the original body
	return body, nil
}
