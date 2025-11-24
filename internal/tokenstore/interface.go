// Copyright © 2025 Ping Identity Corporation

package tokenstore

import (
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
)

type TokenStore interface {
	PutSession(session auth.AuthSession) error
	GetSession() (*auth.AuthSession, error)
	HasSession() (bool, error)
	DeleteSession() error
}
