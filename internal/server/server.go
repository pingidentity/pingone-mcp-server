// Copyright © 2025 Ping Identity Corporation

package server

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
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

func Start(ctx context.Context, transport mcp.Transport, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pingone-mcp-server",
		Version: "v0.0.1",
	}, &mcp.ServerOptions{
		Logger:       logger.FromContext(ctx),
		HasPrompts:   false, // to flip on first prompt
		HasResources: false, // to flip on first resource
		HasTools:     true,
	})

	logger.FromContext(ctx).Debug("Registering MCP tool collections")
	err := capabilities.RegisterToolCollections(ctx, server, clientFactory, legacySdkClientFactory, authClientFactory, tokenStore, toolFilter, grantType)
	if err != nil {
		return err
	}

	logger.FromContext(ctx).Debug("Registering MCP resources")
	err = capabilities.RegisterResources(ctx, server, clientFactory, legacySdkClientFactory, authClientFactory, tokenStore, grantType)
	if err != nil {
		return err
	}

	logger.FromContext(ctx).Debug("Registering MCP prompts")
	err = capabilities.RegisterPrompts(ctx, server)
	if err != nil {
		return err
	}

	// Create and add environment validation middleware
	// This middleware validates that:
	// 1. Environment exists and is accessible
	// 2. Write operations are not performed on PRODUCTION environments
	logger.FromContext(ctx).Debug("Setting up environment validation middleware")
	allTools := capabilities.ListTools()
	toolRegistry := validation.NewToolRegistry(allTools)
	environmentsFactory := environments.NewPingOneClientEnvironmentsWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)
	validator := validation.NewCachingEnvironmentValidator(environmentsFactory, initializeAuthContext)
	validationMiddleware := validation.NewEnvironmentValidationMiddleware(validator, toolRegistry)
	server.AddReceivingMiddleware(validationMiddleware.Handler)
	logger.FromContext(ctx).Info("Environment validation middleware enabled - production environments are protected from write operations")

	log.Println("Starting PingOne MCP server...")

	if err := server.Run(ctx, transport); err != nil {
		return err
	}
	return nil

}
