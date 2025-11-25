// Copyright © 2025 Ping Identity Corporation

package logout_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
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
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	token := oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	session := auth.NewAuthSession(token, "test-session-id")
	err := tokenStore.PutSession(session)
	require.NoError(t, err)

	err = testutils.ExecuteCliLogoutCommand(t, ctx, tokenStoreFactory)
	require.NoError(t, err, "Logout should succeed with valid session")

	// Verify session was deleted
	hasSession, err := tokenStore.HasSession()
	require.NoError(t, err, "HasSession should not error")
	assert.False(t, hasSession, "Session should be deleted after logout")
	afterLogoutSession, err := tokenStore.GetSession()
	assert.Nil(t, afterLogoutSession, "Session should be nil after logout")
	assert.Error(t, err, "Session should be deleted after logout")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLogoutCommand_Direct_NoExistingSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	err := testutils.ExecuteCliLogoutCommand(t, ctx, tokenStoreFactory)
	require.NoError(t, err, "Logout should succeed when no session exists")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLogoutCommand_Direct_TokenStoreFactoryError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedError := assert.AnError
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithError(expectedError)

	err := testutils.ExecuteCliLogoutCommand(t, ctx, tokenStoreFactory)
	require.Error(t, err, "Logout should fail when token store factory returns error")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the factory error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLogoutCommand_Direct_StoreTypeSelection(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedStoreType tokenstore.StoreType
		description       string
	}{
		{
			name:              "default store type is keychain",
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Logout should use keychain store type by default",
		},
		{
			name:              "explicit keychain store type",
			args:              []string{"--store-type", "keychain"},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Logout should use keychain store type when explicitly specified",
		},
		{
			name:              "file store type",
			args:              []string{"--store-type", "file"},
			expectedStoreType: tokenstore.StoreTypeFile,
			description:       "Logout should use file store type when specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStore()
			tokenStoreFactory := testutils.NewMockTokenStoreFactory()
			tokenStoreFactory.On("NewTokenStore", tt.expectedStoreType).Return(tokenStore, nil)

			err := testutils.ExecuteCliLogoutCommand(t, ctx, tokenStoreFactory, tt.args...)
			require.NoError(t, err, tt.description)
			// Verify token store was created with expected store type
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestLogoutCommand_Direct_InvalidStoreType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStoreFactory := testutils.NewMockTokenStoreFactory()

	err := testutils.ExecuteCliLogoutCommand(t, ctx, tokenStoreFactory, "--store-type", "invalid")
	require.Error(t, err, "Logout should fail with invalid store type")
	assert.Contains(t, err.Error(), "unable to parse store type from string: invalid", "Error should indicate invalid store type")
}
