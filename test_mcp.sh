#!/bin/bash

# Test MCP server without credentials
unset PINGONE_CLIENT_ID PINGONE_CLIENT_SECRET PINGONE_ENV_ID

echo "Testing tools/list without credentials..."
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./server --transport stdio &
PID=$!
sleep 2
kill $PID 2>/dev/null
wait $PID 2>/dev/null

echo ""
echo "Testing initialize..."
echo '{"jsonrpc":"2.0","id":2,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"Claude Desktop","version":"1.0.0"}}}' | ./server --transport stdio &
PID=$!
sleep 2
kill $PID 2>/dev/null
wait $PID 2>/dev/null