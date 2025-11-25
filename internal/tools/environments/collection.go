// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
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

	environmentsClientFactory := NewPingOneClientEnvironmentsWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)

	if toolFilter.ShouldIncludeTool(ListEnvironmentsDef.McpTool.Name, ListEnvironmentsDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", ListEnvironmentsDef.McpTool.Name))
		mcp.AddTool(server, ListEnvironmentsDef.McpTool, ListEnvironmentsHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(CreateEnvironmentDef.McpTool.Name, CreateEnvironmentDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", CreateEnvironmentDef.McpTool.Name))
		mcp.AddTool(server, CreateEnvironmentDef.McpTool, CreateEnvironmentHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(GetEnvironmentByIdDef.McpTool.Name, GetEnvironmentByIdDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetEnvironmentByIdDef.McpTool.Name))
		mcp.AddTool(server, GetEnvironmentByIdDef.McpTool, GetEnvironmentByIdHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(UpdateEnvironmentByIdDef.McpTool.Name, UpdateEnvironmentByIdDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdateEnvironmentByIdDef.McpTool.Name))
		mcp.AddTool(server, UpdateEnvironmentByIdDef.McpTool, UpdateEnvironmentByIdHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(GetEnvironmentServicesByIdDef.McpTool.Name, GetEnvironmentServicesByIdDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetEnvironmentServicesByIdDef.McpTool.Name))
		mcp.AddTool(server, GetEnvironmentServicesByIdDef.McpTool, GetEnvironmentServicesByIdHandler(environmentsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(UpdateEnvironmentServicesByIdDef.McpTool.Name, UpdateEnvironmentServicesByIdDef.IsReadOnly) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdateEnvironmentServicesByIdDef.McpTool.Name))
		mcp.AddTool(server, UpdateEnvironmentServicesByIdDef.McpTool, UpdateEnvironmentServicesByIdHandler(environmentsClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *EnvironmentsCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		ListEnvironmentsDef,
		CreateEnvironmentDef,
		GetEnvironmentByIdDef,
		UpdateEnvironmentByIdDef,
		GetEnvironmentServicesByIdDef,
		UpdateEnvironmentServicesByIdDef,
	}
}
