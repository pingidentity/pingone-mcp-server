// Copyright © 2025 Ping Identity Corporation

package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/server"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultGrantType = auth.GrantTypeAuthorizationCode

func TestServer_MCPClient(t *testing.T) {
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverDone := make(chan error, 1)
	go func() {
		// Pass in dummy client for now, not testing tool functionality
		err := server.Start(context.Background(), serverTransport, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), testutils.NewInMemoryTokenStore(), filter.PassthroughFilter(), defaultGrantType)
		serverDone <- err
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	client := mcptestutils.TestMcpClient(t)

	session, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err, "MCP client should connect to server successfully")
	require.NotNil(t, session, "Session should not be nil")
	defer session.Close()

	toolsResult, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
	require.NoError(t, err, "ListTools should not return error")
	require.NotNil(t, toolsResult, "ListTools result should not be nil")
	assert.Greater(t, len(toolsResult.Tools), 0, "Server should have at least one tool")
}

func TestServer_ToolFiltering(t *testing.T) {
	tests := []struct {
		name                    string
		readOnly                bool
		includedTools           []string
		excludedTools           []string
		includedToolCollections []string
		excludedToolCollections []string
		expectedTools           []string
		unexpectedTools         []string
	}{
		{
			name:          "no filtering",
			expectedTools: testutils.AllServerToolNames(),
		},
		{
			name:          "inclusion",
			includedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
			expectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclusion",
			excludedTools:   []string{environments.ListEnvironmentsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclusion takes priority over inclusion",
			includedTools:   []string{environments.ListEnvironmentsDef.McpTool.Name},
			excludedTools:   []string{environments.ListEnvironmentsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:                    "include collection",
			includedToolCollections: []string{environments.CollectionName},
			expectedTools:           []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:                    "exclude collection",
			excludedToolCollections: []string{environments.CollectionName},
			unexpectedTools:         []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:          "read-only mode includes only read-only tools",
			readOnly:      true,
			expectedTools: testutils.ReadOnlyToolNames(),
		},
		{
			name:            "read-only mode excludes non-read-only tools",
			readOnly:        true,
			unexpectedTools: testutils.WriteToolNames(),
		},
		{
			name:            "read-only mode with included tools still filters by read-only",
			readOnly:        true,
			includedTools:   testutils.AllServerToolNames(),
			expectedTools:   testutils.ReadOnlyToolNames(),
			unexpectedTools: testutils.WriteToolNames(),
		},
		{
			name:            "read-only mode with excluded read-only tool",
			readOnly:        true,
			excludedTools:   []string{environments.ListEnvironmentsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverTransport, clientTransport := mcp.NewInMemoryTransports()

			serverDone := make(chan error, 1)
			go func() {
				toolFilter := filter.NewFilter(tt.readOnly, tt.includedTools, tt.excludedTools, tt.includedToolCollections, tt.excludedToolCollections)
				err := server.Start(context.Background(), serverTransport, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), testutils.NewInMemoryTokenStore(), toolFilter, defaultGrantType)
				serverDone <- err
			}()

			time.Sleep(100 * time.Millisecond)

			client := mcptestutils.TestMcpClient(t)

			session, err := client.Connect(t.Context(), clientTransport, nil)
			require.NoError(t, err)
			defer session.Close()

			toolsResult, err := session.ListTools(t.Context(), &mcp.ListToolsParams{})
			require.NoError(t, err)

			toolNames := make([]string, len(toolsResult.Tools))
			for i, tool := range toolsResult.Tools {
				toolNames[i] = tool.Name
			}

			for _, expectedTool := range tt.expectedTools {
				assert.Contains(t, toolNames, expectedTool)
			}

			for _, unexpectedTool := range tt.unexpectedTools {
				assert.NotContains(t, toolNames, unexpectedTool)
			}
		})
	}
}
