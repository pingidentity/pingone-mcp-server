package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// stubErrorTool implements tools.Tool for testing error scenarios
type stubErrorTool struct{}

func (s *stubErrorTool) Name() string        { return "error_tool" }
func (s *stubErrorTool) Description() string { return "a tool that always errors" }
func (s *stubErrorTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"input":{"type":"string"}},"required":["input"]}`)
}
func (s *stubErrorTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("this tool always fails")
}

// TestNewJSONRPCError tests error response creation
func TestNewJSONRPCError(t *testing.T) {
	tests := []struct {
		name    string
		id      interface{}
		code    int
		message string
		data    interface{}
	}{
		{
			name:    "basic error with string ID",
			id:      "test-id",
			code:    ErrorCodeInternalError,
			message: "Internal error",
			data:    nil,
		},
		{
			name:    "error with numeric ID",
			id:      42,
			code:    ErrorCodeMethodNotFound,
			message: "Method not found",
			data:    nil,
		},
		{
			name:    "error with data",
			id:      nil,
			code:    ErrorCodeInvalidParams,
			message: "Invalid parameters",
			data:    map[string]string{"detail": "missing required field"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewJSONRPCError(tt.id, tt.code, tt.message, tt.data)
			
			if resp.JSONRPC != "2.0" {
				t.Errorf("expected JSONRPC 2.0, got %s", resp.JSONRPC)
			}
			if resp.ID != tt.id {
				t.Errorf("expected ID %v, got %v", tt.id, resp.ID)
			}
			if resp.Error == nil {
				t.Fatal("expected error to be non-nil")
			}
			if resp.Error.Code != tt.code {
				t.Errorf("expected error code %d, got %d", tt.code, resp.Error.Code)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("expected error message %q, got %q", tt.message, resp.Error.Message)
			}
			// For data comparison, use a different approach since maps are not comparable
			if tt.data == nil && resp.Error.Data != nil {
				t.Errorf("expected error data to be nil, got %v", resp.Error.Data)
			} else if tt.data != nil && resp.Error.Data == nil {
				t.Errorf("expected error data %v, got nil", tt.data)
			} else if tt.data != nil && resp.Error.Data != nil {
				// For maps, we can marshal and compare JSON
				expectedJSON, _ := json.Marshal(tt.data)
				actualJSON, _ := json.Marshal(resp.Error.Data)
				if string(expectedJSON) != string(actualJSON) {
					t.Errorf("expected error data %v, got %v", tt.data, resp.Error.Data)
				}
			}
		})
	}
}

// TestParseJSONRPCRequest_ErrorCases tests various parsing error scenarios
func TestParseJSONRPCRequest_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "missing jsonrpc field",
			input:   `{"id":1,"method":"test"}`,
			wantErr: true,
		},
		{
			name:    "wrong jsonrpc version",
			input:   `{"jsonrpc":"1.0","id":1,"method":"test"}`,
			wantErr: true,
		},
		{
			name:    "missing method field",
			input:   `{"jsonrpc":"2.0","id":1}`,
			wantErr: true,
		},
		{
			name:    "valid request",
			input:   `{"jsonrpc":"2.0","id":1,"method":"test","params":{}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseJSONRPCRequest([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONRPCRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHandleToolsCall tests the tools/call JSON-RPC method handler
func TestHandleToolsCall(t *testing.T) {
	// Register test tools
	tools.Reset()
	tools.Register(&stubTool{})
	tools.Register(&stubErrorTool{})

	handler := NewJSONRPCHandler("0.1.0")

	tests := []struct {
		name           string
		request        JSONRPCRequest
		expectError    bool
		expectedErrCode int
	}{
		{
			name: "successful tool call",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "echo",
					"arguments": map[string]interface{}{
						"msg": "hello world",
					},
				},
			},
			expectError: false,
		},
		{
			name: "tool not found",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "nonexistent",
					"arguments": map[string]interface{}{},
				},
			},
			expectError:     true,
			expectedErrCode: ErrorCodeMethodNotFound,
		},
		{
			name: "tool execution error",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "error_tool",
					"arguments": map[string]interface{}{
						"input": "test",
					},
				},
			},
			expectError:     true,
			expectedErrCode: ErrorCodeInternalError,
		},
		{
			name: "missing tool name",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      4,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"arguments": map[string]interface{}{},
				},
			},
			expectError:     true,
			expectedErrCode: ErrorCodeInvalidParams,
		},
		{
			name: "missing arguments",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      5,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "echo",
					// missing arguments field - should trigger tool execution error since tool requires "msg" 
				},
			},
			expectError:     true,
			expectedErrCode: ErrorCodeInternalError, // tool execution will fail, not invalid params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := handler.HandleRequest(context.Background(), &tt.request)
			
			if tt.expectError {
				if response.Error == nil {
					t.Fatal("expected error, got none")
				}
				if response.Error.Code != tt.expectedErrCode {
					t.Errorf("expected error code %d, got %d", tt.expectedErrCode, response.Error.Code)
				}
			} else {
				if response.Error != nil {
					t.Fatalf("unexpected error: %v", response.Error)
				}
				if response.Result == nil {
					t.Fatal("expected result, got nil")
				}
			}
		})
	}
}

