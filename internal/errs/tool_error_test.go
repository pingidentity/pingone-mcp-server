// Copyright © 2025 Ping Identity Corporation

package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/errs"
)

func TestToolError_Error(t *testing.T) {
	tests := []struct {
		name          string
		toolName      string
		originalError error
		expected      string
	}{
		{
			name:     "no error and no tool name",
			expected: "unknown tool error",
		},
		{
			name:     "no error with tool name",
			toolName: "list_environments",
			expected: "pingone-mcp-server list_environments tool failed: unknown error",
		},
		{
			name:          "original error only (empty tool name)",
			toolName:      "",
			originalError: errors.New("connection failed"),
			expected:      "connection failed",
		},
		{
			name:          "original error with tool name",
			toolName:      "list_populations",
			originalError: errors.New("authentication failed"),
			expected:      "pingone-mcp-server list_populations tool failed: authentication failed",
		},
		{
			name:     "nil error with tool name",
			toolName: "test-tool",
			expected: "pingone-mcp-server test-tool tool failed: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolErr := &errs.ToolError{
				ToolName:      tt.toolName,
				OriginalError: tt.originalError,
			}

			result := toolErr.Error()
			if result != tt.expected {
				t.Errorf("Expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}

func TestNewToolError(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		originalErr  error
		expectedType string
		checkFunc    func(t *testing.T, err error)
	}{
		{
			name:         "nil error",
			toolName:     "test-tool",
			originalErr:  nil,
			expectedType: "*errs.ToolError",
			checkFunc: func(t *testing.T, err error) {
				toolErr := err.(*errs.ToolError)
				if toolErr.ToolName != "test-tool" {
					t.Errorf("Expected tool name 'test-tool', got: %s", toolErr.ToolName)
				}
				if toolErr.OriginalError != nil {
					t.Errorf("Expected nil original error, got: %v", toolErr.OriginalError)
				}
			},
		},
		{
			name:         "with original error",
			toolName:     "list_environments",
			originalErr:  errors.New("database connection failed"),
			expectedType: "*errs.ToolError",
			checkFunc: func(t *testing.T, err error) {
				toolErr := err.(*errs.ToolError)
				if toolErr.ToolName != "list_environments" {
					t.Errorf("Expected tool name 'list_environments', got: %s", toolErr.ToolName)
				}
				if toolErr.OriginalError == nil || toolErr.OriginalError.Error() != "database connection failed" {
					t.Errorf("Expected original error 'database connection failed', got: %v", toolErr.OriginalError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.NewToolError(tt.toolName, tt.originalErr)

			if err == nil {
				t.Fatal("Expected non-nil error")
			}

			// Check type
			if got := fmt.Sprintf("%T", err); got != tt.expectedType {
				t.Errorf("Expected type %s, got: %s", tt.expectedType, got)
			}

			// Run custom checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}
