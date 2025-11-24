// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMcpClient(t *testing.T) *mcp.Client {
	t.Helper()
	return mcp.NewClient(&mcp.Implementation{
		Name:    "test-mcp-client",
		Version: "v0.0.1-test",
	}, nil)
}

func TestMcpServer(t *testing.T) *mcp.Server {
	t.Helper()
	return mcp.NewServer(&mcp.Implementation{
		Name:    "test-pingone-mcp-server",
		Version: "v0.0.1-test",
	}, nil)
}

func CallToolOverMcp(t *testing.T, server *mcp.Server, toolName string, toolInput any) (*mcp.CallToolResult, error) {
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
	defer session.Close()

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
