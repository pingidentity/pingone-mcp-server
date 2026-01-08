// Copyright © 2025 Ping Identity Corporation

package capabilities

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func RegisterToolCollections(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter) error {
	// Get SDK collections
	defaultCollections := getDefaultCollections()

	for _, collection := range defaultCollections {
		if !toolFilter.ShouldIncludeCollection(collection.Name()) {
			logger.FromContext(ctx).Debug("MCP tool collection skipped", slog.String("collection", collection.Name()))
			continue
		}
		logger.FromContext(ctx).Debug("Registering MCP tool collection", slog.String("collection", collection.Name()))

		if err := collection.RegisterTools(ctx, server, clientFactory, tokenStore, toolFilter); err != nil {
			return err
		}
	}

	// Get legacy SDK collections
	legacyCollections := getLegacySdkCollections()

	for _, collection := range legacyCollections {
		if !toolFilter.ShouldIncludeCollection(collection.Name()) {
			logger.FromContext(ctx).Debug("MCP tool collection skipped", slog.String("collection", collection.Name()))
			continue
		}
		logger.FromContext(ctx).Debug("Registering MCP tool collection", slog.String("collection", collection.Name()))

		if err := collection.RegisterTools(ctx, server, legacySdkClientFactory, tokenStore, toolFilter); err != nil {
			return err
		}
	}
	return nil
}

func ListTools() []types.ToolDefinition {
	var tools []types.ToolDefinition
	defaultCollections := getDefaultCollections()
	for _, collection := range defaultCollections {
		tools = append(tools, collection.ListTools()...)
	}

	// List tools from legacy collections
	legacyCollections := getLegacySdkCollections()
	for _, collection := range legacyCollections {
		tools = append(tools, collection.ListTools()...)
	}
	return tools
}
