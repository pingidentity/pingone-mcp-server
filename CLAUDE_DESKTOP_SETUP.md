# Claude Desktop MCP Setup Guide

This guide will help you configure Claude Desktop to use your PingOne MCP server.

## Prerequisites

1. **Claude Desktop installed** and running
2. **PingOne MCP server built**: `go build -o pingone-mcp-server ./cmd/server`
3. **PingOne credentials** (Client ID, Client Secret, Environment ID)

## Step 1: Build the Server

```bash
cd /path/to/pingone-mcp-server
go build -o pingone-mcp-server ./cmd/server
chmod +x pingone-mcp-server
```

## Step 2: Test the Server (Optional)

Before configuring Claude Desktop, test that the server works:

```bash
# Test without credentials (should show configuration_status tool)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./pingone-mcp-server

# Test with credentials
export PINGONE_CLIENT_ID="your-client-id"
export PINGONE_CLIENT_SECRET="your-client-secret" 
export PINGONE_ENV_ID="your-environment-id"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./pingone-mcp-server
```

## Step 3: Configure Claude Desktop

### Location of Configuration File

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

### Configuration Content

Create or edit the `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "pingone": {
      "command": "/full/path/to/your/pingone-mcp-server",
      "args": [],
      "env": {
        "PINGONE_CLIENT_ID": "your-client-id-here",
        "PINGONE_CLIENT_SECRET": "your-client-secret-here",
        "PINGONE_ENV_ID": "your-environment-id-here",
        "PINGONE_REGION": "com",
        "PINGONE_MCP_ALLOW_MUTATION": "true"
      }
    }
  }
}
```

### Configuration Options

| Environment Variable | Required | Description | Default |
|---------------------|----------|-------------|---------|
| `PINGONE_CLIENT_ID` | ✅ | Your PingOne OAuth client ID | - |
| `PINGONE_CLIENT_SECRET` | ✅ | Your PingOne OAuth client secret | - |
| `PINGONE_ENV_ID` | ✅ | Your PingOne environment ID | - |
| `PINGONE_REGION` | ❌ | PingOne region (com, eu, ca, asia) | `com` |
| `PINGONE_MCP_TRANSPORT` | ❌ | Transport mode: 'stdio' for Claude Desktop, 'http' for REST API | `stdio` |
| `PINGONE_MCP_ALLOW_MUTATION` | ❌ | Enable write operations | `false` |
| `PINGONE_MCP_DEBUG_API` | ❌ | Log API requests and responses to PingOne | `false` |
| `PINGONE_MCP_ALLOW_INSECURE` | ❌ | Disable API key requirement for HTTP mode | `false` |
| `PINGONE_MCP_SERVER_PORT` | ❌ | HTTP server port (HTTP mode only) | `8080` |
| `PINGONE_MCP_API_KEY_PATH` | ❌ | Path to API key file (HTTP mode only) | `pingone-mcp-server-api.key` |

## Step 4: Restart Claude Desktop

After saving the configuration file, **completely quit and restart Claude Desktop**.

## Step 5: Test Integration

1. Open a new conversation in Claude Desktop
2. Type a message like: "What PingOne tools are available?"
3. Claude should respond with a list of available PingOne tools

## Troubleshooting

### Claude Desktop Not Showing Tools

1. **Check configuration file syntax**: 
   ```bash
   cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | python -m json.tool
   ```

2. **Check file path**: Ensure the `command` path points to your built binary
   ```bash
   ls -la /full/path/to/your/pingone-mcp-server
   ```

3. **Test server manually**:
   ```bash
   ./pingone-mcp-server
   # Then type: {"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
   ```

### Common Issues

**"Configuration Status" Tool Only**: This means PingOne credentials are missing or invalid. Check your environment variables.

**Server Not Starting**: Check that the binary path is correct and executable.

**No Response from Tools**: Check PingOne credentials and network connectivity.

### Debug Mode

Enable debug logging by adding to the `env` section:

```json
"env": {
  "PINGONE_CLIENT_ID": "...",
  "PINGONE_MCP_DEBUG_API": "true"
}
```

## Example Working Configuration

Here's a complete working example (replace with your actual values):

```json
{
  "mcpServers": {
    "pingone": {
      "command": "/Users/yourname/workspace/pingone-mcp-server/pingone-mcp-server",
      "args": [],
      "env": {
        "PINGONE_CLIENT_ID": "abcd1234-5678-90ef-ghij-klmnopqrstuv",
        "PINGONE_CLIENT_SECRET": "your-secret-here",
        "PINGONE_ENV_ID": "12345678-abcd-efgh-ijkl-mnopqrstuvwx",
        "PINGONE_REGION": "com",
        "PINGONE_MCP_ALLOW_MUTATION": "true"
      }
    }
  }
}
```

## Available Tools

Once configured, you'll have access to these PingOne tools:

### Read-Only Tools (Always Available)
- `get_user` - Get user details by ID
- `get_user_password_state` - Check user password status
- `get_population` - Get population details
- `get_group` - Get group details  
- `get_environment` - Get environment details
- `get_environment_bom` - Get environment bill of materials
- `get_license` - Get license details

### Write Tools (When `PINGONE_MCP_ALLOW_MUTATION=true`)
- `create_user` - Create new users
- `update_user` - Update user attributes
- `delete_user` - Delete users
- `set_user_enabled` - Enable/disable users
- `reset_user_password` - Reset user passwords
- `unlock_user_password` - Unlock user accounts
- `add_user_to_group` - Add users to groups
- `remove_user_from_group` - Remove users from groups
- `create_population` - Create populations
- `delete_population` - Delete populations
- `create_group` - Create groups
- `update_group` - Update groups
- `delete_group` - Delete groups
- `create_environment` - Create environments
- `delete_environment` - Delete environments
- `update_environment_status` - Update environment status

## Security Notes

- Keep your PingOne credentials secure
- Use read-only mode (`PINGONE_MCP_ALLOW_MUTATION=false`) unless you need write operations
- Consider using environment-specific credentials for development vs production