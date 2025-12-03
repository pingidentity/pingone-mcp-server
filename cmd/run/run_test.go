// Copyright © 2025 Ping Identity Corporation

package run_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand_FromRoot_NoServerRun(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "run help flag",
			args:        []string{"run", "--help"},
			expectError: false,
			description: "Run command help should execute without error",
		},
		{
			name:          "run invalid flag",
			args:          []string{"run", "--invalid-flag"},
			expectError:   true,
			errorContains: "unknown flag",
			description:   "Run command should return error for invalid flag",
		},
		{
			name:          "run invalid grant-type value",
			args:          []string{"run", "--grant-type", "invalid"},
			expectError:   true,
			errorContains: "unable to parse grant type",
			description:   "Run command should return error for invalid flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// The server will exit immediately here, but the command can still be run
			err := testutils.ExecuteCliRootCommand(t, ctx, tt.args...)

			if tt.expectError {
				require.Error(t, err, tt.description)
				if tt.errorContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errorContains),
						"Error should contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				require.NoError(t, err, tt.description)
			}
		})
	}
}

func TestRunCommand_FromSubcommand_RunServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStoreWithDefaultSession()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	r, w, _ := os.Pipe()
	os.Stdin = r
	os.Stdout = w

	// Run the server in a goroutine so the test doesn't block.
	var wg sync.WaitGroup
	wg.Go(func() {
		err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), &mcp.StdioTransport{})
		assert.ErrorIs(t, err, context.Canceled, "server should stop due to context cancellation")
	})

	// Give the server a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to signal the server to shut down.
	cancel()
	wg.Wait()
	tokenStoreFactory.AssertExpectations(t)
}

func TestRunCommand_FromSubcommand_NoValidSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r, w, _ := os.Pipe()
	os.Stdin = r
	os.Stdout = w

	emptyTokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(emptyTokenStore)

	// Run the server in a goroutine so the test doesn't block.
	var wg sync.WaitGroup
	wg.Go(func() {
		err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), &mcp.StdioTransport{})
		assert.ErrorIs(t, err, context.Canceled, "server should stop due to context cancellation")
	})

	// Give the server a moment to start up.
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to signal the server to shut down.
	cancel()
	wg.Wait()
	tokenStoreFactory.AssertExpectations(t)
}

func TestRunCommand_FromSubcommand_TokenStoreFactoryError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedError := assert.AnError
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithError(expectedError)

	err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), &mcp.StdioTransport{})
	require.Error(t, err, "Run should fail when token store factory returns error")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the factory error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestRunCommand_FromSubcommand_StoreTypeSelection(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedStoreType tokenstore.StoreType
		description       string
	}{
		{
			name:              "default store type is keychain",
			args:              []string{},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Run should use keychain store type by default",
		},
		{
			name:              "explicit keychain store type",
			args:              []string{"--store-type", "keychain"},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Run should use keychain store type when explicitly specified",
		},
		{
			name:              "file store type",
			args:              []string{"--store-type", "file"},
			expectedStoreType: tokenstore.StoreTypeFile,
			description:       "Run should use file store type when specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStoreWithDefaultSession()
			tokenStoreFactory := testutils.NewMockTokenStoreFactory()
			tokenStoreFactory.On("NewTokenStore", tt.expectedStoreType).Return(tokenStore, nil)

			r, w, _ := os.Pipe()
			os.Stdin = r
			os.Stdout = w

			// Run the server in a goroutine so the test doesn't block.
			var wg sync.WaitGroup
			wg.Go(func() {
				err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), &mcp.StdioTransport{}, tt.args...)
				assert.ErrorIs(t, err, context.Canceled, "server should stop due to context cancellation")
			})

			// Give the server a moment to start up.
			time.Sleep(100 * time.Millisecond)

			// Cancel the context to signal the server to shut down.
			cancel()
			wg.Wait()
			// Verify token store was created with expected store type
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestRunCommand_FromSubcommand_InvalidStoreType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStoreFactory := testutils.NewMockTokenStoreFactory()

	err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), &mcp.StdioTransport{}, "--store-type", "invalid")
	require.Error(t, err, "Run should fail with invalid store type")
	assert.Contains(t, err.Error(), "unable to parse store type from string: invalid", "Error should indicate invalid store type")
}

