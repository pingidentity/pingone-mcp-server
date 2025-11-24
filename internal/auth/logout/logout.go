// Copyright © 2025 Ping Identity Corporation

package logout

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func Logout(ctx context.Context, tokenStore tokenstore.TokenStore) error {
	if tokenStore == nil {
		return errors.New("provided tokenStore is nil")
	}

	sessionExists, err := tokenStore.HasSession()
	if err != nil {
		return err
	}
	if !sessionExists {
		logger.FromContext(ctx).Info("No existing login session found.")
		return nil
	}

	authSession, err := tokenStore.GetSession()
	if err != nil {
		return err
	}
	if authSession == nil {
		// Should not happen as we checked HasSession above
		return errors.New("token store indicated session exists but returned nil session")
	}
	ctx = logger.ContextWithLogger(ctx, logger.FromContext(ctx).With(slog.String("sessionId", authSession.SessionId)))
	logger.FromContext(ctx).Debug("Local auth session retrieved")

	if err := tokenStore.DeleteSession(); err != nil {
		return err
	}
	logger.FromContext(ctx).Info("Local auth session deleted")

	return nil
}
