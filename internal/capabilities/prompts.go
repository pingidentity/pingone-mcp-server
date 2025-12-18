// Copyright © 2025 Ping Identity Corporation

package capabilities

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/prompts"
)

func RegisterPrompts(ctx context.Context, server *mcp.Server) error {
	return prompts.RegisterPrompts(ctx, server)
}
