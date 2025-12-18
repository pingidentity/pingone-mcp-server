# Troubleshooting Guide

This guide provides solutions to common issues you may encounter when setting up and using the PingOne MCP Server.

## Table of Contents

- [Debug Mode](#debug-mode)
- [Authentication Issues](#authentication-issues)
- [Configuration Issues](#configuration-issues)
- [Runtime and Tool Execution Issues](#runtime-and-tool-execution-issues)
- [MCP Client Integration Issues](#mcp-client-integration-issues)
- [Getting Additional Help](#getting-additional-help)

## Debug Mode

For troubleshooting issues with the MCP server, you can enable debug mode to collect detailed logging information.

### Enabling Debug Mode

Set the `PINGONE_MCP_DEBUG` environment variable to `true` in your MCP client configuration:

```json
{
  "servers": {
    "pingOne": {
      "type": "stdio",
      "command": "pingone-mcp-server",
      "args": ["run"],
      "env": {
        "PINGONE_MCP_DEBUG": "true",
        "PINGONE_MCP_ENVIRONMENT_ID": "your-env-id",
        "PINGONE_AUTHORIZATION_CODE_CLIENT_ID": "your-client-id",
        "PINGONE_ROOT_DOMAIN": "pingone.com"
      }
    }
  }
}
```

### What Debug Mode Logs

When debug mode is enabled, the server will output detailed logs including:

- HTTP request and response details
- Authentication flow information
- Tool invocation details
- Error stack traces
- Token acquisition and refresh operations
- API endpoint calls and responses

> **Security Warning:** Debug logs may contain sensitive information including API responses, token metadata, and configuration details. Only enable debug mode in development environments and be cautious when sharing logs. Never share logs publicly without redacting sensitive information.

### Viewing Debug Logs

Debug logs are output to stderr and will be visible in:

- **VS Code**: Check the "Output" panel, select "GitHub Copilot Chat" or "MCP" from the dropdown
- **Cursor**: Check the MCP logs in the settings or debug console
- **Claude Desktop**: Check the application logs (location varies by OS)
- **Terminal/CLI**: Logs appear in the terminal where the MCP client is running

## Authentication Issues

### Issue: "Invalid redirect URI" error

**Symptoms:**
- Authentication fails with an error about invalid redirect URI
- Browser shows "redirect_uri_mismatch" error

**Solution:**
1. Verify the redirect URI in your PingOne worker application exactly matches `http://127.0.0.1:7464/callback`
2. Ensure there are no trailing slashes or extra characters
3. Check that the port number is 7464 (not 8080 or another port)
4. Verify the protocol is `http` (not `https`) for local development

### Issue: "Invalid client" error

**Symptoms:**
- Authentication fails with "invalid_client" error
- Server logs show client authentication failed

**Solution:**
1. Verify the Client ID in your MCP configuration matches the worker application's Client ID in PingOne
2. Ensure the worker application is **enabled** in PingOne (toggle switch in the admin console)
3. Check that you're using the correct environment ID
4. Verify the application type is set to "Worker" in PingOne

### Issue: "Insufficient permissions" error

**Symptoms:**
- Authentication succeeds but API calls fail with 403 Forbidden errors
- Error messages indicate insufficient permissions or unauthorized access

**Solution:**
1. Verify your PingOne user account has appropriate administrative roles assigned
2. Check that roles are assigned at the environment level where you're making API calls
3. Confirm the worker application has no roles assigned (it should inherit roles from the user)
4. Review the [Setting Up Administrative Users](setup-pingone-worker-application.md#setting-up-administrative-users) guide
5. Common required roles include:
   - Environment Admin (full access)
   - Identity Data Admin (user/group management)
   - Client Application Developer (application management)
   - Configuration Read Only (read-only access)

### Issue: "PKCE verification failed" error

**Symptoms:**
- Authentication flow fails at the token exchange step
- Error mentions PKCE code challenge or verifier mismatch

**Solution:**
1. Ensure PKCE Enforcement is set to `S256_REQUIRED` in the worker application configuration
2. Verify the application's OAuth 2.0 settings include the Authorization Code grant type
3. Clear any cached tokens and try authenticating again
4. If using a custom HTTP client or proxy, ensure PKCE parameters are not being modified

### Issue: Device code not working

**Symptoms:**
- Device authorization flow fails to initiate
- Error about unsupported grant type or missing device code endpoint

**Solution:**
1. Verify the worker application has the **Device Authorization** grant type enabled in PingOne
2. Check that you're using the `--grant-type device_code` argument when starting the server
3. Ensure the `PINGONE_DEVICE_CODE_CLIENT_ID` environment variable is set correctly
4. Verify you're using the correct environment ID

### Issue: "Client is missing required grant type: DEVICE_CODE"

**Symptoms:**
- Error specifically mentions missing DEVICE_CODE grant type
- Device authorization flow cannot be initiated

**Solution:**
1. Open your worker application in the PingOne admin console
2. Navigate to the Configuration tab
3. Edit the OAuth 2.0 settings
4. Enable the **Device Authorization** grant type
5. Save the changes and ensure the application is enabled

### Issue: Authentication Blocked - "Authentication has failed for unknown reason"

**Symptoms:**
- Device authorization flow displays user code and verification URL
- After entering the code in the browser, authentication fails with a generic error
- Error message: "Authentication has failed for unknown reason"

**Solution:**
1. Ensure the `PINGONE_DEVICE_CODE_SCOPES` environment variable is defined in your MCP configuration
2. Set the value to `openid` at minimum:
   ```json
   "env": {
     "PINGONE_DEVICE_CODE_SCOPES": "openid",
     "PINGONE_DEVICE_CODE_CLIENT_ID": "your-client-id",
     ...
   }
   ```
3. Restart the MCP server after adding the environment variable
4. Try the device authorization flow again

### Issue: Browser doesn't open automatically (Authorization Code)

**Symptoms:**
- No browser window opens when using a tool for the first time
- Server appears to hang waiting for authentication

**Solution:**
1. Check if a browser is already open with the authorization URL but in the background
2. Look for a URL in the server logs that starts with your PingOne domain
3. Manually copy and paste the URL into a browser
4. Check if your system's default browser is properly configured
5. Try setting a different default browser

### Issue: Session expired or tokens not refreshing

**Symptoms:**
- Authentication works initially but fails after some time
- Error about expired tokens or invalid refresh tokens

**Solution:**
1. This is expected behavior - re-authenticate when prompted
2. The server will automatically open a browser for re-authentication
3. Check if refresh tokens are enabled in the worker application configuration
4. Verify token storage is working correctly (keychain access on macOS/Windows)
5. Clear stored tokens and re-authenticate:
   - Run `pingone-mcp-server session --logout` (if available)
   - Or manually clear tokens from your OS keychain/credential manager

## Configuration Issues

### Issue: Environment variables not recognized

**Symptoms:**
- Server starts but doesn't use configured values
- Error about missing required configuration

**Solution:**
1. Verify environment variables are set in the correct location:
   - For VS Code: In the MCP configuration's `env` object
   - For Cursor: In `~/.cursor/mcp.json`
   - For Claude Desktop: In the Claude configuration file
2. Check for typos in environment variable names
3. Ensure values don't have extra quotes or spaces
4. Restart the MCP client application after configuration changes

### Issue: Wrong PingOne region

**Symptoms:**
- Authentication fails with DNS or connection errors
- 404 errors when calling PingOne APIs

**Solution:**
1. Verify the `PINGONE_ROOT_DOMAIN` environment variable matches your tenant's region:
   - North America: `pingone.com`
   - Europe: `pingone.eu`
   - Asia Pacific: `pingone.asia`
   - Canada: `pingone.ca`
2. Check your PingOne admin console URL to confirm the correct region
3. Update the configuration and restart the MCP server

### Issue: "Command not found" error

**Symptoms:**
- MCP client reports that `pingone-mcp-server` command cannot be found
- Server fails to start with command not found error

**Solution:**
1. Verify the server is installed correctly:
   ```bash
   which pingone-mcp-server
   ```
2. For Homebrew installation:
   ```bash
   brew list pingone-mcp-server
   ```
3. Ensure `/usr/local/bin` is in your PATH
4. Try specifying the full path to the binary in your MCP configuration:
   ```json
   "command": "/usr/local/bin/pingone-mcp-server"
   ```
5. If building from source, use the full path to the built binary

### Issue: Configuration file format errors

**Symptoms:**
- MCP client fails to start with JSON parsing errors
- Error about invalid configuration format

**Solution:**
1. Validate your JSON configuration using a JSON validator
2. Common issues:
   - Trailing commas in JSON (not allowed in strict JSON)
   - Missing quotes around strings
   - Incorrect nesting of objects
3. Use a JSON formatter to identify syntax errors
4. Compare your configuration against the examples in the documentation

## Runtime and Tool Execution Issues

### Issue: Tools not appearing in MCP client

**Symptoms:**
- MCP server starts successfully but no tools are available
- MCP client shows empty tools list

**Solution:**
1. Enable debug mode to see if tools are being registered
2. Check if read-only mode is preventing write tools from loading:
   - Add `--disable-read-only` if you need write tools
3. Verify tool filtering flags:
   - Check `--include-tools` or `--exclude-tools` arguments
   - Check `--include-tool-collections` or `--exclude-tool-collections` arguments
4. Ensure the server is fully started and connected to the MCP client
5. Try restarting the MCP client application

### Issue: Write tools not available despite specifying them

**Symptoms:**
- Specified write tools with `--include-tools` but they don't appear
- Only read-only tools are available

**Solution:**
1. Add the `--disable-read-only` flag to enable write tools:
   ```json
   "args": ["run", "--disable-read-only"]
   ```
2. The server will log a warning if you specify write tools without this flag
3. Check debug logs for the warning message about filtered write tools

### Issue: "Production environment write protection" error

**Symptoms:**
- Write operations fail on production environments
- Error message about production environment protection

**Solution:**
1. This is expected behavior - production environments are protected by default
2. The server prevents write operations on environments marked as `PRODUCTION` type
3. To disable this protection (not recommended for actual production):
   - Use a sandbox or development environment instead
   - Contact the project maintainers if you have a legitimate need to modify this behavior
4. For testing, create a non-production environment in PingOne

### Issue: Tool execution fails with timeout

**Symptoms:**
- Tool calls hang and eventually timeout
- No error message or response from the tool

**Solution:**
1. Check your network connectivity to PingOne APIs
2. Verify firewall or proxy settings aren't blocking API calls
3. Enable debug mode to see where the request is failing
4. Check PingOne service status for outages
5. Increase timeout values if working with slow networks

### Issue: Unexpected tool responses or errors

**Symptoms:**
- Tool returns unexpected data or format
- Error messages that don't match the documentation

**Solution:**
1. Enable debug mode to see the full API request and response
2. Verify you're passing parameters in the correct format
3. Check if the PingOne API version has changed
4. Review the [tool documentation](../README.md#available-tools) for correct usage
5. Ensure your worker application has necessary permissions

## MCP Client Integration Issues

### Issue: VS Code GitHub Copilot not detecting server

**Symptoms:**
- MCP server doesn't appear in VS Code
- Copilot Chat doesn't show PingOne tools

**Solution:**
1. Ensure you have the latest version of GitHub Copilot extension
2. Check that MCP support is enabled in VS Code settings
3. Verify the MCP configuration file is in the correct location
4. Restart VS Code completely after configuration changes
5. Check VS Code Output panel for MCP-related errors

### Issue: Cursor not loading MCP server

**Symptoms:**
- MCP server not visible in Cursor settings
- Configuration appears correct but server doesn't start

**Solution:**
1. Verify `~/.cursor/mcp.json` has correct syntax
2. Ensure Cursor has been restarted after configuration changes
3. Check Cursor logs for MCP initialization errors
4. Try the "Reload MCP Servers" command in Cursor
5. Verify you're using a recent version of Cursor with MCP support

### Issue: Claude Desktop connection problems

**Symptoms:**
- Claude Desktop doesn't connect to the MCP server
- Server process starts but Claude doesn't recognize it

**Solution:**
1. Verify the Claude Desktop MCP configuration file location (varies by OS):
   - macOS: `~/Library/Application Support/Claude/`
   - Windows: `%APPDATA%\Claude\`
   - Linux: `~/.config/Claude/`
2. Ensure the configuration format matches Claude Desktop's requirements
3. Restart Claude Desktop after configuration changes
4. Check Claude Desktop logs for connection errors

### Issue: Multiple MCP servers conflicting

**Symptoms:**
- Unexpected behavior when multiple MCP servers are configured
- Tools from one server affecting another

**Solution:**
1. Ensure each MCP server has a unique name in the configuration
2. Check for port conflicts if servers use network connections
3. Review the MCP client's server management settings
4. Disable unused MCP servers temporarily to isolate issues

## Getting Additional Help

If you continue to experience issues after trying the solutions above:

### Before Seeking Help

1. **Enable debug mode** and reproduce the issue to capture detailed logs
2. **Redact sensitive information** from logs (tokens, client secrets, personal data)
3. **Document the issue** including:
   - What you were trying to do
   - What you expected to happen
   - What actually happened
   - Steps to reproduce the issue
   - Your environment (OS, MCP client, versions)

### Support Channels

1. **Documentation**
   - Review the [main README](../README.md) for general usage information
   - Check the [setup guide](setup-pingone-worker-application.md) for configuration details

2. **GitHub Issues**
   - Search [existing issues](https://github.com/pingidentity/pingone-mcp-server/issues) for similar problems
   - Create a new issue with detailed information and debug logs
   - Add a 👍 reaction to existing issues to help prioritize them

3. **PingOne Community**
   - Visit the [PingOne Community forums](https://support.pingidentity.com/s/topic/0TO1W000000ddO4WAI/pingone)
   - Ask questions and share experiences with other users

4. **Ping Identity Support**
   - Where PingOne service issues are encountered, for licensed customers: Contact [Ping Identity Support](https://support.pingidentity.com/)
   - Note: MCP Server is preview software with limited support during the public preview phase.

### What to Include in Bug Reports

When reporting issues, include:

1. **Environment Information**
   - Operating System and version
   - MCP client (VS Code/Cursor/Claude Desktop) and version
   - PingOne MCP Server version
   - PingOne region

2. **Configuration** (with secrets redacted)
   - MCP server configuration
   - Command-line arguments used
   - Environment variables set

3. **Debug Logs**
   - Enable debug mode and capture relevant log output
   - Redact sensitive information
   - Include timestamps if available

4. **Steps to Reproduce**
   - Clear, numbered steps to recreate the issue
   - Expected vs. actual behavior
   - Screenshots or error messages if applicable

5. **Workarounds Attempted**
   - List what you've already tried
   - Note any temporary solutions that partially work

## Additional Resources

- [PingOne MCP Server README](../README.md)
- [Setting Up PingOne Worker Applications](setup-pingone-worker-application.md)
- [PingOne Documentation](https://docs.pingidentity.com/r/en-us/pingone/p1_c_pingone_overview)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [GitHub Repository](https://github.com/pingidentity/pingone-mcp-server)
