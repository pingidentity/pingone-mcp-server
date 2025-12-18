// Copyright © 2025 Ping Identity Corporation

package collections

import (
	"context"
	"fmt"

	legacypingone "github.com/patrickcping/pingone-go-sdk-v2/pingone"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

func InitializeAuthenticatedClient(clientFactory sdk.ClientFactory, tokenStore tokenstore.TokenStore) (*pingone.APIClient, error) {
	hasSession, err := tokenStore.HasSession()
	if err != nil {
		return nil, fmt.Errorf("failed to check for auth session: %w", err)
	}
	if !hasSession {
		return nil, fmt.Errorf("no active auth session found, unable to create authenticated client")
	}
	authSession, err := tokenStore.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session: %w", err)
	}
	client, err := clientFactory.NewClient(authSession.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create PingOne API client: %w", err)
	}
	return client, nil
}

func InitializeAuthenticatedLegacyClient(ctx context.Context, clientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore) (*legacypingone.Client, error) {
	hasSession, err := tokenStore.HasSession()
	if err != nil {
		return nil, fmt.Errorf("failed to check for auth session: %w", err)
	}
	if !hasSession {
		return nil, fmt.Errorf("no active auth session found, unable to create authenticated client")
	}
	authSession, err := tokenStore.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth session: %w", err)
	}
	client, err := clientFactory.NewClient(ctx, authSession.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create PingOne API client: %w", err)
	}
	return client, nil
}
