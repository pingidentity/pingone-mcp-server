// Copyright © 2025 Ping Identity Corporation

package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSessionCommand_FromRoot_Basic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	// Create a session to display
	token := oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}
	session := auth.NewAuthSession(token, "test-session-id")
	err := tokenStore.PutSession(session)
	require.NoError(t, err)

	err = testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory)
	require.NoError(t, err, "Session command should succeed with valid session")
	tokenStoreFactory.AssertExpectations(t)
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
			tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)
			session := tt.sessionFn()
			if session != nil {
				err := tokenStore.PutSession(*session)
				require.NoError(t, err)
			}

			err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory)
			require.NoError(t, err, tt.description)
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestSessionCommand_Direct_HasSessionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a mock token store that returns errors
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)
	tokenStore.HasSessionError = assert.AnError

	err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory)

	require.Error(t, err, "Session command should fail when HasSession returns error")
	assert.True(t, errors.Is(err, assert.AnError), "Error should contain the mock error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestSessionCommand_Direct_GetSessionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a mock token store that returns errors
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)
	tokenStore.GetSessionError = assert.AnError
	err := tokenStore.PutSession(auth.AuthSession{}) // HasSession will return true
	require.NoError(t, err)

	err = testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory)

	require.Error(t, err, "Session command should fail when GetSession returns error")
	assert.True(t, errors.Is(err, assert.AnError), "Error should contain the mock error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestSessionCommand_Direct_NilTokenStoreFactory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := testutils.ExecuteCliSessionCommand(t, ctx, nil)

	require.Error(t, err, "Session command should fail with nil token store factory")
	assert.Contains(t, err.Error(), "tokenStoreFactory is nil", "Error should indicate nil token store factory")
}

func TestSessionCommand_Direct_TokenStoreFactoryError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedError := assert.AnError
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithError(expectedError)

	err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory)
	require.Error(t, err, "Session command should fail when token store factory returns error")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the factory error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestSessionCommand_Direct_StoreTypeSelection(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		expectedStoreType tokenstore.StoreType
		description       string
	}{
		{
			name:              "default store type is keychain",
			args:              []string{},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Session should use keychain store type by default",
		},
		{
			name:              "explicit keychain store type",
			args:              []string{"--store-type", "keychain"},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Session should use keychain store type when explicitly specified",
		},
		{
			name:              "file store type",
			args:              []string{"--store-type", "file"},
			expectedStoreType: tokenstore.StoreTypeFile,
			description:       "Session should use file store type when specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStore()
			tokenStoreFactory := testutils.NewMockTokenStoreFactory()
			tokenStoreFactory.On("NewTokenStore", tt.expectedStoreType).Return(tokenStore, nil)

			err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory, tt.args...)
			require.NoError(t, err, tt.description)
			// Verify token store was created with expected store type
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestSessionCommand_Direct_InvalidStoreType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStoreFactory := testutils.NewMockTokenStoreFactory()

	err := testutils.ExecuteCliSessionCommand(t, ctx, tokenStoreFactory, "--store-type", "invalid")
	require.Error(t, err, "Session command should fail with invalid store type")
	assert.Contains(t, err.Error(), "unable to parse store type from string: invalid", "Error should indicate invalid store type")
}
