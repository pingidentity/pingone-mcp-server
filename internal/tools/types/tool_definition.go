// Copyright © 2025 Ping Identity Corporation

package types

import "github.com/modelcontextprotocol/go-sdk/mcp"

type ToolDefinition struct {
	// IsReadOnly indicates whether the tool is read-only (true) or can perform write operations on the API (false)
	IsReadOnly bool
	// McpTool is the MCP tool definition (including name and description)
	McpTool *mcp.Tool
}
