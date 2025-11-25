// Copyright © 2025 Ping Identity Corporation

package initialize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/login"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

type ContextInitializer func(ctx context.Context) (context.Context, error)

func AuthContextInitializer(authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		authClient, err := authClientFactory.NewAuthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create auth client: %w", err)
		}
		return InitializeAuthContext(ctx, authClient, tokenStore, grantType)
	}
}

func InitializeAuthContext(ctx context.Context, authClient client.AuthClient, tokenStore tokenstore.TokenStore, grantType auth.GrantType) (context.Context, error) {
	var authSession *auth.AuthSession
	var err error
	// If grant type is authorization code, we can attempt to auto-login if no session exists
	if grantType == auth.GrantTypeAuthorizationCode {
		authSession, err = login.LoginIfNecessary(ctx, authClient, tokenStore, grantType)
		if err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	} else {
		hasSession, err := tokenStore.HasSession()
		if err != nil {
			return nil, fmt.Errorf("failed to check for auth session: %w", err)
		}
		if !hasSession {
			return nil, fmt.Errorf("no active auth session found. Use the login command to authenticate")
		}
		authSession, err = tokenStore.GetSession()
		if err != nil {
			return nil, fmt.Errorf("failed to get auth session: %w", err)
		}
	}
	ctx = audit.ContextWithSessionId(ctx, authSession.SessionId)
	return logger.ContextWithLogger(ctx, logger.FromContext(ctx).With(slog.String("sessionId", authSession.SessionId))), nil
}
