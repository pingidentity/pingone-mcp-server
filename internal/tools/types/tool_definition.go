// Copyright © 2025 Ping Identity Corporation

package types

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolDefinition struct {
	// McpTool is the MCP tool definition (including name and description)
	McpTool *mcp.Tool
	// ValidationPolicy allows modification of in-built validation rules and constraints for the tool's execution
	ValidationPolicy *ToolValidationPolicy
}

// IsReadOnly returns true if the tool is read-only and does not modify its environment.
// It checks the McpTool.Annotations.ReadOnlyHint field.
func (t *ToolDefinition) IsReadOnly() bool {
	if t == nil {
		return true
	}
	if t.McpTool != nil && t.McpTool.Annotations != nil {
		return t.McpTool.Annotations.ReadOnlyHint
	}
	// Default to false (not read-only) if annotations are not set
	return false
}
