# Setting Up PingOne Worker Applications for MCP Server

This guide provides detailed instructions for creating and configuring worker applications in PingOne for use with the PingOne MCP Server.

## Overview

The MCP server requires a worker application in your PingOne tenant to authenticate and access the management APIs. The worker application acts as the OAuth 2.0 client that facilitates secure authentication flows.

**Important:** You'll need to capture two values during setup:

- **Environment ID** - The PingOne environment containing your worker application
- **Client ID** - The worker application's client identifier

These values will be used when configuring your MCP client to connect to the server.

## Prerequisites

- Access to a PingOne tenant (trial or licensed)
- Administrative privileges in the target PingOne environment
- For automated setup: [Ping CLI](https://github.com/pingidentity/pingcli) installed and configured

## Authorization Code with PKCE (Recommended)

The Authorization Code grant with PKCE (Proof Key for Code Exchange) is the recommended authentication flow for local development and environments with browser access. This flow provides enhanced security by preventing authorization code interception attacks.

### Option 1: Manual Setup via PingOne Console

1. **Log in to PingOne Administration Console**
   - Navigate to your PingOne tenant
   - Open the environment where your administrative users are located

2. **Create the Worker Application**
   - Go to **Applications** → **Applications**
   - Click the **(+)** icon to add a new application
   - Provide a descriptive name (e.g., "PingOne MCP Server")
   - Add a meaningful description (e.g., "Worker application for PingOne MCP Server authentication")
   - Select **Worker** as the application type
   - Click **Save**

3. **Configure Application Roles**
   - Navigate to the **Roles** tab
   - **Do not assign any administrative roles** to the application
   - The MCP server will inherit administrator roles from the authenticated user's account, not from the worker application

4. **Configure OAuth 2.0 Settings**
   - Navigate to the **Configuration** tab
   - Click the **pencil icon** to edit the configuration
   - Configure the following settings:

   | Setting | Value |
   |---------|-------|
   | **Response Type** | Code |
   | **Grant Type** | Authorization Code |
   | **PKCE Enforcement** | S256_REQUIRED |
   | **Refresh Token** | Enabled (checked) |
   | **Redirect URIs** | `http://127.0.0.1:7464/callback` |
   | **Token Endpoint Authentication Method** | None |

   - Leave all other settings at their default values
   - Click **Save**

5. **Enable the Application**
   - Toggle the **Enabled** switch at the top of the application configuration to activate it
   - The application is now ready for use

6. **Capture Configuration Values**
   - Note the **Environment ID** from the environment details
   - Note the **Client ID** from the application's Overview tab
   - You'll need these values when configuring your MCP client

### Option 2: Automated Setup with Ping CLI

For automated or scripted deployments, you can use Ping CLI to create the worker application programmatically.

**Prerequisites:**
- Ping CLI installed and configured with a valid profile
- Administrative access to the target environment

**Steps:**

1. **Set the Target Environment ID**
   
   Replace `{{my_admin_environment_uuid}}` with your environment ID:

   ```bash
   TARGET_ENVIRONMENT_ID={{my_admin_environment_uuid}}
   ```

2. **Create the Worker Application**

   Execute the following command to create the worker application:

   ```bash
   pingcli request \
     --service pingone \
     --http-method POST \
     --data-raw '{
       "name": "PingOne MCP Server",
       "description": "Worker application for PingOne MCP Server authentication",
       "enabled": true,
       "type": "WORKER",
       "accessControl": {
         "role": {
           "type": "ADMIN_USERS_ONLY"
         }
       },
       "protocol": "OPENID_CONNECT",
       "responseTypes": ["CODE"],
       "pkceEnforcement": "S256_REQUIRED",
       "redirectUris": ["http://127.0.0.1:7464/callback"],
       "grantTypes": ["REFRESH_TOKEN", "AUTHORIZATION_CODE"],
       "tokenEndpointAuthMethod": "NONE"
     }' \
     environments/$TARGET_ENVIRONMENT_ID/applications
   ```

3. **Capture the Response**
   
   The command will return a JSON response containing the application details. Save the `id` field (Client ID) for later use.

## Device Authorization Grant (Alternative)

The Device Authorization grant is ideal for headless environments, containerized deployments, or scenarios where browser-based authentication is not feasible (e.g., CI/CD pipelines, remote servers).

### Option 1: Manual Setup via PingOne Console

1. **Log in to PingOne Administration Console**
   - Navigate to your PingOne tenant
   - Open the environment where your administrative users are located

2. **Create the Worker Application**
   - Go to **Applications** → **Applications**
   - Click the **(+)** icon to add a new application
   - Provide a descriptive name (e.g., "PingOne MCP Server - Device Flow")
   - Add a meaningful description (e.g., "Worker application for headless PingOne MCP Server authentication")
   - Select **Worker** as the application type
   - Click **Save**

3. **Configure Application Roles**
   - Navigate to the **Roles** tab
   - **Do not assign any administrative roles** to the application
   - The MCP server will inherit administrator roles from the authenticated user's account

4. **Configure OAuth 2.0 Settings**
   - Navigate to the **Configuration** tab
   - Click the **pencil icon** to edit the configuration
   - Configure the following settings:

   | Setting | Value |
   |---------|-------|
   | **Grant Type** | Device Authorization |
   | **Token Endpoint Authentication Method** | None |

   - Leave all other settings at their default values
   - Click **Save**

5. **Enable the Application**
   - Toggle the **Enabled** switch at the top of the application configuration to activate it
   - The application is now ready for use

6. **Capture Configuration Values**
   - Note the **Environment ID** from the environment details
   - Note the **Client ID** from the application's Overview tab
   - You'll need these values when configuring your MCP client

### Option 2: Automated Setup with Ping CLI

For automated or scripted deployments using the Device Authorization grant:

1. **Set the Target Environment ID**
   
   Replace `{{my_admin_environment_uuid}}` with your environment ID:

   ```bash
   TARGET_ENVIRONMENT_ID={{my_admin_environment_uuid}}
   ```

2. **Create the Worker Application**

   Execute the following command to create the worker application:

   ```bash
   pingcli request \
     --service pingone \
     --http-method POST \
     --data-raw '{
       "name": "PingOne MCP Server - Device Flow",
       "description": "Worker application for headless PingOne MCP Server authentication",
       "enabled": true,
       "type": "WORKER",
       "accessControl": {
         "role": {
           "type": "ADMIN_USERS_ONLY"
         }
       },
       "protocol": "OPENID_CONNECT",
       "grantTypes": ["DEVICE_CODE"],
       "tokenEndpointAuthMethod": "NONE"
     }' \
     environments/$TARGET_ENVIRONMENT_ID/applications
   ```

3. **Capture the Response**
   
   The command will return a JSON response containing the application details. Save the `id` field (Client ID) for later use.

## Setting Up Administrative Users

For the MCP server to function properly, the PingOne environment must have at least one administrative user account that can authenticate through the worker application.

### Prerequisites

- An existing PingOne user account
- Appropriate administrative roles assigned to the user

### Assigning Administrative Roles

1. **Navigate to User Management**
   - In the PingOne console, go to **Directory** → **Users**
   - Locate the user account that will authenticate with the MCP server

2. **Assign Roles**
   - Click on the user to view their details
   - Navigate to the **Roles** tab
   - Click **Add Roles**
   - Select the appropriate administrative roles based on your requirements:
     - **Environment Admin** - Full administrative access to the environment
     - **Identity Data Admin** - Manage users, groups, and populations
     - **Client Application Developer** - Manage applications and connections
     - **Configuration Read Only** - Read-only access to configuration

   > **Important:** The MCP server will inherit these user roles when making API calls. Follow the principle of least privilege and only assign the minimum roles required for your use case.  [Custom roles](https://docs.pingidentity.com/pingone/directory/p1_custom_role_add.html) may be created to assign only a subset of permissions to the user instead of using predefined roles.

3. **Save the Role Assignments**
   - Click **Save** to apply the role assignments

### Verifying User Access

After assigning roles:

1. Test that the user can log in to the PingOne administration console
2. Verify that the user has the expected permissions
3. Confirm that the user belongs to a population accessible by the worker application

## Verifying the Configuration

After setting up your worker application, verify the configuration:

### For Authorization Code Grant

1. **Check Application Settings**
   - Confirm the application is **Enabled**
   - Verify **Grant Type** includes Authorization Code
   - Verify **PKCE Enforcement** is set to S256_REQUIRED
   - Verify **Redirect URI** is `http://127.0.0.1:7464/callback`
   - Verify **Token Endpoint Authentication Method** is None

2. **Test Authentication**
   - Configure your MCP client with the captured Environment ID and Client ID
   - Start the MCP server and trigger a tool call
   - You should be redirected to a PingOne login page in your browser
   - After successful authentication, the server should complete the authorization flow

### For Device Authorization Grant

1. **Check Application Settings**
   - Confirm the application is **Enabled**
   - Verify **Grant Type** includes Device Authorization
   - Verify **Token Endpoint Authentication Method** is None

2. **Test Authentication**
   - Configure your MCP client with the captured Environment ID and Client ID
   - Configure the server to use device authorization grant (`--grant-type device_code`)
   - Start the MCP server and trigger a tool call
   - You should receive a device code and verification URL
   - Navigate to the verification URL in a browser and enter the device code
   - After successful authentication, the server should complete the authorization flow

## Troubleshooting

For detailed troubleshooting information including authentication issues, configuration problems, and runtime errors, see the comprehensive [Troubleshooting Guide](troubleshooting.md).

### Common Setup Issues

**Issue: "Invalid redirect URI" error**
- **Solution:** Verify that the redirect URI in your worker application exactly matches `http://127.0.0.1:7464/callback` (for Authorization Code grant)

**Issue: "Invalid client" error**
- **Solution:** Ensure the Client ID is correct and the application is enabled in PingOne

**Issue: "Insufficient permissions" error**
- **Solution:** Verify that your user account has the appropriate administrative roles assigned

For more issues and detailed solutions, see the [Troubleshooting Guide](troubleshooting.md#authentication-issues).

## Security Best Practices

1. **Principle of Least Privilege**
   - Only assign the minimum required administrative roles to user accounts
   - Regularly review and audit role assignments

2. **Application Access Control**
   - Use the `ADMIN_USERS_ONLY` access control setting to restrict application access
   - Regularly review which users have access to the worker application

3. **Token Management**
   - The MCP server stores tokens securely in the OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
   - Tokens are automatically refreshed and expired tokens are discarded
   - For containerized deployments, tokens are stored ephemerally

4. **Environment Separation**
   - Use separate worker applications for different environments (dev, test, prod)
   - Never use production worker applications in development/test environments

5. **Regular Audits**
   - Monitor PingOne audit logs for MCP server authentication events
   - Review API usage patterns and investigate anomalies

## Additional Resources

- [PingOne Documentation](https://docs.pingidentity.com/r/en-us/pingone/p1_c_pingone_overview)
- [Ping CLI Documentation](https://github.com/pingidentity/pingcli)
