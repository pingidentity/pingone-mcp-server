// Copyright © 2025 Ping Identity Corporation

package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
)

// AuthMiddleware ensures all tool calls have proper authentication context.
// It intercepts tool call requests and initializes the auth context before the tool handler executes.
//
// This middleware should be added to the MCP server via AddReceivingMiddleware.
// It runs the initializeAuthContext function to establish authentication, which may:
// 1. Check for an existing session
// 2. Trigger browser-based login if necessary
// 3. Add session information to the context
type AuthMiddleware struct {
	authClientFactory client.AuthClientFactory
	tokenStore        tokenstore.TokenStore
	grantType         auth.GrantType
}

// NewAuthMiddleware creates middleware with auth dependencies.
// The authClientFactory is used to create auth clients for login flows.
// The tokenStore manages session persistence.
// The grantType determines the authentication method (authorization_code or device_code).
func NewAuthMiddleware(
	authClientFactory client.AuthClientFactory,
	tokenStore tokenstore.TokenStore,
	grantType auth.GrantType,
) *AuthMiddleware {
	return &AuthMiddleware{
		authClientFactory: authClientFactory,
		tokenStore:        tokenStore,
		grantType:         grantType,
	}
}

// Handler implements the middleware pattern by returning a MethodHandler that wraps the next handler.
// This handler intercepts all MCP method calls and ensures tool calls have proper authentication context.
func (m *AuthMiddleware) Handler(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Only authenticate tool calls, not other MCP methods (initialize, list_tools, etc.)
		if method != "tools/call" {
			return next(ctx, method, req)
		}

		// Extract tool call details for logging
		callToolReq, ok := req.(*mcp.CallToolRequest)
		if !ok {
			// Should never happen for tools/call method, but fail safe
			return nil, fmt.Errorf("authentication failed: invalid tool call request")
		}

		toolName := callToolReq.Params.Name

		logger.FromContext(ctx).Debug("Initializing authentication for tool",
			slog.String("tool", toolName))

		// Initialize auth context using the same logic as individual tool handlers
		initializeAuthContext := initialize.AuthContextInitializer(callToolReq.Session, m.authClientFactory, m.tokenStore, m.grantType)
		authenticatedCtx, err := initializeAuthContext(ctx)
		if err != nil {
			logger.FromContext(ctx).Error("Authentication initialization failed",
				slog.String("tool", toolName),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("authentication failed: %w", err)
		}

		logger.FromContext(ctx).Debug("Authentication initialized successfully",
			slog.String("tool", toolName))

		// Authentication successful, continue to next handler with authenticated context
		return next(authenticatedCtx, method, req)
	}
}
