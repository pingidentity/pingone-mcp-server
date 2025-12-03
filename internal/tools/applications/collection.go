// Copyright © 2025 Ping Identity Corporation

package applications

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

const CollectionName = "applications"

var _ collections.LegacySdkCollection = &ApplicationsCollection{}

type ApplicationsCollection struct{}

func (c *ApplicationsCollection) Name() string {
	return CollectionName
}

func (c *ApplicationsCollection) RegisterTools(ctx context.Context, server *mcp.Server, clientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	if clientFactory == nil {
		return fmt.Errorf("PingOne API client factory is nil")
	}
	if tokenStore == nil {
		return fmt.Errorf("token store is nil")
	}
	if authClientFactory == nil {
		return fmt.Errorf("auth client factory is nil")
	}

	applicationsClientFactory := NewPingOneClientApplicationsWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)

	if toolFilter.ShouldIncludeTool(&ListApplicationsDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", ListApplicationsDef.McpTool.Name))
		mcp.AddTool(server, ListApplicationsDef.McpTool, ListApplicationsHandler(applicationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&GetApplicationByIdDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetApplicationByIdDef.McpTool.Name))
		mcp.AddTool(server, GetApplicationByIdDef.McpTool, GetApplicationByIdHandler(applicationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&CreateApplicationDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", CreateApplicationDef.McpTool.Name))
		mcp.AddTool(server, CreateApplicationDef.McpTool, CreateApplicationHandler(applicationsClientFactory, initializeAuthContext))
	}

	if toolFilter.ShouldIncludeTool(&UpdateApplicationByIdDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", UpdateApplicationByIdDef.McpTool.Name))
		mcp.AddTool(server, UpdateApplicationByIdDef.McpTool, UpdateApplicationByIdHandler(applicationsClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *ApplicationsCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		ListApplicationsDef,
		GetApplicationByIdDef,
		CreateApplicationDef,
		UpdateApplicationByIdDef,
	}
}
