// Copyright © 2025 Ping Identity Corporation

package types

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type StaticResourceDefinition struct {
	// McpResource is the MCP resource definition (including name and description)
	McpResource *mcp.Resource
}
