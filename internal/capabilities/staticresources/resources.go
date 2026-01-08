package staticresources

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
)

func RegisterStaticResources(_ context.Context, _ *mcp.Server) error {
	// No static resources to register
	return nil
}

func ListStaticResources() []types.StaticResourceDefinition {
	// No static resources defined
	return []types.StaticResourceDefinition{}
}
