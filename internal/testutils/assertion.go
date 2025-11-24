// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertHandlerError is a helper to assert that a tool handler returned an expected error
func AssertHandlerError(t *testing.T, err error, mcpResult *mcp.CallToolResult, output interface{}, wantErrContains string) {
	t.Helper()
	require.Error(t, err)
	if wantErrContains != "" {
		assert.Contains(t, err.Error(), wantErrContains)
	}
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

// AssertHandlerSuccess is a helper to assert that a tool handler succeeded
func AssertHandlerSuccess(t *testing.T, err error, mcpResult *mcp.CallToolResult, output interface{}) {
	t.Helper()
	require.NoError(t, err)
	assert.Nil(t, mcpResult)
	require.NotNil(t, output)
}

// AssertMcpCallError is a helper to assert that an MCP call returned an expected error
func AssertMcpCallError(t *testing.T, result *mcp.CallToolResult, wantErrContains string) {
	t.Helper()
	assert.True(t, result.IsError)
	assert.GreaterOrEqual(t, len(result.Content), 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")
	if wantErrContains != "" {
		assert.Contains(t, textContent.Text, wantErrContains)
	}
}

// AssertMcpCallSuccess is a helper to assert that an MCP call succeeded
func AssertMcpCallSuccess(t *testing.T, err error, result *mcp.CallToolResult) {
	t.Helper()
	require.NoError(t, err)
	assert.False(t, result.IsError)
	require.NotNil(t, result)
}
