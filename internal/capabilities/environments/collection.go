// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

const CollectionName = "environments"

var _ collections.Collection = &EnvironmentsCollection{}

type EnvironmentsCollection struct{}

func (c *EnvironmentsCollection) Name() string {
	return CollectionName
}

func (c *EnvironmentsCollection) RegisterTools(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	if clientFactory == nil {
		return fmt.Errorf("PingOne API client factory is nil")
	}
	if tokenStore == nil {
		return fmt.Errorf("token store is nil")
	}
	if authClientFactory == nil {
		return fmt.Errorf("auth client factory is nil")
	}

	environmentsClientFactory := NewPingOneClientEnvironmentsWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)

	if toolFilter.ShouldIncludeTool(&ListEnvironmentsDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", ListEnvironmentsDef.McpTool.Name))
		mcp.AddTool(server, ListEnvironmentsDef.McpTool, ListEnvironmentsHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&CreateEnvironmentDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", CreateEnvironmentDef.McpTool.Name))
		mcp.AddTool(server, CreateEnvironmentDef.McpTool, CreateEnvironmentHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&GetEnvironmentDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetEnvironmentDef.McpTool.Name))
		mcp.AddTool(server, GetEnvironmentDef.McpTool, GetEnvironmentHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&UpdateEnvironmentDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdateEnvironmentDef.McpTool.Name))
		mcp.AddTool(server, UpdateEnvironmentDef.McpTool, UpdateEnvironmentHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&GetEnvironmentServicesDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetEnvironmentServicesDef.McpTool.Name))
		mcp.AddTool(server, GetEnvironmentServicesDef.McpTool, GetEnvironmentServicesHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&UpdateEnvironmentServicesDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdateEnvironmentServicesDef.McpTool.Name))
		mcp.AddTool(server, UpdateEnvironmentServicesDef.McpTool, UpdateEnvironmentServicesHandler(environmentsClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *EnvironmentsCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		ListEnvironmentsDef,
		CreateEnvironmentDef,
		GetEnvironmentDef,
		UpdateEnvironmentDef,
		GetEnvironmentServicesDef,
		UpdateEnvironmentServicesDef,
	}
}

func (c *EnvironmentsCollection) RegisterDynamicResources(_ context.Context, _ *mcp.Server, _ sdk.ClientFactory, _ client.AuthClientFactory, _ tokenstore.TokenStore, _ auth.GrantType) error {
	// No dynamic resources to register
	return nil
}

func (c *EnvironmentsCollection) ListDynamicResources() []types.DynamicResourceDefinition {
	// No dynamic resources defined
	return []types.DynamicResourceDefinition{}
}
