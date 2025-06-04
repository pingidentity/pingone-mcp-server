# PingOne MCP Server

A **Model Context Protocol (MCP) server** that provides an interface between MCP clients (like Claude) and the PingOne Platform API for identity management operations.

## Overview

This server acts as a bridge between AI assistants and PingOne's identity services, enabling natural language interactions with user management, group administration, environment configuration, and other identity operations. It supports both real-time communication via stdio (for Claude Desktop) and HTTP REST API access.

**Key Features:**
- 🔐 **Secure Authentication** - OAuth2 client credentials flow with automatic token refresh
- 🛡️ **Safety First** - Read-only by default, with optional mutation controls
- 🔧 **Flexible Deployment** - Works with Claude Desktop, HTTP clients, or containerized environments
- 📊 **Comprehensive Coverage** - Supports users, groups, populations, environments, and licenses
- 🐳 **Container Ready** - Docker and Podman support with environment-based configuration

## High-Level Architecture

```
┌─────────────────┐    MCP Protocol     ┌──────────────────┐    HTTP/OAuth2    ┌─────────────────┐
│                 │ ◄──────────────────► │                  │ ◄─────────────────► │                 │
│   Claude AI     │    JSON-RPC over    │  PingOne MCP     │   SDK Requests    │   PingOne API   │
│   (or any MCP   │    stdio/HTTP       │     Server       │                   │                 │
│     client)     │                     │                  │                   │                 │
└─────────────────┘                     └──────────────────┘                   └─────────────────┘
                                                │
                                                │
                                                ▼
                                        ┌──────────────────┐
                                        │   Tool Registry  │
                                        │                  │
                                        │ • User Tools     │
                                        │ • Group Tools    │
                                        │ • Environment    │
                                        │ • Population     │
                                        │ • License Tools  │
                                        └──────────────────┘
```

### Architecture Components

**MCP Protocol Layer** (`pkg/mcp/`):
- **stdio Transport**: Direct communication with Claude Desktop via JSON-RPC over stdin/stdout
- **HTTP Transport**: REST API server with endpoints for web clients and integrations
- **Authentication**: API key-based security for HTTP mode (auto-generated or configurable)

**Tool Registry** (`pkg/tools/`):
- **Plugin Architecture**: Each tool implements a common `Tool` interface
- **Dynamic Discovery**: Tools register themselves and are available via MCP `tools/list`
- **JSON Schema Validation**: Input parameters validated against OpenAPI-style schemas

**PingOne Integration** (`pkg/tools/client.go`):
- **SDK Wrapper**: Abstracts the PingOne Go SDK management client
- **Token Management**: Automatic OAuth2 refresh on 401 responses  
- **Response Transformation**: Converts SDK responses to JSON-friendly maps

**Configuration** (`pkg/config/`):
- **Environment Variables**: All settings configurable via env vars or CLI flags
- **OAuth2 Discovery**: Automatic endpoint discovery via OpenID Connect
- **Multi-Region Support**: Supports com, eu, ca, and asia regions

## Getting Started

### Prerequisites

- **Go 1.21+** for building from source
- **PingOne Account** with OAuth2 client credentials
- **Environment Setup**: Client ID, Client Secret, and Environment ID

### Build and Run

#### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd pingone-mcp-server

# Build the server
go build -o pingone-mcp-server ./cmd/server

# Make executable
chmod +x pingone-mcp-server
```

#### Run with Environment Variables

```bash
# Set required PingOne credentials
export PINGONE_CLIENT_ID="your-client-id"
export PINGONE_CLIENT_SECRET="your-client-secret"
export PINGONE_ENV_ID="your-environment-id"
export PINGONE_REGION="com"  # Optional: com, eu, ca, asia

# Optional: Enable write operations (read-only by default)
export PINGONE_MCP_ALLOW_MUTATION="true"

# Run the server (stdio mode for Claude Desktop)
./pingone-mcp-server
```

#### Run with Command-Line Arguments

```bash
# All configuration via CLI flags
./pingone-mcp-server \
  --client-id "your-client-id" \
  --client-secret "your-client-secret" \
  --env-id "your-environment-id" \
  --region "com" \
  --allow-mutation

# HTTP mode for REST API access
./pingone-mcp-server \
  --transport http \
  --server-port 8080 \
  --allow-mutation
```

#### Development and Testing

```bash
# Run tests
go test ./...

# Run specific package tests
go test ./pkg/tools

# Test the server manually
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./pingone-mcp-server

# Enable API debugging
./pingone-mcp-server --debug-api
```

### Docker Usage

#### Build Docker Image

```bash
# Build the image
docker build -t pingone-mcp-server .

# Or with Podman
podman build -t pingone-mcp-server .
```

#### Run with Docker

```bash
# Run with environment variables
docker run -e PINGONE_CLIENT_ID="your-client-id" \
           -e PINGONE_CLIENT_SECRET="your-client-secret" \
           -e PINGONE_ENV_ID="your-environment-id" \
           -e PINGONE_MCP_ALLOW_MUTATION="true" \
           pingone-mcp-server

