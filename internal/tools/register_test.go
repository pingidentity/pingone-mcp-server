// Copyright © 2025 Ping Identity Corporation

package tools_test

import (
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/tools"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/directory"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllToolsRegistered(t *testing.T) {
	// Get tools from ListTools that are actually registered with the server
	allTools := tools.ListTools()

	// Get tools from individual collections
	var expectedTools []types.ToolDefinition
	expectedTools = append(expectedTools, (&directory.DirectoryCollection{}).ListTools()...)
	expectedTools = append(expectedTools, (&environments.EnvironmentsCollection{}).ListTools()...)
	expectedTools = append(expectedTools, (&populations.PopulationsCollection{}).ListTools()...)
	expectedTools = append(expectedTools, (&applications.ApplicationsCollection{}).ListTools()...)

	// Verify lists match
	if len(allTools) != len(expectedTools) {
		t.Errorf("ListTools() returned %d tools, but individual collections returned %d tools", len(allTools), len(expectedTools))
	}

	expectedToolNames := make(map[string]bool)
	for _, tool := range expectedTools {
		expectedToolNames[tool.McpTool.Name] = true
	}

	for _, tool := range allTools {
		if !expectedToolNames[tool.McpTool.Name] {
			t.Errorf("ListTools() returned unexpected tool: %s", tool.McpTool.Name)
		}
	}

	actualToolNames := make(map[string]bool)
	for _, tool := range allTools {
		actualToolNames[tool.McpTool.Name] = true
	}

	for _, tool := range expectedTools {
		if !actualToolNames[tool.McpTool.Name] {
			t.Errorf("ListTools() missing expected tool, collection may not be registered: %s", tool.McpTool.Name)
		}
	}
}

func TestAllToolsHaveSchemas(t *testing.T) {
	for _, toolDef := range tools.ListTools() {
		t.Run(toolDef.McpTool.Name, func(t *testing.T) {
			require.NotNil(t, toolDef.McpTool, "McpTool should not be nil for tool %s", toolDef.McpTool.Name)
			assert.NotNil(t, toolDef.McpTool.InputSchema, "tool InputSchema should not be nil for tool %s", toolDef.McpTool.Name)
			assert.NotNil(t, toolDef.McpTool.OutputSchema, "tool OutputSchema should not be nil for tool %s", toolDef.McpTool.Name)
		})
	}
}
