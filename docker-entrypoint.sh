#!/bin/sh
set -e

echo "Starting MCP server..."
exec ./pingone-mcp-server run --grant-type=device_code --store-type=file "$@"
