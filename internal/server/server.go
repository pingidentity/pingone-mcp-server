// Copyright © 2025 Ping Identity Corporation

package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	authmiddleware "github.com/pingidentity/pingone-mcp-server/internal/auth/middleware"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/validation"
)

func Start(ctx context.Context, version string, transport mcp.Transport, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pingone-mcp-server",
		Title:   "PingOne MCP Server",
		Version: version,
	}, &mcp.ServerOptions{
		Logger: logger.FromContext(ctx),
	})

	logger.FromContext(ctx).Debug("Registering MCP tool collections")
	err := tools.RegisterCollections(ctx, server, clientFactory, legacySdkClientFactory, tokenStore, toolFilter)
	if err != nil {
		return err
	}

	// Setup middleware
	invocationMiddleware := setupInvocationMiddleware(ctx, server)
	authMiddleware := setupAuthMiddleware(ctx, server, authClientFactory, tokenStore, grantType)
	validationMiddleware := setupValidationMiddleware(ctx, server, clientFactory, tokenStore)

	// Register middleware in order: invocation -> auth -> validation
	// Order matters: invocation sets up logging/audit, auth establishes session, validation checks permissions using the auth context
	server.AddReceivingMiddleware(invocationMiddleware, authMiddleware, validationMiddleware)
	logger.FromContext(ctx).Info("Middleware enabled - all tool calls will be authenticated and validated")

	logger.FromContext(ctx).Info("Starting PingOne MCP server...")

	if err := server.Run(ctx, transport); err != nil {
		return err
	}
	return nil

}

func setupInvocationMiddleware(ctx context.Context, server *mcp.Server) mcp.Middleware {
	invocationMiddleware := initialize.NewToolInvocationMiddleware()
	return invocationMiddleware.Handler
}

func setupAuthMiddleware(ctx context.Context, server *mcp.Server, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) mcp.Middleware {
	authMiddleware := authmiddleware.NewAuthMiddleware(authClientFactory, tokenStore, grantType)
	return authMiddleware.Handler
}

func setupValidationMiddleware(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, tokenStore tokenstore.TokenStore) mcp.Middleware {
	allTools := tools.ListTools()
	toolRegistry := validation.NewToolRegistry(allTools)
	environmentsFactory := environments.NewPingOneClientEnvironmentsWrapperFactory(clientFactory, tokenStore)

	validator := validation.NewCachingEnvironmentValidator(environmentsFactory)
	validationMiddleware := validation.NewEnvironmentValidationMiddleware(validator, toolRegistry)
	return validationMiddleware.Handler
}
