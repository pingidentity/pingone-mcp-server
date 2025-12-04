// Copyright © 2025 Ping Identity Corporation

package directory

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

const CollectionName = "directory"

var _ collections.Collection = &DirectoryCollection{}

type DirectoryCollection struct{}

func (c *DirectoryCollection) Name() string {
	return CollectionName
}

func (c *DirectoryCollection) RegisterTools(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	if clientFactory == nil {
		return fmt.Errorf("PingOne API client factory is nil")
	}
	if tokenStore == nil {
		return fmt.Errorf("token store is nil")
	}
	if authClientFactory == nil {
		return fmt.Errorf("auth client factory is nil")
	}

	directoryClientFactory := NewPingOneClientDirectoryWrapperFactory(clientFactory, tokenStore)
	initializeAuthContext := initialize.AuthContextInitializer(authClientFactory, tokenStore, grantType)

	if toolFilter.ShouldIncludeTool(&GetTotalIdentitiesByEnvironmentIdDef) {
		logger.FromContext(ctx).Debug("Registering MCP tool", slog.String("collection", c.Name()), slog.String("tool", GetTotalIdentitiesByEnvironmentIdDef.McpTool.Name))
		mcp.AddTool(server, GetTotalIdentitiesByEnvironmentIdDef.McpTool, GetTotalIdentitiesByEnvironmentIdHandler(directoryClientFactory, initializeAuthContext))
	}

	return nil
}

func (c *DirectoryCollection) ListTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		GetTotalIdentitiesByEnvironmentIdDef,
	}
}
