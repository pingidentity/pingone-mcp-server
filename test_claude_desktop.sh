#!/bin/bash

# Test script to simulate Claude Desktop connecting to our MCP server

echo "Testing MCP server compatibility with Claude Desktop..."
echo

# Set environment variables (you'll need to replace these)
export PINGONE_CLIENT_ID="test-client-id"
export PINGONE_CLIENT_SECRET="test-client-secret"  
export PINGONE_ENV_ID="test-env-id"
export PINGONE_REGION="com"
export MCP_TRANSPORT="stdio"

# Test initialization
echo "1. Testing initialization..."
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"Claude Desktop","version":"1.0.0"}}}' | timeout 3s ./server 2>/dev/null | head -1

echo
echo "2. Testing tools list..."
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | timeout 3s ./server 2>/dev/null | head -1

echo
echo "Note: Tests may fail due to missing PingOne credentials, but JSON-RPC protocol should work"