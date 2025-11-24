// Copyright © 2025 Ping Identity Corporation

package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSessionCommand_FromRoot_Basic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStore()

	// Create a session to display
	token := oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	session := auth.NewAuthSession(token, "test-session-id")
	err := tokenStore.PutSession(session)
	require.NoError(t, err)

	err = testutils.ExecuteCliSessionCommand(t, ctx, tokenStore)
	require.NoError(t, err, "Session command should succeed with valid session")
}

func TestSessionCommand_Direct_Success(t *testing.T) {
	tests := []struct {
		name        string
		description string
		sessionFn   func() *auth.AuthSession
	}{
		{
			name:        "session with active token",
			description: "Session command should succeed with active token",
			sessionFn: func() *auth.AuthSession {
				token := oauth2.Token{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					Expiry:       time.Now().Add(time.Hour),
				}
				return testutils.Pointer(auth.NewAuthSession(token, "active-session-id"))
			},
		},
		{
			name:        "session with expired token",
			description: "Session command should succeed with expired token and show expired status",
			sessionFn: func() *auth.AuthSession {
				token := oauth2.Token{
					AccessToken:  "expired-access-token",
					RefreshToken: "expired-refresh-token",
					Expiry:       time.Now().Add(-time.Hour),
				}
				return testutils.Pointer(auth.NewAuthSession(token, "expired-session-id"))
			},
		},
		{
			name:        "nil session",
			description: "Session command should succeed when no session exists",
			sessionFn: func() *auth.AuthSession {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStore()
			session := tt.sessionFn()
			if session != nil {
				err := tokenStore.PutSession(*session)
				require.NoError(t, err)
			}

			err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStore)
			require.NoError(t, err, tt.description)
		})
	}
}

func TestSessionCommand_Direct_HasSessionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a mock token store that returns errors
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStore.HasSessionError = assert.AnError

	err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStore)

	require.Error(t, err, "Session command should fail when HasSession returns error")
	assert.True(t, errors.Is(err, assert.AnError), "Error should contain the mock error")
}

func TestSessionCommand_Direct_GetSessionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a mock token store that returns errors
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStore.GetSessionError = assert.AnError
	err := tokenStore.PutSession(auth.AuthSession{}) // HasSession will return true
	require.NoError(t, err)

	err = testutils.ExecuteCliSessionCommand(t, ctx, tokenStore)

	require.Error(t, err, "Session command should fail when GetSession returns error")
	assert.True(t, errors.Is(err, assert.AnError), "Error should contain the mock error")
}

func TestSessionCommand_Direct_NilTokenStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := testutils.ExecuteCliSessionCommand(t, ctx, nil)

	require.Error(t, err, "Session command should fail with nil token store")
	assert.Contains(t, err.Error(), "tokenStore is nil", "Error should indicate nil token store")
}
