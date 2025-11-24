// Copyright © 2025 Ping Identity Corporation

package initialize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

type ContextInitializer func(ctx context.Context) (context.Context, error)

func AuthContextInitializer(tokenStore tokenstore.TokenStore) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return InitializeAuthContext(ctx, tokenStore)
	}
}

func InitializeAuthContext(ctx context.Context, tokenStore tokenstore.TokenStore) (context.Context, error) {
	hasSession, err := tokenStore.HasSession()
	if err != nil {
		return nil, fmt.Errorf("failed to check for auth session: %w", err)
	}
	if !hasSession {
		return nil, fmt.Errorf("no active auth session found. Use the login command to authenticate")
	}
	authSession, err := tokenStore.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session: %w", err)
	}
	ctx = audit.ContextWithSessionId(ctx, authSession.SessionId)
	return logger.ContextWithLogger(ctx, logger.FromContext(ctx).With(slog.String("sessionId", authSession.SessionId))), nil
}
