// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationsCollection_Name(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	assert.Equal(t, "applications", collection.Name())
}

func TestApplicationsCollection_ListTools(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	tools := collection.ListTools()

	// Verify we have tools registered
	assert.NotEmpty(t, tools, "Should have at least one tool registered")

	// Verify all tools have unique names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		assert.False(t, toolNames[tool.McpTool.Name], "Tool name %s should be unique", tool.McpTool.Name)
		toolNames[tool.McpTool.Name] = true
	}
}

func TestApplicationsCollection_RegisterTools_NilClientFactory(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "v0.0.1",
	}, nil)
	toolFilter := filter.PassthroughFilter()

	// Attempt to register tools with nil client
	err := collection.RegisterTools(t.Context(), server, nil, testutils.NewInMemoryTokenStore(), toolFilter)

	// Should return an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PingOne API client factory is nil")
}

func TestApplicationsCollection_RegisterTools_NilTokenStore(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "v0.0.1",
	}, nil)
	toolFilter := filter.PassthroughFilter()

	// Attempt to register tools with nil token store
	err := collection.RegisterTools(t.Context(), server, legacy.NewEmptyClientFactory(), nil, toolFilter)

	// Should return an error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token store is nil")
}

func TestApplicationsCollection_RegisterTools_ReadOnlyToolsMarkedCorrectly(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	tools := collection.ListTools()

	// Define known read-only tools
	readOnlyTools := []string{
		"list_applications",
		"get_application",
	}

	// Define known write tools
	writeTools := []string{
		"create_oidc_application",
		"update_oidc_application",
	}

	for _, tool := range tools {
		inReadOnly := slices.Contains(readOnlyTools, tool.McpTool.Name)
		inWrite := slices.Contains(writeTools, tool.McpTool.Name)

		// Every tool must be categorized as either read-only or write
		assert.True(t, inReadOnly || inWrite,
			"Tool %s must be categorized as either read-only or write in this test", tool.McpTool.Name)

		if inReadOnly {
			assert.True(t, tool.IsReadOnly(), "Tool %s should be marked as read-only", tool.McpTool.Name)
		}
		if inWrite {
			assert.False(t, tool.IsReadOnly(), "Tool %s should NOT be marked as read-only", tool.McpTool.Name)
		}
	}
}

func TestApplicationsCollection_ToolDefinitionsHaveRequiredFields(t *testing.T) {
	collection := &applications.ApplicationsCollection{}
	tools := collection.ListTools()

	for _, tool := range tools {
		t.Run(tool.McpTool.Name, func(t *testing.T) {
			// Check that the tool definition is valid
			assert.NotNil(t, tool.McpTool, "McpTool should not be nil")

			// Every tool should have a name
			assert.NotEmpty(t, tool.McpTool.Name, "Tool name should not be empty")

			// Every tool should have a description
			assert.NotEmpty(t, tool.McpTool.Description, "Tool description should not be empty")

			// Tool names should follow kebab-case convention
			assert.NotContains(t, tool.McpTool.Name, "pingone", "Tool name should not contain 'pingone'")
			assert.NotContains(t, tool.McpTool.Name, "-", "Tool name should use snake_case, not kebab-case")
			assert.NotContains(t, tool.McpTool.Name, " ", "Tool name should not contain spaces")
		})
	}
}
