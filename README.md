# PingOne MCP Server

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/pingidentity/pingone-mcp-server?include_prereleases&sort=semver)](https://github.com/pingidentity/pingone-mcp-server/releases)
[![Go Security Scan](https://github.com/pingidentity/pingone-mcp-server/actions/workflows/gosec-scan.yml/badge.svg)](https://github.com/pingidentity/pingone-mcp-server/actions/workflows/gosec-scan.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pingidentity/pingone-mcp-server)](https://goreportcard.com/report/github.com/pingidentity/pingone-mcp-server)

The PingOne MCP (Model Context Protocol) server enables AI assistants to review and manage PingOne tenants by integrating the PingOne management API to AI assistant conversations.

> [!CAUTION]
> **Preview Software Notice**
>
> This is preview software provided AS IS with no warranties of any kind.
>
> - Functionality, features, and APIs are subject to change at any time without prior notice
> - Use against production environments or mission-critical workloads is not advised
> - Limited support is available during the public preview phase — please report bugs and provide feedback via the [GitHub issue tracker](https://github.com/pingidentity/pingone-mcp-server/issues)
>
> Your use of this software constitutes acceptance of these terms.

> [!CAUTION]
> **Security Notice**
>
> Depending on the requests made to the MCP server, tenant configuration or data may be returned.  Do not use the MCP server with untrusted MCP clients, agent code or LLM inference and ensure least privilege principles are followed when granting role permissions to MCP server users.

> [!WARNING]
> **Review Generated Configuration**
>
> Configuration can be generated dynamically using LLM and user feedback represented dynamically back to agents/conversations. Be sure to review generated configuration before promoting to production environments, or those serving live identity/access requests.

## Features

- **Administer your PingOne environment using natural language** - Interact with PingOne from whichever AI IDE or MCP client tool you use daily.

- **Secure authentication** - Supports OAuth 2.0 PKCE flow for local deployment and Device Code Flow for containerized deployment. All actions are user-based and auditable. Tokens stored securely in OS keychain (local) or ephemerally (Docker).

- **Environment, application and population operations** - Provides tool integrations to create, update and analyze configurations for tenant activities.

## Use Cases

This MCP server is designed to help developers integrate PingOne capabilities into their applications, while also helping tenant administrators monitor and troubleshoot issues. Common use cases include:

- Accelerate application development
- Generate sample applications
- Monitor tenants and environment configuration

Have you got an interesting use case or project you'd like to share with the community? [We'd love to hear about it on the PingOne Community pages!](https://support.pingidentity.com/s/topic/0TO1W000000ddO4WAI/pingone)

## Getting Started

### Prerequisites

- **A licensed or trial PingOne cloud subscription.** - Don't have a tenant? [Sign up for a free trial here](https://www.pingidentity.com/en/try-ping.html).
- **MCP-compatible client** (E.g. Claude Desktop, VS Code Copilot Chat, Cursor, Zed, etc.)
- **Homebrew** (for macOS and Linux package install)

### Prepare PingOne for MCP Server Use

The MCP server requires a worker application in your PingOne tenant to access the management APIs. You'll need to capture two values during setup:

- **Environment ID** - The PingOne environment containing your worker application (referred to later as `{{admin environment id}}`)
- **Client ID** - The worker application's client identifier (referred to later as `{{mcp application client id}}`)

#### Default: Authorization Code with PKCE

The server uses the Authorization Code grant with PKCE by default. Configure your worker application with:

- **Grant Type**: Authorization Code with PKCE required
- **Response Type**: Code
- **Token Endpoint Authentication**: None
- **Redirect URI**: `http://127.0.0.1:7464/callback`
- **Application Roles**: None required (the MCP server inherits roles from the authenticated user)

> **Note:** For detailed instructions on creating the application and setting up admin users, see [Setting Up PingOne Worker Applications](docs/setup-pingone-worker-application.md).

<details>
<summary>Alternative: Device Authorization Grant (for headless/containerized environments)</summary>

#### Using the Device Authorization Grant

For headless or containerized environments, use the Device Authorization grant. Configure your worker application with:

- **Grant Type**: Device Authorization
- **Token Endpoint Authentication**: None
- **Redirect URI**: `http://127.0.0.1:7464/callback`
- **Application Roles**: None required (the MCP server inherits roles from the authenticated user)

This grant type is ideal for environments without a browser, such as CI/CD pipelines or remote servers.

</details>

### Install the MCP Server

#### macOS and Linux

Use Ping Identity's Homebrew tap to install the PingOne MCP server

```shell
brew tap pingidentity/tap
brew install pingone-mcp-server
```

Alternatively, expand the instructions below to install manually from GitHub release artifacts.

<details>
<summary>macOS - GitHub Release Manual Installation Instructions</summary>

##### macOS Manual Installation Instructions

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for artifact downloads, artifact signatures, and the checksum file. To verify package downloads, see the [Verify Section](#verify).

OR

Use the following single-line command to install the server into '/usr/local/bin' directly.

```shell
RELEASE_VERSION=$(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/pingidentity/pingone-mcp-server/releases/latest)); \
OS_NAME=$(uname -s); \
HARDWARE_PLATFORM=$(uname -m | sed s/aarch64/arm64/ | sed s/x86_64/amd64/); \
URL="https://github.com/pingidentity/pingone-mcp-server/releases/download/${RELEASE_VERSION}/pingone-mcp-server_${RELEASE_VERSION#v}_${OS_NAME}_${HARDWARE_PLATFORM}"; \
curl -Ls -o pingone-mcp-server "${URL}"; \
chmod +x pingone-mcp-server; \
sudo mv pingone-mcp-server /usr/local/bin/pingone-mcp-server;
```

##### Verify with Checksums

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for the checksums.txt file. The checksums are in the format of SHA256.

</details>

<details>
<summary>Linux - GitHub Release Manual Installation Instructions</summary>

##### Linux Manual Installation Instructions

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for artifact downloads, artifact signatures, and the checksum file. To verify package downloads, see the [Verify Section](#verify).

OR

Use the following single-line command to install the server into '/usr/local/bin' directly.

```shell
RELEASE_VERSION=$(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/pingidentity/pingone-mcp-server/releases/latest)); \
OS_NAME=$(uname -s); \
HARDWARE_PLATFORM=$(uname -m | sed s/aarch64/arm64/ | sed s/x86_64/amd64/); \
URL="https://github.com/pingidentity/pingone-mcp-server/releases/download/${RELEASE_VERSION}/pingone-mcp-server_${RELEASE_VERSION#v}_${OS_NAME}_${HARDWARE_PLATFORM}"; \
curl -Ls -o pingone-mcp-server "${URL}"; \
chmod +x pingone-mcp-server; \
sudo mv pingone-mcp-server /usr/local/bin/pingone-mcp-server;
```

##### Verify with Checksums

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for the checksums.txt file. The checksums are in the format of SHA256.

</details>

Test the installation:

```shell
pingone-mcp-server --version
```

### Use with VS Code

[![Install in VS Code](https://img.shields.io/badge/VS_Code-Install_Server-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=pingOne&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_environment_id%22%2C%22description%22%3A%22The%20environment%20ID%20containing%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_mcp_client_id%22%2C%22description%22%3A%22The%20client%20ID%20of%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_api_root_domain%22%2C%22description%22%3A%22The%20root%20domain%20of%20your%20PingOne%20tenant%20%28e.g.%2C%20%60pingone.com%60%20%2C%20%60pingone.eu%60%20%2C%20%60pingone.ca%60%29%22%2C%22password%22%3Afalse%7D%5D&config=%7B%22type%22%3A%22stdio%22%2C%22command%22%3A%22pingone-mcp-server%22%2C%22args%22%3A%5B%22run%22%5D%2C%22env%22%3A%7B%22PINGONE_MCP_ENVIRONMENT_ID%22%3A%22%24%7Binput%3Apingone_environment_id%7D%22%2C%22PINGONE_AUTHORIZATION_CODE_CLIENT_ID%22%3A%22%24%7Binput%3Apingone_mcp_client_id%7D%22%2C%22PINGONE_ROOT_DOMAIN%22%3A%22%24%7Binput%3Apingone_api_root_domain%7D%22%7D%7D) [![Install in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Install_Server-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=pingOne&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_environment_id%22%2C%22description%22%3A%22The%20environment%20ID%20containing%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_mcp_client_id%22%2C%22description%22%3A%22The%20client%20ID%20of%20the%20MCP%20server%20worker%20application%22%2C%22password%22%3Afalse%7D%2C%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22pingone_api_root_domain%22%2C%22description%22%3A%22The%20root%20domain%20of%20your%20PingOne%20tenant%20%28e.g.%2C%20%60pingone.com%60%20%2C%20%60pingone.eu%60%20%2C%20%60pingone.ca%60%29%22%2C%22password%22%3Afalse%7D%5D&config=%7B%22type%22%3A%22stdio%22%2C%22command%22%3A%22pingone-mcp-server%22%2C%22args%22%3A%5B%22run%22%5D%2C%22env%22%3A%7B%22PINGONE_MCP_ENVIRONMENT_ID%22%3A%22%24%7Binput%3Apingone_environment_id%7D%22%2C%22PINGONE_AUTHORIZATION_CODE_CLIENT_ID%22%3A%22%24%7Binput%3Apingone_mcp_client_id%7D%22%2C%22PINGONE_ROOT_DOMAIN%22%3A%22%24%7Binput%3Apingone_api_root_domain%7D%22%7D%7D&quality=insiders)

For quick installation, use one of the install buttons above.

To add the MCP server configuration manually, add the following configuration to your MCP configuration file:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run",
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "${input:pingone_environment_id}",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "${input:pingone_mcp_client_id}",
        "PINGONE_ROOT_DOMAIN": "${input:pingone_api_root_domain}",
      },
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

<details>
<summary>Alternative: Using the Device Authorization Grant</summary>

To configure the MCP server to use the Device Authorization grant type, add the `--grant-type` command argument with a value of `device_code` and add the environment variables `PINGONE_DEVICE_CODE_CLIENT_ID` and `PINGONE_DEVICE_CODE_SCOPES` as shown in the example:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run",
        "--grant-type",
        "device_code"
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "${input:pingone_environment_id}",
        "PINGONE_DEVICE_CODE_CLIENT_ID": "${input:pingone_mcp_client_id}",
        "PINGONE_DEVICE_CODE_SCOPES": "openid",
        "PINGONE_ROOT_DOMAIN": "${input:pingone_api_root_domain}",
      },
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

</details>

### Use with Claude Desktop

To add the MCP server configuration manually, add the following configuration to your Claude Desktop config (`claude_desktop_config.json`) or via `Settings` -> `Developer` -> `Local MCP Servers`:

```json
{
  "mcpServers": {
    "pingone": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run"
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "<<paste worker application environment UUID {{admin environment id}} here>>",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "<<paste worker application client ID UUID {{mcp application client id}} here>>",
        "PINGONE_ROOT_DOMAIN": "<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>"
      }
    }
  }
}
```

If you've downloaded the binary manually to a location not on the `PATH`, change the `command` to refer to the full path to the binary file.

<details>
<summary>Alternative: Using the Device Authorization Grant</summary>

To configure the MCP server to use the Device Authorization grant type, add the `--grant-type` command argument with a value of `device_code` and add the environment variables `PINGONE_DEVICE_CODE_CLIENT_ID` and `PINGONE_DEVICE_CODE_SCOPES` as shown in the example:

```json
{
  "mcpServers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run",
        "--grant-type",
        "device_code"
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "<<paste worker application environment UUID {{admin environment id}} here>>",
        "PINGONE_DEVICE_CODE_CLIENT_ID": "<<paste worker application client ID UUID {{mcp application client id}} here>>",
        "PINGONE_DEVICE_CODE_SCOPES": "openid",
        "PINGONE_ROOT_DOMAIN": "<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>"
      }
    }
  }
}
```

</details>

### Use with Claude Code

To install the MCP server in Claude code, run the following commands, changing the `PINGONE_MCP_ENVIRONMENT_ID`, `PINGONE_AUTHORIZATION_CODE_CLIENT_ID` and `PINGONE_ROOT_DOMAIN` environment variables for your tenant.

```shell
export PINGONE_MCP_ENVIRONMENT_ID="<<paste worker application environment UUID {{admin environment id}} here>>"
export PINGONE_AUTHORIZATION_CODE_CLIENT_ID="<<paste worker application client ID UUID {{mcp application client id}} here>>"
export PINGONE_ROOT_DOMAIN="<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>"
```

```shell
claude mcp add --transport stdio pingOne \
--env PINGONE_MCP_ENVIRONMENT_ID=$PINGONE_MCP_ENVIRONMENT_ID \
--env PINGONE_AUTHORIZATION_CODE_CLIENT_ID=$PINGONE_AUTHORIZATION_CODE_CLIENT_ID \
--env PINGONE_ROOT_DOMAIN=$PINGONE_ROOT_DOMAIN \
-- pingone-mcp-server run
```

Check the MCP server has been loaded correctly:

```shell
claude mcp list
```

```shell
Checking MCP server health...

pingOne: pingone-mcp-server run - ✓ Connected
```

### Use with Cursor

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/en-US/install-mcp?name=pingOne&config=eyJlbnYiOnsiUElOR09ORV9NQ1BfRU5WSVJPTk1FTlRfSUQiOiI8PHBhc3RlIHdvcmtlciBhcHBsaWNhdGlvbiBlbnZpcm9ubWVudCBVVUlEIGhlcmU%2BPiIsIlBJTkdPTkVfQVVUSE9SSVpBVElPTl9DT0RFX0NMSUVOVF9JRCI6Ijw8cGFzdGUgd29ya2VyIGFwcGxpY2F0aW9uIGNsaWVudCBJRCBVVUlEIGhlcmU%2BPiIsIlBJTkdPTkVfUk9PVF9ET01BSU4iOiI8PHBhc3RlIHJvb3QgZG9tYWluIG9mIHlvdXIgUGluZ09uZSB0ZW5hbnQgaGVyZSAoZS5nLiwgcGluZ29uZS5jb20pPj4ifSwiY29tbWFuZCI6InBpbmdvbmUtbWNwLXNlcnZlciBydW4ifQ%3D%3D)

For quick installation, the install button above.

Be sure to change the `PINGONE_MCP_ENVIRONMENT_ID`, `PINGONE_AUTHORIZATION_CODE_CLIENT_ID` and `PINGONE_ROOT_DOMAIN` environment variables for your tenant.

To add the MCP server configuration manually, add the following configuration to your Cursor config (`~/.cursor/mcp.json`) or via `Settings` -> `Cursor Settings` -> `Tools & MCP`:

```json
{
  "mcpServers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run"
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "<<paste worker application environment UUID {{admin environment id}} here>>",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "<<paste worker application client ID UUID {{mcp application client id}} here>>",
        "PINGONE_ROOT_DOMAIN": "<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>"
      }
    }
  }
}
```

<details>
<summary>Alternative: Using the Device Authorization Grant</summary>

To configure the MCP server to use the Device Authorization grant type, add the `--grant-type` command argument with a value of `device_code` and add the environment variables `PINGONE_DEVICE_CODE_CLIENT_ID` and `PINGONE_DEVICE_CODE_SCOPES` as shown in the example:

```json
{
  "mcpServers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": [
        "run",
        "--grant-type",
        "device_code"
      ],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "<<paste worker application environment UUID {{admin environment id}} here>>",
        "PINGONE_DEVICE_CODE_CLIENT_ID": "<<paste worker application client ID UUID {{mcp application client id}} here>>",
        "PINGONE_DEVICE_CODE_SCOPES": "openid",
        "PINGONE_ROOT_DOMAIN": "<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>"
      }
    }
  }
}
```

</details>

### Building from Source

If you'd like to build and run the project from source, use `make build` which will compile the code to a binary at `./bin/pingone-mcp-server`.

When configuring MCP clients, ensure that the `command` value refers to the full path up to the built binary.  For example:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "/path/to/cloned_projects/pingone-mcp-server/bin/pingone-mcp-server",
      "args": [
        "run",
      ],
      "env": {
        "PINGONE_MCP_DEBUG": "true",
        "PINGONE_MCP_ENVIRONMENT_ID": "<<paste worker application environment UUID {{admin environment id}} here>>",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "<<paste worker application client ID UUID {{mcp application client id}} here>>",
        "PINGONE_ROOT_DOMAIN": "<<paste root domain of your PingOne tenant here (e.g., pingone.com)>>",
      },
    },
  }
}
```

## Authentication and Authorization

The server uses **OAuth 2.0 Authorization Code flow with PKCE** for secure administrator authentication by default.  The server can be configured to use the **Device Authorization grant type (also using PKCE)** as an optional feature.

1. **First Tool Use** - Browser opens automatically for administrator login to your configured PingOne tenant when you use a tool for the first time in a session
2. **Token Storage** - Access tokens stored securely in OS keychain where available (macOS Keychain, Windows Credential Manager, Linux Secret Service)
3. **Automatic Reuse** - Cached tokens used for subsequent tool calls within the same session
4. **Auto Re-authentication** - When tokens expire during a session, browser opens again for new login

## Tool Configuration

> [!IMPORTANT]
> **Restrictions for Production Environments**
>
> By default, any tool that has the capability of writing both configuration and/or data, or any tool that can read production data are restricted for use on environments that are of type `PRODUCTION`. This is to safeguard against unintended access to sensitive data or accidental configuration changes to live systems.

> [!IMPORTANT]
> **Read Only by Default**
>
> By default the server starts in "Read Only" mode, to protect against accidental changes.  To enable write tools, add the `--disable-read-only` command line argument. For more information, see [Enabling Write Tools](#enabling-write-tools).

The MCP server provides a set of tools to interact with PingOne environments. Tools are organized into tool collections, that allow groups of tools to be enabled and disabled globally when the server starts.

Enabling/disabling tools (or collections of tools) provides the user control over which tools are made available to the MCP client at runtime, which can both reduce the number of unneeded tools for the agent (reducing tool and context bloat) but can also provide a backstop measure against accidental changes to unrelated configurations in the environment.

### Enabling Write Tools

By default, the server starts in **read-only mode**, which only exposes tools that retrieve information without modifying any configuration or data. This provides a safety layer against accidental changes.

To enable write operations (create, update, delete), add the `--disable-read-only` flag when starting the server:

```bash
pingone-mcp-server run --disable-read-only
```

Or in your MCP client configuration:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": ["run", "--disable-read-only"],
      "env": {
        "PINGONE_MCP_ENVIRONMENT_ID": "your-env-id",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "your-client-id",
        "PINGONE_ROOT_DOMAIN": "pingone.com"
      }
    }
  }
}
```

> [!WARNING]
> **Important Behavior**: If you explicitly include specific tools/collections that are capable of writing configuration to the PingOne tenant, you must also use `--disable-read-only` to make those write tools available to the MCP client.

### Specifying Tools and Tool Collections

You can fine-tune which tools are available using inclusion and exclusion flags. These flags accept comma-separated lists of tool names or collection names.

#### Available Flags

- `--include-tools` - Enable only specified tools
- `--exclude-tools` - Disable specified tools
- `--include-tool-collections` - Enable only specified collections
- `--exclude-tool-collections` - Disable specified collections
- `--disable-read-only` - Include write tools (required for create/update/delete operations)

#### Filtering Behavior

**Priority Rules:**

1. **Exclusions take priority** - If a tool appears in both include and exclude lists, it will be excluded
2. **Empty inclusion lists allow all** - If no `--include-*` flags are specified, all tools/collections are included by default (subject to read-only filter and exclusions)
3. **Read-only filter applies to tools** - The `--disable-read-only` flag must be set to include any write tools, even if explicitly included

> [!NOTE]
> **Conflicting Arguments**: If you specify write tools in `--include-tools` without adding `--disable-read-only`, the server will log a warning message listing which write tools will be excluded, along with a suggestion to add the flag.

**Examples:**

**Enable only read tools from specific collections:**

```bash
pingone-mcp-server run \
  --include-tool-collections applications,environments
```

This enables `list_applications`, `get_application`, `list_environments`, `get_environment`, and `get_environment_services` but excludes all write tools.

**Enable all tools including writes from specific collections:**

```bash
pingone-mcp-server run \
  --disable-read-only \
  --include-tool-collections applications,environments
```

This enables all application and environment tools including create, update operations.

**Enable everything except specific collections:**

```bash
pingone-mcp-server run \
  --disable-read-only \
  --exclude-tool-collections populations
```

This enables all tools across all collections except the populations collection.

**Enable specific tools only:**

```bash
pingone-mcp-server run \
  --include-tools list_applications,get_application,list_environments
```

This enables only the three specified read-only tools.

**Enable specific write tools (requires --disable-read-only):**

```bash
pingone-mcp-server run \
  --disable-read-only \
  --include-tools create_oidc_application,update_oidc_application
```

This enables only the two application write tools.

**Complex filtering - include collection but exclude specific tools:**

```bash
pingone-mcp-server run \
  --disable-read-only \
  --include-tool-collections applications \
  --exclude-tools update_oidc_application
```

This enables all application tools except `update_oidc_application`.

> [!CAUTION]
> **Known Limitation**: When using `--include-tools` with write tool names but forgetting `--disable-read-only`, the tools will be silently excluded. Always remember to add `--disable-read-only` when you intend to enable write operations.

> [!TIP]
> **Best Practice**: Start with read-only mode and specific collections, then gradually enable write tools as needed. This reduces cognitive load for AI agents and minimizes risk of unintended changes.

### Tool Collections

Tool collections group related tools together for easier management. Each collection corresponds to a PingOne resource type.

| Collection | Description | Tools Included |
|------------|-------------|----------------|
| `applications` | Manage OIDC/OAuth 2.0 applications in PingOne environments | `list_applications`, `get_application`, `create_oidc_application`, `update_oidc_application` |
| `environments` | Manage PingOne environments and their service configurations | `list_environments`, `get_environment`, `create_environment`, `update_environment`, `get_environment_services`, `update_environment_services` |
| `populations` | Manage user populations within PingOne environments | `list_populations`, `get_population`, `create_population`, `update_population` |

### Available Tools

The server provides tools for AI agents to interact with your PingOne environment:

#### Applications

Create, update, view applications within an environment.

| Tool | Collections | Read Only | Description | Usage Examples |
|------|-------------|-------------|-------------|----------------|
| `create_oidc_application` | `applications` | | Create an OpenID Connect/OAuth 2.0 application | - `Create an OIDC app called "My Web App"` <br> - `Create an application using PKCE with redirect URI https://myapp-dev.bxretail.org/callback` |
| `get_application` | `applications` | ✓ | Retrieve the detailed configuration of an application | - `Show me application abc-123` <br> - `Get the config for My Web App` <br> - `Display the OIDC settings for app xyz` |
| `list_applications` | `applications` | ✓ | List applications accessible to the authenticated user, each with a basic configuration summary, within an environment | - `Show all applications in environment xyz` <br> - `List OIDC apps` <br> - `What applications exist and are enabled?` |
| `update_oidc_application` | `applications` | | Update an OpenID Connect/OAuth 2.0 application's configuration | - `Update app xyz to add a new redirect URI` <br> - `Change the token lifetime for My Web App` <br> - `Modify the grant types for application abc-123` <br> - `Disable application abc-123` |

#### Environments

Manage PingOne environments and their services.

| Tool | Collections | Read Only | Description | Usage Examples |
|------|-------------|-------------|-------------|----------------|
| `create_environment` | `environments` |  | Create a new sandbox PingOne environment | - `Create an environment called Dev` <br> - `Add a new environment in the NA region` <br> - `Create a test environment for our team` |
| `get_environment` | `environments` | ✓ | Retrieve an environment's full configuration | - `Show me environment abc-123` <br> - `Get the config for Dev environment` <br> - `Display environment xyz details` |
| `list_environments` | `environments` | ✓ | List all PingOne environments accessible to the authenticated user | - `Show all environments` <br> - `List active environments` <br> - `Find environments starting with "Prod"` |
| `update_environment` | `environments` | | Update environment configuration | - `Rename environment to Testing` <br> - `Change description of Dev environment` |
| `get_environment_services` | `environments` | ✓ | Retrieve all PingOne shared services assigned to a specified environment | - `What services are enabled in environment xyz?` <br> - `Show me the bill of materials` <br> - `List services for Dev environment` <br> - `Are MFA and Neo enabled on environment abc-123` |
| `update_environment_services` | `environments` | | Update the services assigned to an environment | - `Enable DaVinci in environment xyz` <br> - `Add PingOne Verify service` <br> - `Update the environment abc-123 services to include MFA` |

#### Populations

Manage user populations within environments.

| Tool | Collections | Read Only | Description | Usage Examples |
|------|-------------|-------------|-------------|----------------|
| `create_population` | `populations` | | Create a population in an environment | - `Create a population called External Users` <br> - `Add population for employees` <br> - `Create Customers population with French language` |
| `get_population` | `populations` | ✓ | Retrieve population configuration by ID | - `Show me population abc-123` <br> - `Get the External Users population config` <br> - `Display population xyz details` |
| `list_populations` | `populations` | ✓ | List populations in an environment | - `Show all populations in environment xyz` <br> - `List populations` <br> - `Find populations starting with "External"` |
| `update_population` | `populations` | | Update population configuration | - `Change population description` <br> - `Update External Users to use new password policy` <br> - `Modify preferred language for population xyz` |

## Security

The PingOne MCP Server implements multiple security layers:

- **Secure credential storage** - Tokens stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service) for local deployment, or ephemerally in container filesystem for Docker
- **No plain text secrets** - No sensitive information stored in configuration files
- **OAuth 2.0 authentication** - PKCE flow for local deployment prevents authorization code interception; Device Code flow for containerized deployment
- **User-based authentication** - All API calls are authenticated as the user who logged in, providing complete audit trails

## Troubleshooting

Having issues with the MCP server? Check the comprehensive [Troubleshooting Guide](docs/troubleshooting.md) for solutions to common problems including:

- Authentication and permission errors
- Configuration issues
- Tool execution problems
- MCP client integration issues
- Debug mode and logging

For quick debugging, enable debug mode by setting `PINGONE_MCP_DEBUG=true` in your environment variables. See the [troubleshooting guide](docs/troubleshooting.md#debug-mode) for details.

## Contributing

### Feedback and Issues

We welcome your feedback!  Please use this repository's issue tracker to submit feedback, bug reports, or enhancement requests. For existing issues, you can add a 👍 reaction to help our team gauge priority.

### Pull Request Guidelines

We welcome pull requests for:

  * Repository management (e.g., scripts, GitHub Actions)
  * Documentation updates
  * Code contributions to advance the project such as adding new tools or collections. _For larger or more structural changes, please raise an issue on the project as a proposal first so the project team can provide guidance._

Please see our [contributing guide](CONTRIBUTING.md) for more information.
