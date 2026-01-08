// Copyright © 2025 Ping Identity Corporation

package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	authmiddleware "github.com/pingidentity/pingone-mcp-server/internal/auth/middleware"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/validation"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func Start(ctx context.Context, version string, transport mcp.Transport, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pingone-mcp-server",
		Title:   "PingOne MCP Server",
		Version: version,
	}, &mcp.ServerOptions{
		Logger: logger.FromContext(ctx),
		Capabilities: &mcp.ServerCapabilities{
			Completions: nil,
			Logging:     &mcp.LoggingCapabilities{},
			Prompts: &mcp.PromptCapabilities{
				ListChanged: true,
			},
			Resources: &mcp.ResourceCapabilities{
				ListChanged: true,
			},
			Tools: &mcp.ToolCapabilities{
				ListChanged: true,
			},
		},
	})

	logger.FromContext(ctx).Debug("Registering MCP tool collections")
	err := capabilities.RegisterToolCollections(ctx, server, clientFactory, legacySdkClientFactory, tokenStore, toolFilter)
	if err != nil {
		return err
	}

	logger.FromContext(ctx).Debug("Registering MCP resources")
	err = capabilities.RegisterResources(ctx, server, clientFactory, legacySdkClientFactory, tokenStore)
	if err != nil {
		return err
	}

	logger.FromContext(ctx).Debug("Registering MCP prompts")
	err = capabilities.RegisterPrompts(ctx, server)
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
	allTools := capabilities.ListTools()
	toolRegistry := validation.NewToolRegistry(allTools)
	environmentsFactory := environments.NewPingOneClientEnvironmentsWrapperFactory(clientFactory, tokenStore)

	validator := validation.NewCachingEnvironmentValidator(environmentsFactory)
	validationMiddleware := validation.NewEnvironmentValidationMiddleware(validator, toolRegistry)
	return validationMiddleware.Handler
}
