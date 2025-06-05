package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pingidentity/pingone-mcp-server/pkg/tools"
)

// JSONRPCHandler processes JSON-RPC requests and returns responses
type JSONRPCHandler struct {
	serverVersion string
}

// NewJSONRPCHandler creates a new JSON-RPC handler
func NewJSONRPCHandler(serverVersion string) *JSONRPCHandler {
	return &JSONRPCHandler{
		serverVersion: serverVersion,
	}
}

// HandleRequest processes a JSON-RPC request and returns a response
func (h *JSONRPCHandler) HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(req)
	case "tools/list":
		return h.handleToolsList(req)
	case "tools/call":
		return h.handleToolsCall(ctx, req)
	default:
		return NewJSONRPCError(req.ID, ErrorCodeMethodNotFound, 
			fmt.Sprintf("method not found: %s", req.Method), nil)
	}
}

// handleInitialize processes the MCP initialize method
func (h *JSONRPCHandler) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	var params InitializeParams
	if req.Params != nil {
		if paramsBytes, err := json.Marshal(req.Params); err != nil {
			return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "invalid params", err.Error())
		} else if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "invalid params", err.Error())
		}
	}

	// For now, we accept any protocol version but respond with our supported version
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "PingOne MCP Server",
			Version: h.serverVersion,
		},
	}

	return NewJSONRPCResponse(req.ID, result)
}

// handleToolsList processes the tools/list method
func (h *JSONRPCHandler) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	toolList := tools.List()
	mcpTools := make([]MCPTool, 0, len(toolList))
	
	for _, t := range toolList {
		mcpTools = append(mcpTools, MCPTool{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}

	result := MCPToolsResult{
		Tools: mcpTools,
	}

	return NewJSONRPCResponse(req.ID, result)
}

// handleToolsCall processes the tools/call method
func (h *JSONRPCHandler) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params ToolCallParams
	if req.Params == nil {
		return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "missing params", nil)
	}

	if paramsBytes, err := json.Marshal(req.Params); err != nil {
		return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "invalid params", err.Error())
	} else if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "invalid params", err.Error())
	}

	if params.Name == "" {
		return NewJSONRPCError(req.ID, ErrorCodeInvalidParams, "missing tool name", nil)
	}

	// Look up tool
	tool, ok := tools.Get(params.Name)
	if !ok {
		return NewJSONRPCError(req.ID, ErrorCodeMethodNotFound, 
			fmt.Sprintf("tool not found: %s", params.Name), nil)
	}

	// Execute tool
	output, err := tool.Run(ctx, params.Arguments)
	if err != nil {
		// Convert tool execution error to MCP error format
		return NewJSONRPCError(req.ID, ErrorCodeInternalError, 
			fmt.Sprintf("tool execution failed: %s", err.Error()), map[string]interface{}{
				"tool": params.Name,
				"arguments": params.Arguments,
			})
	}

	// Convert output to MCP content format
	result := h.convertToolOutputToMCPResult(output, false)
	return NewJSONRPCResponse(req.ID, result)
}

// convertToolOutputToMCPResult converts tool output to MCP-compliant content format
func (h *JSONRPCHandler) convertToolOutputToMCPResult(output map[string]interface{}, isError bool) ToolCallResult {
	// Create a more user-friendly text representation
	var text string
	
	if isError {
		text = fmt.Sprintf("Tool execution failed: %v", output)
	} else {
		text = h.formatToolOutput(output)
	}

	return ToolCallResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: isError,
	}
}

// formatToolOutput creates a JSON representation of tool output
func (h *JSONRPCHandler) formatToolOutput(output map[string]interface{}) string {
	// Always return the complete JSON payload for full transparency
	if len(output) == 0 {
		return "Operation completed successfully."
	}

	// Return the full JSON payload with pretty formatting
	if jsonOutput, err := json.MarshalIndent(output, "", "  "); err == nil {
		return fmt.Sprintf("Operation completed successfully.\n\nResult:\n%s", string(jsonOutput))
	}

	// Fallback in case JSON marshaling fails
	return fmt.Sprintf("Operation completed successfully. Output: %v", output)
}