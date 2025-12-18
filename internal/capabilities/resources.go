// Copyright © 2025 Ping Identity Corporation

package capabilities

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/staticresources"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func RegisterResources(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) error {
	err := RegisterStaticResources(ctx, server)
	if err != nil {
		return err
	}

	err = RegisterDynamicResources(ctx, server, clientFactory, legacySdkClientFactory, authClientFactory, tokenStore, grantType)
	if err != nil {
		return err
	}

	return nil
}

func RegisterStaticResources(ctx context.Context, server *mcp.Server) error {
	return staticresources.RegisterStaticResources(ctx, server)
}

func RegisterDynamicResources(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) error {
	// Get SDK collections
	defaultCollections := getDefaultCollections()

	for _, collection := range defaultCollections {
		logger.FromContext(ctx).Debug("Registering MCP dynamic resources", slog.String("collection", collection.Name()))

		if err := collection.RegisterDynamicResources(ctx, server, clientFactory, authClientFactory, tokenStore, grantType); err != nil {
			return err
		}
	}

	// Get legacy SDK collections
	legacyCollections := getLegacySdkCollections()

	for _, collection := range legacyCollections {
		logger.FromContext(ctx).Debug("Registering MCP dynamic resources", slog.String("collection", collection.Name()))

		if err := collection.RegisterDynamicResources(ctx, server, legacySdkClientFactory, authClientFactory, tokenStore, grantType); err != nil {
			return err
		}
	}
	return nil
}
