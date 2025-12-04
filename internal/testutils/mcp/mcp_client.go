// Copyright © 2025 Ping Identity Corporation

package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// TestMcpClient creates a test MCP client for testing purposes.
func TestMcpClient(t *testing.T) *mcp.Client {
	t.Helper()
	return mcp.NewClient(&mcp.Implementation{
		Name:    "test-mcp-client",
		Version: "v0.0.1-test",
	}, nil)
}

// TestMcpServer creates a test MCP server for testing purposes.
func TestMcpServer(t *testing.T) *mcp.Server {
	t.Helper()
	return mcp.NewServer(&mcp.Implementation{
		Name:    "test-pingone-mcp-server",
		Version: "v0.0.1-test",
	}, nil)
}

// CallToolOverMcp invokes a tool through a full MCP client-server connection.
func CallToolOverMcp(t *testing.T, server *mcp.Server, toolName string, toolInput any) (*mcp.CallToolResult, error) {
	t.Helper()

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverDone := make(chan error, 1)
	go func() {
		err := server.Run(ctx, serverTransport)
		serverDone <- err
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	client := TestMcpClient(t)

	session, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err, "MCP client should connect to server successfully")
	require.NotNil(t, session, "Session should not be nil")
	defer func() {
		if closeErr := session.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close session: %v", closeErr)
		}
	}()

	// Call the tool via MCP
	result, err := session.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      toolName,
		Arguments: toolInput,
	})

	// Wait for server to finish
	cancel()
	select {
	case <-serverDone:
		// Server stopped
	case <-time.After(500 * time.Millisecond):
		t.Error("Test MCP server did not stop as expected")
	}

	return result, err
}
