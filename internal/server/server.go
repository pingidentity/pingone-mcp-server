// Copyright © 2025 Ping Identity Corporation

package server

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
)

func Start(ctx context.Context, transport mcp.Transport, clientFactory sdk.ClientFactory, legacySdkClientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pingone-mcp-server",
		Version: "v0.0.1",
	}, &mcp.ServerOptions{
		Logger: logger.FromContext(ctx),
	})

	logger.FromContext(ctx).Debug("Registering MCP tool collections")
	err := tools.RegisterCollections(ctx, server, clientFactory, legacySdkClientFactory, tokenStore, toolFilter)
	if err != nil {
		return err
	}

	log.Println("Starting PingOne MCP server...")

	if err := server.Run(ctx, transport); err != nil {
		return err
	}
	return nil

}
