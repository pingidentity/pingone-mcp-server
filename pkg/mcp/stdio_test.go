package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStdioServer_Initialize(t *testing.T) {
	// Create input/output buffers
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`)
	output := &bytes.Buffer{}

	// Create stdio server
	server := NewStdioServerWithIO("0.1.0", input, output)

	// Run server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected since we only sent one request
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(output).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.ID != float64(1) { // JSON unmarshaling makes numbers float64
		t.Errorf("Expected id 1, got %v", response.ID)
	}

	if response.Error != nil {
		t.Errorf("Unexpected error: %v", response.Error)
	}

	// Verify result has server info
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be object, got %T", response.Result)
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo in result, got %v", result)
	}

	if serverInfo["name"] != "PingOne MCP Server" {
		t.Errorf("Expected server name 'PingOne MCP Server', got %v", serverInfo["name"])
	}
}

func TestStdioServer_ToolsList(t *testing.T) {
	// Create input/output buffers
	input := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	output := &bytes.Buffer{}

	// Create stdio server
	server := NewStdioServerWithIO("0.1.0", input, output)

	// Run server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(output).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response
	if response.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.Error != nil {
		t.Errorf("Unexpected error: %v", response.Error)
	}

	// Verify result has tools array
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be object, got %T", response.Result)
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected tools array in result, got %v", result)
	}

	// Should have empty tools list since no tools are registered in test
	if len(tools) != 0 {
		t.Logf("Found %d tools (expected 0 in test environment)", len(tools))
	}
}

// TestNewStdioServer tests the basic constructor
func TestNewStdioServer(t *testing.T) {
	server := NewStdioServer("0.1.0")
	if server == nil {
		t.Fatal("NewStdioServer returned nil")
	}
}

// TestStdioServer_InvalidJSON tests handling of malformed JSON input
func TestStdioServer_InvalidJSON(t *testing.T) {
	input := strings.NewReader(`{invalid json}`)
	output := &bytes.Buffer{}

	server := NewStdioServerWithIO("0.1.0", input, output)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected
	}

	// Should have an error response
	var response JSONRPCResponse
	if err := json.NewDecoder(output).Decode(&response); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestStdioServer_UnknownMethod tests handling of unknown JSON-RPC method
func TestStdioServer_UnknownMethod(t *testing.T) {
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"unknown/method","params":{}}`)
	output := &bytes.Buffer{}

	server := NewStdioServerWithIO("0.1.0", input, output)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil && err != context.DeadlineExceeded {
			t.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		// Timeout is expected
	}

	// Should have an error response
	var response JSONRPCResponse
	if err := json.NewDecoder(output).Decode(&response); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}

	if response.Error == nil {
		t.Error("Expected error for unknown method")
	}
	if response.Error.Code != ErrorCodeMethodNotFound {
		t.Errorf("Expected method not found error code %d, got %d", ErrorCodeMethodNotFound, response.Error.Code)
	}
}