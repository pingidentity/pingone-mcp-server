// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
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

	if toolFilter.ShouldIncludeTool(&GetPopulationByIdDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetPopulationByIdDef.McpTool.Name))
		mcp.AddTool(server, GetPopulationByIdDef.McpTool, GetPopulationByIdHandler(populationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&UpdatePopulationByIdDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdatePopulationByIdDef.McpTool.Name))
		mcp.AddTool(server, UpdatePopulationByIdDef.McpTool, UpdatePopulationByIdHandler(populationsClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *PopulationsCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		ListPopulationsDef,
		CreatePopulationDef,
		GetPopulationByIdDef,
		UpdatePopulationByIdDef,
	}
}
