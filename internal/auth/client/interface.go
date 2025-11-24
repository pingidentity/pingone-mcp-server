// Copyright © 2025 Ping Identity Corporation

package client

import (
	"context"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"golang.org/x/oauth2"
)

type AuthClient interface {
	TokenSource(ctx context.Context, grantType auth.GrantType) (oauth2.TokenSource, error)
}
