// Copyright © 2025 Ping Identity Corporation

package initialize

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/login"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

type ContextInitializer func(ctx context.Context) (context.Context, error)

func AuthContextInitializer(mcpServerSession *mcp.ServerSession, authClientFactory client.AuthClientFactory, tokenStore tokenstore.TokenStore, grantType auth.GrantType) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		authClient, err := authClientFactory.NewAuthClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create auth client: %w", err)
		}
		return InitializeAuthContext(ctx, mcpServerSession, authClient, tokenStore, grantType)
	}
}

func InitializeAuthContext(ctx context.Context, mcpServerSession *mcp.ServerSession, authClient client.AuthClient, tokenStore tokenstore.TokenStore, grantType auth.GrantType) (context.Context, error) {
	var authSession *auth.AuthSession
	var err error

	// If the browser login is not available, and the grant type is not device code, return an error
	if !authClient.BrowserLoginAvailable(grantType) && grantType != auth.GrantTypeDeviceCode {
		return nil, fmt.Errorf("browser login is not available in this environment and grant type %s cannot be used. Use %s grant type instead for headless auth", grantType, auth.GrantTypeDeviceCode.String())
	}

	authSession, err = login.LoginIfNecessary(ctx, authClient, tokenStore, grantType, mcpServerSession)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	ctx = audit.ContextWithSessionId(ctx, authSession.SessionId)
	return logger.ContextWithLogger(ctx, logger.FromContext(ctx).With(slog.String("sessionId", authSession.SessionId))), nil
}
