# Asgard Local MCP Server

A local MCP (Model Control Protocol) server that provides a stdio interface to interact with MCP clients like Claude Desktop.

## Overview

This server acts as a proxy that connects to remote MCP-compatible API endpoints and exposes their tool capabilities through a stdio interface compatible with the MCP protocol.

## Requirements

- Go 1.21 or later

## Installation

```bash
go install github.com/asgard-ai-platform/asgard-mcp-asgard-mcp-server/cmd/asgard-mcp-server@latest
```

Alternatively, clone the repo and build manually:

```bash
git clone https://github.com/asgard-ai-platform/asgard-mcp-server.git
cd asgard-mcp-asgard-mcp-server
go build -o asgard-mcp-asgard-mcp-server ./cmd/asgard-mcp-server
```

## Usage

Run the server with the required parameters:

```bash
asgard-mcp-asgard-mcp-server --endpoint <endpoint-url> --api-key <api-key>
```

Example:

```bash
asgard-mcp-asgard-mcp-server --endpoint "https://api.asgard-ai.com/ns/your-asgard-name-space/toolset/your-asgard-toolset-1/manifest" --api-key "YOUR_ASGARD_API_KEY"
```

The server will:

1. Connect to the specified endpoint
2. Fetch the toolset manifest
3. Process MCP requests via stdio
4. Forward tool invocation requests to the appropriate endpoints

### Integrating with Claude Desktop

To use this server with Claude Desktop:

1. Run the server with the appropriate parameters
2. In Claude Desktop, configure the MCP provider to use the stdio interface
3. Point Claude Desktop to the running instance of asgard-local-mcp

## Protocol

This implementation is based on the MCP (Model Control Protocol) specification using the [mcp-go](https://github.com/mark3labs/mcp-go) library v0.27.0.

## Testing

A simple test script is provided to verify that the server is working correctly:

```bash
# Make the test script executable
chmod +x test.sh

# Run the test
./test.sh
```

## Development

This project uses Go modules for dependency management. To add a new dependency, use:

```bash
go get <dependency>
```

To update dependencies, use:

```bash
go mod tidy
```
