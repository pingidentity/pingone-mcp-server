// Copyright © 2025 Ping Identity Corporation

package initialize

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolInvocationMiddleware initializes context for all tool calls.
// It intercepts tool call requests and sets up logging and audit context before the tool handler executes.
//
// This middleware should be added to the MCP server via AddReceivingMiddleware.
// It performs the following initialization for each tool call:
// 1. Generates a unique transaction ID for tracing
// 2. Initializes the tool logger context with tool name and request details
// 3. Adds transaction ID to the context for audit tracking
// 4. Logs the tool invocation
type ToolInvocationMiddleware struct{}

// NewToolInvocationMiddleware creates middleware for tool invocation initialization.
func NewToolInvocationMiddleware() *ToolInvocationMiddleware {
	return &ToolInvocationMiddleware{}
}

// Handler implements the middleware pattern by returning a MethodHandler that wraps the next handler.
// This handler intercepts all MCP method calls and initializes context for tool calls.
func (m *ToolInvocationMiddleware) Handler(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Only initialize context for tool calls, not other MCP methods (initialize, list_tools, etc.)
		if method != "tools/call" {
			return next(ctx, method, req)
		}

		// Extract tool call details
		callToolReq, ok := req.(*mcp.CallToolRequest)
		if !ok {
			// Should never happen for tools/call method, but fail safe
			return nil, fmt.Errorf("tool invocation initialization failed: invalid tool call request")
		}

		toolName := callToolReq.Params.Name

		// Initialize tool invocation context
		initializedCtx := initializeToolInvocation(ctx, toolName, callToolReq)

		// Continue to next handler with initialized context
		return next(initializedCtx, method, req)
	}
}
