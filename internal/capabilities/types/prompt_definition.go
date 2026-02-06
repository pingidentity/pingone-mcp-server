// Copyright © 2026 Ping Identity Corporation

package types

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PromptDefinition struct {
	// McpPrompt is the MCP prompt definition (including name and description)
	McpPrompt *mcp.Prompt
}
