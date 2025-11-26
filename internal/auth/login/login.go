// Copyright © 2025 Ping Identity Corporation

package login

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/logout"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

const authTimeout = 5 * time.Minute

func ForceLogin(ctx context.Context, authClient client.AuthClient, tokenStore tokenstore.TokenStore, grantType auth.GrantType) (*auth.AuthSession, error) {
	return login(ctx, authClient, tokenStore, grantType, true)
}

func LoginIfNecessary(ctx context.Context, authClient client.AuthClient, tokenStore tokenstore.TokenStore, grantType auth.GrantType) (*auth.AuthSession, error) {
	return login(ctx, authClient, tokenStore, grantType, false)
}

func login(ctx context.Context, authClient client.AuthClient, tokenStore tokenstore.TokenStore, grantType auth.GrantType, forceReAuth bool) (*auth.AuthSession, error) {
	hasSession, err := tokenStore.HasSession()
	if err != nil {
		return nil, err
	}
	if hasSession {
		activeSession, err := tokenStore.GetSession()
		if err != nil {
			return nil, err
		}
		if activeSession == nil {
			// Shouldn't happen
			return nil, errors.New("token store indicated session exists but returned nil session")
		}
		if !forceReAuth && activeSession.Expiry.After(time.Now()) { //TODO handle token refresh
			// Session is still valid
			logger.FromContext(ctx).Debug("An existing local auth session was found and is still valid", slog.String("sessionId", activeSession.SessionId), slog.String("expiry", activeSession.Expiry.Format(time.RFC3339)))
			return activeSession, nil
		}
		logger.FromContext(ctx).Info("An existing local auth session was found. Logging out before re-authenticating", slog.String("sessionId", activeSession.SessionId))
		err = logout.Logout(ctx, tokenStore)
		if err != nil {
			return nil, err
		}
	}

	logger.FromContext(ctx).Info("Initiating authentication flow")
	// Add a long timeout to the context for the auth process
	authCtx, cancel := context.WithTimeout(ctx, authTimeout)
	defer cancel()

	tokenSource, err := authClient.TokenSource(authCtx, grantType)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("authentication timed out after %v", authTimeout)
		}
		return nil, err
	}
	if tokenSource == nil {
		return nil, errors.New("authClient returned nil TokenSource")
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	if token == nil {
		// Should not happen
		return nil, errors.New("token source returned nil token")
	}
	logger.FromContext(ctx).Debug("Access token retrieved", slog.String("expiry", token.Expiry.Format(time.RFC3339)))

	sessionId := uuid.New().String()
	authSession := auth.NewAuthSession(*token, sessionId)
	if err := tokenStore.PutSession(authSession); err != nil {
		return nil, err
	}
	logger.FromContext(ctx).Debug("Auth session stored", slog.String("sessionId", authSession.SessionId), slog.String("expiry", authSession.Expiry.Format(time.RFC3339)))

	return &authSession, nil
}
