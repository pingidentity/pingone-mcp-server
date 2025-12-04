// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"encoding/json"
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

// AssertStructuredHandlerSuccess is a helper to assert that a tool handler that returns structured output succeeded
func AssertStructuredHandlerSuccess(t *testing.T, err error, mcpResult *mcp.CallToolResult, output interface{}) {
	t.Helper()
	require.NoError(t, err)
	assert.Nil(t, mcpResult)
	require.NotNil(t, output)
}

// AssertUnstructuredHandlerSuccess is a helper to assert that a tool handler that returns unstructured output succeeded
func AssertUnstructuredHandlerSuccess(t *testing.T, err error, mcpResult *mcp.CallToolResult, output interface{}) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, mcpResult)
	require.Nil(t, output)
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

func AssertUnstructuredOutputMatches(t *testing.T, mcpResult *mcp.CallToolResult, expectedOutput any) {
	t.Helper()
	require.Len(t, mcpResult.Content, 1, "Expected exactly one content item in output")
	textContent, ok := mcpResult.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")
	expectedJsonBytes, err := json.Marshal(expectedOutput)
	require.NoError(t, err, "Failed to marshal expected response")

	assert.JSONEq(t, string(expectedJsonBytes), string(textContent.Text), "Output JSON should match expected JSON")
}
