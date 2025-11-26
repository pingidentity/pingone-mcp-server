// Copyright © 2025 Ping Identity Corporation

package collections

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

type Collection interface {
	Name() string
	// RegisterTools registers the tools with the server.
	// The filter determines which tools are registered based on read-only mode, included/excluded tools.
	RegisterTools(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error
	ListTools() []types.ToolDefinition
}

type LegacySdkCollection interface {
	Name() string
	// RegisterTools registers the tools with the server.
	// The filter determines which tools are registered based on read-only mode, included/excluded tools.
	RegisterTools(ctx context.Context, server *mcp.Server, clientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error
	ListTools() []types.ToolDefinition
}
