// Copyright © 2025 Ping Identity Corporation

package populations

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
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

const CollectionName = "populations"

var _ collections.LegacySdkCollection = &PopulationsCollection{}

type PopulationsCollection struct{}

func (c *PopulationsCollection) Name() string {
	return CollectionName
}

func (c *PopulationsCollection) RegisterTools(ctx context.Context, server *mcp.Server, clientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	if clientFactory == nil {
		return fmt.Errorf("PingOne API client factory is nil")
	}
	if tokenStore == nil {
		return fmt.Errorf("token store is nil")
	}
	if authClientFactory == nil {
		return fmt.Errorf("auth client factory is nil")
	}

	populationsClientFactory := NewPingOneClientPopulationsWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)

	if toolFilter.ShouldIncludeTool(&ListPopulationsDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", ListPopulationsDef.McpTool.Name))
		mcp.AddTool(server, ListPopulationsDef.McpTool, ListPopulationsHandler(populationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&CreatePopulationDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", CreatePopulationDef.McpTool.Name))
		mcp.AddTool(server, CreatePopulationDef.McpTool, CreatePopulationHandler(populationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&GetPopulationDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetPopulationDef.McpTool.Name))
		mcp.AddTool(server, GetPopulationDef.McpTool, GetPopulationHandler(populationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&UpdatePopulationDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdatePopulationDef.McpTool.Name))
		mcp.AddTool(server, UpdatePopulationDef.McpTool, UpdatePopulationHandler(populationsClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *PopulationsCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		ListPopulationsDef,
		CreatePopulationDef,
		GetPopulationDef,
		UpdatePopulationDef,
	}
}

func (c *PopulationsCollection) RegisterDynamicResources(_ context.Context, _ *mcp.Server, _ legacy.ClientFactory, _ client.AuthClientFactory, _ tokenstore.TokenStore, _ auth.GrantType) error {
	// No dynamic resources to register
	return nil
}

func (c *PopulationsCollection) ListDynamicResources() []types.DynamicResourceDefinition {
	// No dynamic resources defined
	return []types.DynamicResourceDefinition{}
}
