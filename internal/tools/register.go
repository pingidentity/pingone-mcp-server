// Copyright © 2025 Ping Identity Corporation

package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/directory"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var defaultCollections = []collections.Collection{
	&directory.DirectoryCollection{},
	&environments.EnvironmentsCollection{},
}
var defaultLegacySdkCollections = []collections.LegacySdkCollection{
	&applications.ApplicationsCollection{},
	&populations.PopulationsCollection{},
}

func RegisterCollections(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	for _, collection := range defaultCollections {
		if !toolFilter.ShouldIncludeCollection(collection.Name()) {
			logger.FromContext(ctx).Debug("MCP tool collection skipped", slog.String("collection", collection.Name()))
			continue
		}
		logger.FromContext(ctx).Debug("Registering MCP tool collection", slog.String("collection", collection.Name()))

		if err := collection.RegisterTools(ctx, server, clientFactory, authClientFactory, tokenStore, toolFilter, grantType); err != nil {
			return err
		}
	}
	for _, collection := range defaultLegacySdkCollections {
		if !toolFilter.ShouldIncludeCollection(collection.Name()) {
			logger.FromContext(ctx).Debug("MCP tool collection skipped", slog.String("collection", collection.Name()))
			continue
		}
		logger.FromContext(ctx).Debug("Registering MCP tool collection", slog.String("collection", collection.Name()))

		if err := collection.RegisterTools(ctx, server, legacySdkClientFactory, authClientFactory, tokenStore, toolFilter, grantType); err != nil {
			return err
		}
	}
	return nil
}

func ListTools() []types.ToolDefinition {
	var tools []types.ToolDefinition
	for _, collection := range defaultCollections {
		tools = append(tools, collection.ListTools()...)
	}
	for _, collection := range defaultLegacySdkCollections {
		tools = append(tools, collection.ListTools()...)
	}
	return tools
}
