// Copyright © 2025 Ping Identity Corporation

package types

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolDefinition struct {
	// IsReadOnly indicates whether the tool is read-only (true) or can perform write operations on the API (false)
	IsReadOnly bool
	// McpTool is the MCP tool definition (including name and description)
	McpTool *mcp.Tool
	// Validation allows modification of in-built validation rules and constraints for the tool's execution
	Validation *ToolValidation
}

type ToolValidation struct {
	// SkipProductionEnvironmentWriteRestriction when set to true, allows the tool to make write operations on production-type environments.
	// Typically used where the tool itself performs validation, is trusted or is not acting on environments.
	// Defaults to false, meaning production environments are protected by default.
	SkipProductionEnvironmentWriteRestriction   bool
	EnforceProductionEnvironmentReadRestriction bool
}
