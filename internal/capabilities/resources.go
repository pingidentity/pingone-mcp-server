// Copyright © 2025 Ping Identity Corporation

package capabilities

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/staticresources"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func RegisterResources(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore) error {
	err := RegisterStaticResources(ctx, server)
	if err != nil {
		return err
	}

	err = RegisterDynamicResources(ctx, server, clientFactory, legacySdkClientFactory, tokenStore)
	if err != nil {
		return err
	}

	return nil
}

func RegisterStaticResources(ctx context.Context, server *mcp.Server) error {
	return staticresources.RegisterStaticResources(ctx, server)
}

func ListStaticResources() []types.StaticResourceDefinition {
	return staticresources.ListStaticResources()
}

func RegisterDynamicResources(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore) error {
	// Get SDK collections
	defaultCollections := getDefaultCollections()

	for _, collection := range defaultCollections {
		logger.FromContext(ctx).Debug("Registering MCP dynamic resources", slog.String("collection", collection.Name()))

		if err := collection.RegisterDynamicResources(ctx, server, clientFactory, tokenStore); err != nil {
			return err
		}
	}

	// Get legacy SDK collections
	legacyCollections := getLegacySdkCollections()

	for _, collection := range legacyCollections {
		logger.FromContext(ctx).Debug("Registering MCP dynamic resources", slog.String("collection", collection.Name()))

		if err := collection.RegisterDynamicResources(ctx, server, legacySdkClientFactory, tokenStore); err != nil {
			return err
		}
	}
	return nil
}

func ListDynamicResources() []types.DynamicResourceDefinition {
	var tools []types.DynamicResourceDefinition
	defaultCollections := getDefaultCollections()
	for _, collection := range defaultCollections {
		tools = append(tools, collection.ListDynamicResources()...)
	}

	// List tools from legacy collections
	legacyCollections := getLegacySdkCollections()
	for _, collection := range legacyCollections {
		tools = append(tools, collection.ListDynamicResources()...)
	}
	return tools
}
