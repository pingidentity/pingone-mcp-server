// Copyright © 2025 Ping Identity Corporation

package collections

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

type Collection interface {
	// Name returns the unique identifier for this collection.
	// The returned string is used to identify and organize related tools and resources.
	Name() string

	// RegisterDynamicResources registers dynamic resources with the MCP server.
	// Dynamic resources are generated at runtime based on the current PingOne environment state.
	// The filter determines which resources are registered based on configuration settings.
	// This function requires valid authentication credentials through the clientFactory parameter.
	RegisterDynamicResources(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) error

	// RegisterTools registers the tools with the server.
	// The filter determines which tools are registered based on read-only mode, included/excluded tools.
	// This function requires valid authentication credentials through the clientFactory parameter.
	RegisterTools(ctx context.Context, server *mcp.Server, clientFactory sdk.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error

	// ListDynamicResources returns a slice of all dynamic resource definitions available in this collection.
	// Dynamic resources are generated at runtime based on the current PingOne environment state.
	ListDynamicResources() []types.DynamicResourceDefinition

	// ListTools returns a slice of all tool definitions available in this collection.
	// Tool definitions describe the capabilities and parameters of each tool in the collection.
	ListTools() []types.ToolDefinition
}

type LegacySdkCollection interface {
	// Name returns the unique identifier for this collection.
	// The returned string is used to identify and organize related tools and resources.
	Name() string

	// RegisterDynamicResources registers dynamic resources with the MCP server using the legacy SDK.
	// Dynamic resources are generated at runtime based on the current PingOne environment state.
	// The filter determines which resources are registered based on configuration settings.
	// This function requires valid authentication credentials through the clientFactory parameter.
	RegisterDynamicResources(ctx context.Context, server *mcp.Server, clientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) error

	// RegisterTools registers the tools with the server using the legacy SDK.
	// The filter determines which tools are registered based on read-only mode, included/excluded tools.
	// This function requires valid authentication credentials through the clientFactory parameter.
	RegisterTools(ctx context.Context, server *mcp.Server, clientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, toolFilter *filter.Filter, grantType auth.GrantType) error

	// ListDynamicResources returns a slice of all dynamic resource definitions available in this collection.
	// Dynamic resources are generated at runtime based on the current PingOne environment state.
	ListDynamicResources() []types.DynamicResourceDefinition

	// ListTools returns a slice of all tool definitions available in this collection.
	// Tool definitions describe the capabilities and parameters of each tool in the collection.
	ListTools() []types.ToolDefinition
}