// TestConvertToolOutputToMCPResult tests output conversion
func TestConvertToolOutputToMCPResult(t *testing.T) {
	handler := NewJSONRPCHandler("0.1.0")
	
	tests := []struct {
		name    string
		output  map[string]interface{}
		isError bool
	}{
		{
			name:    "simple output",
			output:  map[string]interface{}{"key": "value"},
			isError: false,
		},
		{
			name: "complex output",
			output: map[string]interface{}{
				"string": "test",
				"number": 42,
				"array":  []interface{}{1, 2, 3},
				"object": map[string]interface{}{"nested": true},
			},
			isError: false,
		},
		{
			name:    "error output",
			output:  map[string]interface{}{"error": "something went wrong"},
			isError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.convertToolOutputToMCPResult(tt.output, tt.isError)
			
			// Should always return a ToolCallResult
			if len(result.Content) == 0 {
				t.Fatal("expected non-empty content array")
			}
			
			if result.IsError != tt.isError {
				t.Errorf("expected IsError %v, got %v", tt.isError, result.IsError)
			}
			
			// First content item should be text
			if result.Content[0].Type != "text" {
				t.Errorf("expected content type 'text', got %s", result.Content[0].Type)
			}
			
			if result.Content[0].Text == "" {
				t.Error("expected non-empty text content")
			}
		})
	}
}

// TestFormatToolOutput tests output formatting
func TestFormatToolOutput(t *testing.T) {
	handler := NewJSONRPCHandler("0.1.0")
	
	tests := []struct {
		name           string
		output         map[string]interface{}
		expectsJSON    bool
		expectedPrefix string
	}{
		{
			name:           "simple string output",
			output:         map[string]interface{}{"message": "hello"},
			expectsJSON:    false, // formatToolOutput returns a formatted string, not pure JSON
			expectedPrefix: "Operation completed successfully.",
		},
		{
			name: "complex nested output",
			output: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John Doe",
				},
				"status": "active",
			},
			expectsJSON:    false,
			expectedPrefix: "Operation completed successfully.",
		},
		{
			name:           "empty output",
			output:         map[string]interface{}{},
			expectsJSON:    false,
			expectedPrefix: "Operation completed successfully.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.formatToolOutput(tt.output)
			
			if result == "" {
				t.Error("expected non-empty formatted output")
			}
			
			// Check that it starts with expected prefix
			if !strings.HasPrefix(result, tt.expectedPrefix) {
				t.Errorf("expected output to start with %q, got %q", tt.expectedPrefix, result)
			}
		})
	}
}

// TestJSONRPCHandler_HandleRequest_EdgeCases tests edge cases in request handling
func TestJSONRPCHandler_HandleRequest_EdgeCases(t *testing.T) {
	handler := NewJSONRPCHandler("0.1.0")

	tests := []struct {
		name    string
		request *JSONRPCRequest
		wantErr bool
		errCode int
	}{
		{
			name: "invalid method format",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "invalidmethod",
				Params:  map[string]interface{}{},
			},
			wantErr: true,
			errCode: ErrorCodeMethodNotFound,
		},
		{
			name: "initialize with invalid params",
			request: &JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params:  "invalid params type",
			},
			wantErr: true,
			errCode: ErrorCodeInvalidParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := handler.HandleRequest(context.Background(), tt.request)
			
			if tt.wantErr {
				if response.Error == nil {
					t.Fatal("expected error, got none")
				}
				if response.Error.Code != tt.errCode {
					t.Errorf("expected error code %d, got %d", tt.errCode, response.Error.Code)
				}
			} else {
				if response.Error != nil {
					t.Fatalf("unexpected error: %v", response.Error)
				}
			}
		})
	}
}