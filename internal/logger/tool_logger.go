// Copyright © 2025 Ping Identity Corporation

package logger

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func InitToolLoggerContext(ctx context.Context, toolName string, req *mcp.CallToolRequest, transactionId string) context.Context {
	attrs := []any{ // The Logger.With method expects []any
		slog.String("tool", toolName),
		slog.String("transactionId", transactionId),
	}
	if req != nil && req.Session != nil && req.Session.ID() != "" {
		attrs = append(attrs, slog.String("mcpSessionId", req.GetSession().ID()))
	}
	return ContextWithLogger(ctx, FromContext(ctx).With(attrs...))
}
