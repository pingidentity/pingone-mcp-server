# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Build and Run
```bash
# Build the server
go build -o dist/pingone-mcp-server ./cmd/server

# Run the server (stdio mode is default - perfect for MCP clients)
dist/pingone-mcp-server

# Run the server with debug API logging
dist/pingone-mcp-server --debug-api

# Run tests
go test ./...

# Run tests for a specific package
go test ./pkg/tools

# Run specific test
go test -v ./pkg/tools -run TestRegistryBasics

# Run tests with coverage
go test -cover ./...

# Generate coverage profile and view detailed report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Build with Docker
docker build -t pingone-mcp-server .
```

### Development Environment
Required environment variables must be set before running:
```bash
export PINGONE_CLIENT_ID="your-client-id"
export PINGONE_CLIENT_SECRET="your-client-secret" 
export PINGONE_ENV_ID="your-environment-id"
export PINGONE_REGION="com"  # Optional, defaults to "com"

# Optional configuration
export PINGONE_MCP_TRANSPORT="stdio"         # Transport mode (stdio is default)
export PINGONE_MCP_DEBUG_API="false"         # Enable API request/response logging
export PINGONE_MCP_ALLOW_MUTATION="true"     # Enable write operations
export PINGONE_MCP_ALLOW_INSECURE="true"     # Disable API key requirement
export PINGONE_MCP_SERVER_PORT="8080"        # Server port (HTTP mode only)
export PINGONE_MCP_API_KEY_PATH="./api.key"  # API key file location (HTTP mode only)
```

### Claude Desktop Integration
```bash
# Build for Claude Desktop
go build -o dist/pingone-mcp-server ./cmd/server

# Test stdio mode (default transport)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./pingone-mcp-server

# Run in HTTP mode for REST API access
dist/pingone-mcp-server --transport http --server-port 8080

# Configure Claude Desktop - see CLAUDE_DESKTOP_SETUP.md for detailed instructions
```

### Command-Line Arguments and Environment Variables

Every configuration option can be set via either command-line arguments or environment variables:

| Command-Line Argument | Environment Variable | Default | Description |
|----------------------|---------------------|---------|-------------|
| `--transport` | `PINGONE_MCP_TRANSPORT` | `stdio` | Transport mode: 'stdio' for MCP clients, 'http' for REST API |
| `--debug-api` | `PINGONE_MCP_DEBUG_API` | `false` | Log API requests and responses to PingOne |
| `--allow-mutation` | `PINGONE_MCP_ALLOW_MUTATION` | `false` | Enable mutation tools (create, update, delete operations) |
| `--allow-insecure` | `PINGONE_MCP_ALLOW_INSECURE` | `false` | Disable API key requirement for HTTP mode |
| `--server-port` | `PINGONE_MCP_SERVER_PORT` | `8080` | HTTP server port (HTTP mode only) |
| `--api-key-path` | `PINGONE_MCP_API_KEY_PATH` | `pingone-mcp-server-api.key` | Path to API key file (HTTP mode only) |
| `--client-id` | `PINGONE_CLIENT_ID` | | PingOne OAuth client ID |
| `--client-secret` | `PINGONE_CLIENT_SECRET` | | PingOne OAuth client secret |
| `--env-id` | `PINGONE_ENV_ID` | | PingOne environment ID |
| `--region` | `PINGONE_REGION` | `com` | PingOne region (com, eu, ca, asia) |

**Examples:**
```bash
# Using command-line arguments
dist/pingone-mcp-server --transport http --allow-mutation --server-port 9090

# Using environment variables
export PINGONE_MCP_TRANSPORT=http
export PINGONE_MCP_ALLOW_MUTATION=true
export PINGONE_MCP_SERVER_PORT=9090
dist/pingone-mcp-server

# Mixed (environment variables take precedence unless flags are explicitly set)
export PINGONE_MCP_TRANSPORT=stdio
dist/pingone-mcp-server --transport http  # This will use http, overriding the env var
```

## Architecture

### High-Level Structure
This is a **Model Context Protocol (MCP) server** that provides an HTTP API for PingOne identity management operations. The server acts as a bridge between MCP clients (like Claude) and the PingOne Platform API using the PingOne Go SDK.

### Key Components

**MCP Protocol Layer** (`pkg/mcp/`):
- HTTP router with endpoints: `/mcp/v1/initialize`, `/mcp/v1/tools`, `/mcp/v1/run`
- Request logging and authentication middleware
- JSON marshaling/unmarshaling for MCP protocol messages

**Tool Registry** (`pkg/tools/`):
- Plugin-like architecture where each tool implements the `Tool` interface
- Global registry for tool discovery and execution
- Tools are organized by domain: `identity/`, `environments/`, etc.

**PingOne Client Wrapper** (`pkg/tools/client.go`):
- Abstracts the PingOne Go SDK management client
- Provides a consistent interface for all tools
- Handles SDK response transformation to JSON-friendly maps

**Configuration & Authentication** (`pkg/config/`):
- OpenID Connect discovery for token endpoints
- OAuth2 client credentials flow with automatic token refresh
- Environment-based configuration loading

### Tool Organization
Tools are organized by functional domain:
- **Users**: `identity/users/` - CRUD operations, password management, search
- **Groups**: `identity/groups/` - Group management and membership  
- **Populations**: `environments/populations/` - Population CRUD operations
- **Environments**: `environments/` - Environment lifecycle management
- **Licenses**: Root level - Organization license management

### Security Model
- **Read-only by default**: Mutation tools require `PINGONE_MCP_ALLOW_MUTATION=true`
- **API key authentication**: Random 32-byte hex key generated on first run
- **Insecure mode**: Bypass API key with `PINGONE_MCP_ALLOW_INSECURE=true`
- **Token management**: Automatic OAuth2 token refresh on 401 responses

### Tool Implementation Pattern
Each tool follows this structure:
1. Implements `Tool` interface with `Name()`, `Description()`, `InputSchema()`, `Run()`
2. JSON schema validation for input parameters
3. SDK client calls with error handling
4. Response transformation to JSON-compatible maps
5. Unit tests with mocked SDK clients

### Testing Strategy
- Unit tests for tool logic using mocked PingOne SDK clients
- Registry tests for tool registration and discovery
- HTTP integration tests would target MCP endpoints
- No external API calls in tests (all mocked)