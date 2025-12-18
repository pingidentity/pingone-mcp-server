// Copyright © 2025 Ping Identity Corporation

package initialize

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

func InitializeToolInvocation(ctx context.Context, name string, req *mcp.CallToolRequest) context.Context {
	transactionId := audit.GenerateTransactionId()
	ctx = logger.InitToolLoggerContext(ctx, name, req, transactionId)
	ctx = audit.ContextWithTransactionId(ctx, transactionId)
	logger.FromContext(ctx).Debug("Invoked MCP tool")
	return ctx
}
