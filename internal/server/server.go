// Copyright © 2025 Ping Identity Corporation

package server

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
)

func Start(ctx context.Context, transport mcp.Transport, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pingone-mcp-server",
		Version: "v0.0.1",
	}, nil)

	logger.FromContext(ctx).Debug("Registering MCP tool collections")
	err := tools.RegisterCollections(ctx, server, clientFactory, legacySdkClientFactory, authClientFactory, tokenStore, toolFilter, grantType)
	if err != nil {
		return err
	}

	log.Println("Starting PingOne MCP server...")

	if err := server.Run(ctx, transport); err != nil {
		return err
	}
	return nil

}