# Run in HTTP mode with port mapping
docker run -p 8080:8080 \
           -e PINGONE_CLIENT_ID="your-client-id" \
           -e PINGONE_CLIENT_SECRET="your-client-secret" \
           -e PINGONE_ENV_ID="your-environment-id" \
           pingone-mcp-server --transport http

# Run with CLI arguments
docker run pingone-mcp-server \
  --client-id "your-client-id" \
  --client-secret "your-client-secret" \
  --env-id "your-environment-id" \
  --transport http \
  --allow-mutation
```

#### Run with Podman

```bash
# Podman usage is identical to Docker
podman run -e PINGONE_CLIENT_ID="your-client-id" \
           -e PINGONE_CLIENT_SECRET="your-client-secret" \
           -e PINGONE_ENV_ID="your-environment-id" \
           pingone-mcp-server

# Rootless container with port mapping
podman run -p 8080:8080 \
           -e PINGONE_CLIENT_ID="your-client-id" \
           -e PINGONE_CLIENT_SECRET="your-client-secret" \
           -e PINGONE_ENV_ID="your-environment-id" \
           pingone-mcp-server --transport http
```

#### Docker Compose Example

```yaml
version: '3.8'
services:
  pingone-mcp:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PINGONE_CLIENT_ID=your-client-id
      - PINGONE_CLIENT_SECRET=your-client-secret
      - PINGONE_ENV_ID=your-environment-id
      - PINGONE_MCP_TRANSPORT=http
      - PINGONE_MCP_ALLOW_MUTATION=true
    command: ["--transport", "http", "--server-port", "8080"]
```

## Configuration Reference

| Environment Variable         | CLI Flag           | Default | Description |
|------------------------------|--------------------|---------|-------------|
| `PINGONE_CLIENT_ID`          | `--client-id`      | - | PingOne OAuth client ID |
| `PINGONE_CLIENT_SECRET`      | `--client-secret`  | - | PingOne OAuth client secret |
| `PINGONE_ENV_ID`             | `--env-id`         | - | PingOne environment ID |
| `PINGONE_REGION`             | `--region`         | `com` | PingOne region (com, eu, ca, asia) |
| `PINGONE_MCP_TRANSPORT`      | `--transport`      | `stdio` | Transport: 'stdio' or 'http' |
| `PINGONE_MCP_ALLOW_MUTATION` | `--allow-mutation` | `false` | Enable write operations |
| `PINGONE_MCP_DEBUG_API`      | `--debug-api`      | `false` | Log API requests/responses |
| `PINGONE_MCP_ALLOW_INSECURE` | `--allow-insecure` | `false` | Disable API key (HTTP mode) |
| `PINGONE_MCP_SERVER_PORT`    | `--server-port`    | `8080` | HTTP server port |
| `PINGONE_MCP_API_KEY_PATH`   | `--api-key-path`   | `pingone-mcp-server-api.key` | API key file location |

## Available Tools

### Read-Only Tools (Always Available)
- `get_user` - Retrieve user details by ID
- `get_user_password_state` - Check user password status  
- `get_population` - Get population details
- `get_group` - Get group details
- `get_environment` - Get environment details
- `get_environment_bom` - Get environment bill of materials
- `get_license` - Get organization license details

### Write Tools (Require `PINGONE_MCP_ALLOW_MUTATION=true`)
- `create_user`, `update_user`, `delete_user` - User lifecycle management
- `set_user_enabled`, `reset_user_password`, `unlock_user_password` - User account management
- `add_user_to_group`, `remove_user_from_group` - Group membership management  
- `create_population`, `delete_population` - Population management
- `create_group`, `update_group`, `delete_group` - Group management
- `create_environment`, `delete_environment`, `update_environment_status` - Environment management

## Claude Desktop Integration

For detailed Claude Desktop setup instructions, see [CLAUDE_DESKTOP_SETUP.md](./CLAUDE_DESKTOP_SETUP.md).

Quick setup:
1. Build the server: `go build -o pingone-mcp-server ./cmd/server`
2. Configure `~/Library/Application Support/Claude/claude_desktop_config.json`
3. Add your PingOne credentials to the `env` section
4. Restart Claude Desktop

## Security and Best Practices

- **Read-Only by Default**: Write operations require explicit enablement
- **Secure Credentials**: Store PingOne credentials securely, never commit to version control
- **Network Security**: Use HTTPS in production, consider API key rotation
- **Principle of Least Privilege**: Use environment-specific credentials with minimal required permissions
- **Audit Logging**: Enable `--debug-api` for audit trails in development

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

[License information would go here]

## Support

- **Documentation**: See [CLAUDE.md](./CLAUDE.md) for development guidance
- **Issues**: Report bugs and feature requests via GitHub Issues
- **PingOne Documentation**: [PingOne Platform API](https://apidocs.pingidentity.com/pingone/platform/v1/api/)