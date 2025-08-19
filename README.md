# Asgard Local MCP Server

A local MCP (Model Control Protocol) server that provides a stdio interface to interact with MCP clients like Claude Desktop.

## Overview

This server acts as a proxy that connects to remote MCP-compatible API endpoints and exposes their tool capabilities through a stdio interface compatible with the MCP protocol.

## Requirements

- Go 1.24 or later

## Installation

```bash
go install github.com/asgard-ai-platform/asgard-mcp-server/cmd/asgard-mcp-server@latest
```

Alternatively, clone the repo and build manually:

```bash
git clone https://github.com/asgard-ai-platform/asgard-mcp-server.git
cd asgard-mcp-server
go build -o asgard-mcp-server ./cmd/asgard-mcp-server
```

### Download from GitHub Releases

You can also download pre-built binaries from the [GitHub Releases](https://github.com/asgard-ai-platform/asgard-mcp-server/releases) page:

1. Go to the [Releases](https://github.com/asgard-ai-platform/asgard-mcp-server/releases) page
2. Download the appropriate archive for your platform:
   - `asgard-mcp-server_Linux_x86_64.tar.gz` for Linux
   - `asgard-mcp-server_Darwin_x86_64.tar.gz` for macOS Intel
   - `asgard-mcp-server_Darwin_arm64.tar.gz` for macOS Apple Silicon
   - `asgard-mcp-server_Windows_x86_64.zip` for Windows

3. Extract the archive:
   ```bash
   # For Linux/macOS
   tar -xzf asgard-mcp-server_*.tar.gz
   
   # For Windows
   unzip asgard-mcp-server_*.zip
   ```

4. **For macOS users**: Remove the quarantine attribute to allow execution:
   ```bash
   xattr -d com.apple.quarantine asgard-mcp-server
   ```

5. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x asgard-mcp-server
   ```

6. Optionally, move the binary to your PATH:
   ```bash
   # Linux/macOS
   sudo mv asgard-mcp-server /usr/local/bin/
   
   # Or add to your user's bin directory
   mkdir -p ~/bin
   mv asgard-mcp-server ~/bin/
   ```

## Usage

Run the server with the required parameters:

```bash
asgard-mcp-server --endpoint <endpoint-url> --api-key <api-key>
```

Example:

```bash
asgard-mcp-server --endpoint "https://api.asgard-ai.com/ns/your-asgard-name-space/toolset/your-asgard-toolset-1/manifest" --api-key "YOUR_ASGARD_API_KEY"
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
3. Point Claude Desktop to the running instance of asgard-mcp-server

## Protocol

This implementation is based on the MCP (Model Control Protocol) specification using the [mcp-go](https://github.com/mark3labs/mcp-go) library v0.36.0.

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
