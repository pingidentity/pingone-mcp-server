#!/bin/sh
set -e

# Check if already authenticated using the session command
echo "Checking for existing session..."
SESSION_OUTPUT=$(./pingone-mcp-server session --store-type=file 2>&1)

if echo "$SESSION_OUTPUT" | grep -q "No existing"; then
    echo "No existing session found. Authenticating with PingOne..."
    echo "Please complete the device authorization flow:"
    
    ./pingone-mcp-server login --grant-type=device_code --store-type=file
    
    if [ $? -ne 0 ]; then
        echo "Authentication failed. Please check your credentials and try again."
        exit 1
    fi
    
    echo "Authentication successful!"
else
    echo "Existing session found. Starting server..."
fi

echo "Starting MCP server..."
exec ./pingone-mcp-server run --store-type=file "$@"
