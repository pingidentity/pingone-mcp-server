// Copyright © 2025 Ping Identity Corporation

package testutils

import tools "github.com/pingidentity/pingone-mcp-server/internal/capabilities"

func AllServerToolNames() []string {
	allTools := tools.ListTools()
	toolNames := make([]string, len(allTools))
	for i, tool := range allTools {
		toolNames[i] = tool.McpTool.Name
	}
	return toolNames
}

func ReadOnlyToolNames() []string {
	allTools := tools.ListTools()
	var toolNames []string
	for _, tool := range allTools {
		if tool.IsReadOnly() {
			toolNames = append(toolNames, tool.McpTool.Name)
		}
	}
	return toolNames
}

func WriteToolNames() []string {
	allTools := tools.ListTools()
	var toolNames []string
	for _, tool := range allTools {
		if !tool.IsReadOnly() {
			toolNames = append(toolNames, tool.McpTool.Name)
		}
	}
	return toolNames
}
