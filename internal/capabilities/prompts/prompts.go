package prompts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
)

func RegisterPrompts(_ context.Context, _ *mcp.Server) error {
	// No prompts to register
	return nil
}

func ListPrompts() []types.PromptDefinition {
	// No prompts defined
	return []types.PromptDefinition{}
}
