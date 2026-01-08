// Copyright © 2025 Ping Identity Corporation

package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

// getDefaultCollections creates SDK collections
func getDefaultCollections() []collections.Collection {
	return []collections.Collection{
		&directory.DirectoryCollection{},
		&environments.EnvironmentsCollection{},
	}
}

// getLegacySdkCollections creates legacy SDK collections
func getLegacySdkCollections() []collections.LegacySdkCollection {
	return []collections.LegacySdkCollection{
		&applications.ApplicationsCollection{},
		&populations.PopulationsCollection{},
	}
}

func RegisterCollections(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter) error {
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