func TestRunCommand_FromSubcommand_ToolFiltering(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectError     bool
		errorContains   string
		expectedTools   []string
		unexpectedTools []string
	}{
		{
			name:            "no filtering defaults to read-only mode",
			args:            []string{"run"},
			expectedTools:   testutils.ReadOnlyToolNames(),
			unexpectedTools: testutils.WriteToolNames(),
		},
		{
			name:          "inclusion",
			args:          []string{"run", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name},
			expectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclusion",
			args:            []string{"run", "--exclude-tools", environments.ListEnvironmentsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclusion takes priority over inclusion",
			args:            []string{"run", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name, "--exclude-tools", environments.ListEnvironmentsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:          "include collection",
			args:          []string{"run", "--include-tool-collections", environments.CollectionName},
			expectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclude collection",
			args:            []string{"run", "--exclude-tool-collections", environments.CollectionName},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "collection exclusion takes priority over inclusion",
			args:            []string{"run", "--include-tool-collections", environments.CollectionName, "--exclude-tool-collections", environments.CollectionName},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "exclude collection overrides included tools",
			args:            []string{"run", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name, "--exclude-tool-collections", environments.CollectionName},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:          "include multiple tools",
			args:          []string{"run", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name + "," + populations.ListPopulationsDef.McpTool.Name},
			expectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name, populations.ListPopulationsDef.McpTool.Name},
		},
		{
			name:            "exclude multiple tools",
			args:            []string{"run", "--exclude-tools", environments.ListEnvironmentsDef.McpTool.Name + "," + populations.ListPopulationsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name, populations.ListPopulationsDef.McpTool.Name},
		},
		{
			name:          "include multiple collections",
			args:          []string{"run", "--include-tool-collections", environments.CollectionName + "," + populations.CollectionName},
			expectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name, populations.ListPopulationsDef.McpTool.Name},
		},
		{
			name:            "exclude multiple collections",
			args:            []string{"run", "--exclude-tool-collections", environments.CollectionName + "," + populations.CollectionName},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name, populations.ListPopulationsDef.McpTool.Name},
		},
		{
			name:            "mixed inclusion exclusion with multiple tools",
			args:            []string{"run", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name + "," + populations.ListPopulationsDef.McpTool.Name, "--exclude-tools", environments.ListEnvironmentsDef.McpTool.Name},
			expectedTools:   []string{populations.ListPopulationsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "mixed inclusion exclusion with multiple collections",
			args:            []string{"run", "--include-tool-collections", environments.CollectionName + "," + populations.CollectionName, "--exclude-tool-collections", environments.CollectionName},
			expectedTools:   []string{populations.ListPopulationsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:          "disable-read-only flag includes all tools",
			args:          []string{"run", "--disable-read-only"},
			expectedTools: testutils.AllServerToolNames(),
		},
		{
			name:            "disable-read-only mixed with tool filtering",
			args:            []string{"run", "--disable-read-only", "--include-tools", environments.ListEnvironmentsDef.McpTool.Name + "," + populations.ListPopulationsDef.McpTool.Name, "--exclude-tools", environments.ListEnvironmentsDef.McpTool.Name},
			expectedTools:   []string{populations.ListPopulationsDef.McpTool.Name},
			unexpectedTools: []string{environments.ListEnvironmentsDef.McpTool.Name},
		},
		{
			name:            "write tool excluded in read-only mode by default",
			args:            []string{"run"},
			unexpectedTools: []string{environments.CreateEnvironmentDef.McpTool.Name},
		},
		{
			name:          "write tool included when disable-read-only flag is set",
			args:          []string{"run", "--disable-read-only"},
			expectedTools: []string{environments.CreateEnvironmentDef.McpTool.Name},
		},
		{
			name:            "write tool explicitly included but still excluded in read-only mode",
			args:            []string{"run", "--include-tools", environments.CreateEnvironmentDef.McpTool.Name},
			unexpectedTools: []string{environments.CreateEnvironmentDef.McpTool.Name},
		},
		{
			name:          "write tool explicitly included and allowed with disable-read-only",
			args:          []string{"run", "--disable-read-only", "--include-tools", environments.CreateEnvironmentDef.McpTool.Name},
			expectedTools: []string{environments.CreateEnvironmentDef.McpTool.Name},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverTransport, clientTransport := mcp.NewInMemoryTransports()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStoreWithDefaultSession()
			tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

			var wg sync.WaitGroup
			wg.Go(func() {
				err := testutils.ExecuteCliRunCommand(t, ctx, tokenStoreFactory, sdk.NewEmptyClientFactory(), legacy.NewEmptyClientFactory(), testutils.NewEmptyMockAuthClientFactory(), serverTransport, tt.args...)
				assert.ErrorIs(t, err, context.Canceled, "server should stop due to context cancellation")
			})

			// Give the server a moment to start up.
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

			// Cancel the context to signal the server to shut down.
			cancel()
			wg.Wait()
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}
