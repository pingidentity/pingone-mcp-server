# Docker Usage Instructions

## Getting Started with Docker

This document provides instructions for using the PingOne MCP server as a Docker OCI container. The Docker deployment method is designed for environments where installing the binary directly is not practical or desired.

> [!IMPORTANT]
> **Requires MCP client URL mode elicitation support**
>
> The Docker method can only be used with MCP clients that support [URL mode elicitation](https://modelcontextprotocol.io/specification/2025-11-25/client/elicitation#url-mode-elicitation-for-oauth-flows) (introduced in the 2025-11-25 MCP specification). This capability is essential for securely providing the authorization URL to the user during device mode authentication, ensuring the URL is only presented to the human user and not presented to be processed by the AI agent.

### Prerequisites

- **A licensed or trial PingOne cloud subscription.** - Don't have a tenant? [Sign up for a free trial here](https://www.pingidentity.com/en/try-ping.html).
- **MCP-compatible client that supports URL mode elicitation** (E.g. VS Code Copilot Chat)
- **Docker install**

### Prepare PingOne for MCP Server Use

The MCP server requires a worker application in your PingOne tenant to access the management APIs. To use Docker, the worker application must be configured using the Device Authorization Grant. You'll need to capture two values during setup:

- **Environment ID** - The PingOne environment containing your worker application (referred to later as `{{admin environment id}}`)
- **Client ID** - The worker application's client identifier (referred to later as `{{mcp application client id}}`)

#### Using the Device Authorization Grant

For headless or containerized environments, use the Device Authorization grant. Configure your worker application with:

- **Grant Type**: Device Authorization
- **Token Endpoint Authentication**: None
- **Redirect URI**: `http://127.0.0.1:7464/callback`
- **Application Roles**: None required (the MCP server inherits roles from the authenticated user)

This grant type is ideal for environments without a browser, such as CI/CD pipelines or remote servers.

### Use with VS Code

[![Install in VS Code](https://img.shields.io/badge/VS_Code-Install_Server-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=pingOne&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_environment_id%22%2C%22description%22%3A%22The%20environment%20ID%20containing%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_mcp_client_id%22%2C%22description%22%3A%22The%20client%20ID%20of%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_api_root_domain%22%2C%22description%22%3A%22The%20root%20domain%20of%20your%20PingOne%20tenant%20%28e.g.%2C%20%60pingone.com%60%20%2C%20%60pingone.eu%60%20%2C%20%60pingone.ca%60%29%22%2C%22password%22%3Afalse%7D%5D&config=%7B%22type%22%3A%22stdio%22%2C%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22-i%22%2C%22--rm%22%2C%22-e%22%2C%22PINGONE_MCP_ENVIRONMENT_ID%22%2C%22-e%22%2C%22PINGONE_DEVICE_CODE_CLIENT_ID%22%2C%22-e%22%2C%22PINGONE_ROOT_DOMAIN%22%2C%22pingidentity%2Fpingone-mcp-server%3Alatest%22%5D%2C%22env%22%3A%7B%22PINGONE_MCP_ENVIRONMENT_ID%22%3A%22%24%7Binput%3Apingone_environment_id%7D%22%2C%22PINGONE_DEVICE_CODE_CLIENT_ID%22%3A%22%24%7Binput%3Apingone_mcp_client_id%7D%22%2C%22PINGONE_ROOT_DOMAIN%22%3A%22%24%7Binput%3Apingone_api_root_domain%7D%22%7D%7D) [![Install in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Install_Server-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=pingOne&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_environment_id%22%2C%22description%22%3A%22The%20environment%20ID%20containing%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_mcp_client_id%22%2C%22description%22%3A%22The%20client%20ID%20of%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_api_root_domain%22%2C%22description%22%3A%22The%20root%20domain%20of%20your%20PingOne%20tenant%20%28e.g.%2C%20%60pingone.com%60%20%2C%20%60pingone.eu%60%20%2C%20%60pingone.ca%60%29%22%2C%22password%22%3Afalse%7D%5D&config=%7B%22type%22%3A%22stdio%22%2C%22command%22%3A%22docker%22%2C%22args%22%3A%5B%22run%22%2C%22-i%22%2C%22--rm%22%2C%22-e%22%2C%22PINGONE_MCP_ENVIRONMENT_ID%22%2C%22-e%22%2C%22PINGONE_DEVICE_CODE_CLIENT_ID%22%2C%22-e%22%2C%22PINGONE_ROOT_DOMAIN%22%2C%22pingidentity%2Fpingone-mcp-server%3Alatest%22%5D%2C%22env%22%3A%7B%22PINGONE_MCP_ENVIRONMENT_ID%22%3A%22%24%7Binput%3Apingone_environment_id%7D%22%2C%22PINGONE_DEVICE_CODE_CLIENT_ID%22%3A%22%24%7Binput%3Apingone_mcp_client_id%7D%22%2C%22PINGONE_ROOT_DOMAIN%22%3A%22%24%7Binput%3Apingone_api_root_domain%7D%22%7D%7D&quality=insiders)

For quick installation, use one of the install buttons above.

To add the MCP server configuration manually, add the following configuration to your MCP configuration file:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "PINGONE_MCP_ENVIRONMENT_ID",
        "-e",
        "PINGONE_DEVICE_CODE_CLIENT_ID",
        "-e",
        "PINGONE_ROOT_DOMAIN",
        "pingidentity/pingone-mcp-server:latest",
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "${input:pingone_environment_id}",
        "PINGONE_DEVICE_CODE_CLIENT_ID": "${input:pingone_mcp_client_id}",
        "PINGONE_ROOT_DOMAIN": "${input:pingone_api_root_domain}",
      }
    },
  },
  "inputs": [
    {
      "type": "promptString",
      "id": "pingone_environment_id",
      "description": "The environment ID containing the MCP server worker application",
      "password": false
    },
    {
      "type": "promptString",
      "id": "pingone_mcp_client_id",
      "description": "The client ID of the MCP server worker application",
      "password": false
    },
    {
      "type": "promptString",
      "id": "pingone_api_root_domain",
      "description": "The root domain of your PingOne tenant (e.g., `pingone.com` , `pingone.eu` , `pingone.ca`)",
      "password": false
    }
  ]
}
```

Once installed, ensure agent mode is turned on and the MCP server has started, then make the first request!

### Use with Claude Desktop

At the time of writing, Claude Desktop does not support URL mode elicitation.  Use the [binary install method](./../README.md#install-the-mcp-server) instead which allows the browser to open automatically.

### Use with Claude Code

At the time of writing, Claude Code does not support URL mode elicitation.  Use the [binary install method](./../README.md#install-the-mcp-server) instead which allows the browser to open automatically.

### Use with Cursor

At the time of writing, Cursor does not support URL mode elicitation.  Use the [binary install method](./../README.md#install-the-mcp-server) instead which allows the browser to open automatically.

### Building from Source

If you'd like to build and run the docker image from source, use `make docker-build` which will compile the code and build the docker image.

You can run the docker image with the following command:

```shell
```

When configuring MCP clients, ensure that the `args` value refers to the development image tag `pingone-mcp-server:dev`.  For example:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "PINGONE_MCP_ENVIRONMENT_ID",
        "-e",
        "PINGONE_DEVICE_CODE_CLIENT_ID",
        "-e",
        "PINGONE_ROOT_DOMAIN",
        "pingone-mcp-server:dev",
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "${input:pingone_environment_id}",
        "PINGONE_DEVICE_CODE_CLIENT_ID": "${input:pingone_mcp_client_id}",
        "PINGONE_ROOT_DOMAIN": "${input:pingone_api_root_domain}",
      }
    }
  }
}
```

## Authentication and Authorization

The server, when using Docker containers uses **OAuth 2.0 Device Code authorization flow with PKCE** for secure administrator authentication by default. The OAuth 2.0 Authorization Code flow is not available when using the MCP server as a Docker container as the admin's browser cannot be automatically opened.

1. **First Tool Use** - Uses URL mode elicitation to present the authorization URL securely to the human user, allowing the administrator to login to the configured PingOne tenant when a tool is used for the first time in a session
2. **Token Storage** - Access tokens stored securely within the container and not shared with the agent
3. **Automatic Reuse** - Cached tokens used for subsequent tool calls within the same session
