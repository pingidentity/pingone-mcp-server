// Copyright © 2025 Ping Identity Corporation

package logout_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestLogoutCommand_FromRoot_Basic(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "logout help flag",
			args:        []string{"logout", "--help"},
			expectError: false,
			description: "Logout command help should execute without error",
		},
		{
			name:          "logout invalid flag",
			args:          []string{"logout", "--invalid-flag"},
			expectError:   true,
			errorContains: "unknown flag",
			description:   "Logout command should return error for invalid flag",
		},
		// Direct test would require human interaction
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := testutils.ExecuteCliRootCommand(t, ctx, tt.args...)

			if tt.expectError {
				require.Error(t, err, tt.description)
				if tt.errorContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errorContains),
						"Error should contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				require.NoError(t, err, tt.description)
			}
		})
	}
}

func TestLogoutCommand_Direct_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStore()

	token := oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	session := auth.NewAuthSession(token, "test-session-id")
	err := tokenStore.PutSession(session)
	require.NoError(t, err)

	err = testutils.ExecuteCliLogoutCommand(t, ctx, tokenStore)
	require.NoError(t, err, "Logout should succeed with valid session")

	// Verify session was deleted
	hasSession, err := tokenStore.HasSession()
	require.NoError(t, err, "HasSession should not error")
	assert.False(t, hasSession, "Session should be deleted after logout")
	afterLogoutSession, err := tokenStore.GetSession()
	assert.Nil(t, afterLogoutSession, "Session should be nil after logout")
	assert.Error(t, err, "Session should be deleted after logout")
}

func TestLogoutCommand_Direct_NoExistingSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStore()

	err := testutils.ExecuteCliLogoutCommand(t, ctx, tokenStore)
	require.NoError(t, err, "Logout should succeed when no session exists")
}
