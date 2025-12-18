// Copyright © 2025 Ping Identity Corporation

package types

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestToolDefinition_IsReadOnly(t *testing.T) {
	tests := []struct {
		name     string
		tool     *ToolDefinition
		expected bool
	}{
		{
			name:     "nil tool definition returns true",
			tool:     nil,
			expected: true,
		},
		{
			name: "nil McpTool returns false",
			tool: &ToolDefinition{
				McpTool: nil,
			},
			expected: false,
		},
		{
			name: "nil Annotations returns false",
			tool: &ToolDefinition{
				McpTool: &mcp.Tool{
					Name:        "test-tool",
					Annotations: nil,
				},
			},
			expected: false,
		},
		{
			name: "ReadOnlyHint set to true returns true",
			tool: &ToolDefinition{
				McpTool: &mcp.Tool{
					Name: "test-tool",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "ReadOnlyHint set to false returns false",
			tool: &ToolDefinition{
				McpTool: &mcp.Tool{
					Name: "test-tool",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: false,
					},
				},
			},
			expected: false,
		},
		{
			name:     "empty ToolDefinition returns false",
			tool:     &ToolDefinition{},
			expected: false,
		},
		{
			name: "complete ToolDefinition with ReadOnlyHint true",
			tool: &ToolDefinition{
				McpTool: &mcp.Tool{
					Name:        "read-only-tool",
					Description: "A read-only tool",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: true,
					},
				},
				ValidationPolicy: &ToolValidationPolicy{},
			},
			expected: true,
		},
		{
			name: "complete ToolDefinition with ReadOnlyHint false",
			tool: &ToolDefinition{
				McpTool: &mcp.Tool{
					Name:        "write-tool",
					Description: "A tool that modifies state",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: false,
					},
				},
				ValidationPolicy: &ToolValidationPolicy{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tool.IsReadOnly()
			assert.Equal(t, tt.expected, result)
		})
	}
}
