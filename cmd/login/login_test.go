// Copyright © 2025 Ping Identity Corporation

package login_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestLoginCommand_FromRoot_Basic(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "login help flag",
			args:        []string{"login", "--help"},
			expectError: false,
			description: "Login command help should execute without error",
		},
		{
			name:          "login invalid flag",
			args:          []string{"login", "--invalid-flag"},
			expectError:   true,
			errorContains: "unknown flag",
			description:   "Login command should return error for invalid flag",
		},
		{
			name:          "login unsupported grant type",
			args:          []string{"login", "--grant-type", "unsupported_grant"},
			expectError:   true,
			errorContains: "unable to parse grant type from string: unsupported_grant",
			description:   "Login command should return error for unsupported grant type",
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

func TestLoginCommand_Direct_Success(t *testing.T) {
	tests := []struct {
		name        string
		tokenSource *testutils.StaticTokenSource
		description string
	}{
		{
			name:        "login with default static token",
			tokenSource: testutils.NewDefaultStaticTokenSource(),
			description: "Login should succeed with default static token source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStore()
			tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)
			clientFactory := testutils.NewMockAuthClientFactory(tt.tokenSource)

			err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory)

			require.NoError(t, err, tt.description)

			storedSession, err := tokenStore.GetSession()
			require.NoError(t, err, "Should be able to retrieve stored session")
			expectedToken, err := tt.tokenSource.Token()
			require.NoError(t, err, "Should be able to get token from source")

			assert.Equal(t, expectedToken.AccessToken, storedSession.AccessToken, "Stored access token should match source")
			assert.Equal(t, expectedToken.RefreshToken, storedSession.RefreshToken, "Stored refresh token should match source")
			assert.Equal(t, expectedToken.Expiry, storedSession.Expiry, "Stored expiry should match source")
			assert.NotEmpty(t, storedSession.SessionId, "Session ID should not be empty")
			_, err = uuid.Parse(storedSession.SessionId)
			require.NoError(t, err, "Session ID should be a valid UUID")
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestLoginCommand_Direct_SessionStoreError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)

	// Use a mock token store that returns errors
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStore.PutSessionError = assert.AnError
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory)

	require.Error(t, err, "Login should fail when session store returns error")
	assert.True(t, errors.Is(err, assert.AnError), "Error should contain the mock error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLoginCommand_Direct_NilAuthClientFactory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStoreFactory := testutils.NewMockTokenStoreFactory()

	err := testutils.ExecuteCliLoginCommand(t, ctx, nil, tokenStoreFactory)

	require.Error(t, err, "Login should fail with nil client")
	assert.Contains(t, err.Error(), "authClientFactory is nil", "Error should indicate nil auth client factory")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLoginCommand_Direct_NilTokenStoreFactory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)

	err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, nil)

	require.Error(t, err, "Login should fail with nil token store factory")
	assert.Contains(t, err.Error(), "tokenStoreFactory is nil", "Error should indicate nil token store factory")
}

func TestLoginCommand_Direct_ReAuthWithUnexpiredSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	// Setup an existing unexpired session
	existingTokenValue := &oauth2.Token{
		AccessToken:  "existing-access-token",
		RefreshToken: "existing-refresh-token",
		Expiry:       time.Now().Add(time.Hour * 2), // Expires in 2 hours
	}

	existingSession := auth.NewAuthSession(*existingTokenValue, "existing-session-id")
	err := tokenStore.PutSession(existingSession)
	require.NoError(t, err)

	require.True(t, existingSession.Expiry.After(time.Now()), "Existing session should be unexpired")

	err = testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory)
	require.NoError(t, err, "Login should succeed and re-authenticate even with existing unexpired session")

	// Verify the session has been refreshed
	storedSession, err := tokenStore.GetSession()
	require.NoError(t, err, "Should be able to retrieve stored session")
	assert.NotEqual(t, existingSession.SessionId, storedSession.SessionId, "Session should be refreshed with new ID")
	assert.NotEqual(t, existingSession.AccessToken, storedSession.AccessToken, "Access token should be updated")
	assert.NotEqual(t, existingSession.RefreshToken, storedSession.RefreshToken, "Refresh token should be updated")
	assert.NotEqual(t, existingSession.Expiry, storedSession.Expiry, "Expiry should be updated")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLoginCommand_Direct_ReAuthWithExpiredSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)
	tokenStore := testutils.NewInMemoryTokenStore()
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)

	// Setup an existing expired session
	expiredToken := oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "expired-refresh-token",
		Expiry:       time.Now().Add(-time.Hour), // Expired 1 hour ago
	}
	expiredSession := auth.NewAuthSession(expiredToken, "expired-session-id")
	err := tokenStore.PutSession(expiredSession)
	require.NoError(t, err)

	err = testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory)
	require.NoError(t, err, "Login should succeed when existing session is expired")

	// Verify a new session was created
	storedSession, err := tokenStore.GetSession()
	require.NoError(t, err, "Should be able to retrieve stored session")

	assert.NotEqual(t, expiredSession.SessionId, storedSession.SessionId, "Should have new session ID")
	assert.NotEqual(t, expiredSession.AccessToken, storedSession.AccessToken, "Should have new access token from token source")
	assert.NotEqual(t, expiredSession.RefreshToken, storedSession.RefreshToken, "Should have new refresh token from token source")
	assert.NotEqual(t, expiredSession.Expiry, storedSession.Expiry, "Should have new expiry from token source")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLoginCommand_Direct_GrantTypeSelection(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		expectedAccessToken  string
		expectedRefreshToken string
		description          string
	}{
		{
			name:                 "default grant type uses authorization_code",
			args:                 []string{},
			expectedAccessToken:  "authz-code-access-token",
			expectedRefreshToken: "authz-code-refresh-token",
			description:          "Login should use authorization_code token source by default",
		},
		{
			name:                 "explicit authorization_code grant type",
			args:                 []string{"--grant-type", "authorization_code"},
			expectedAccessToken:  "authz-code-access-token",
			expectedRefreshToken: "authz-code-refresh-token",
			description:          "Login should use authorization_code token source when explicitly specified",
		},
		{
			name:                 "device_code grant type",
			args:                 []string{"--grant-type", "device_code"},
			expectedAccessToken:  "device-code-access-token",
			expectedRefreshToken: "device-code-refresh-token",
			description:          "Login should use device_code token source when specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Create distinct token sources for each grant type
			authzCodeTokenSource := testutils.NewStaticTokenSource(&oauth2.Token{
				AccessToken:  "authz-code-access-token",
				RefreshToken: "authz-code-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			})
			deviceCodeTokenSource := testutils.NewStaticTokenSource(&oauth2.Token{
				AccessToken:  "device-code-access-token",
				RefreshToken: "device-code-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			})

			tokenStore := testutils.NewInMemoryTokenStore()
			tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithStore(tokenStore)
			mockClient := testutils.MockAuthClient{}
			mockClient.On("TokenSource", mock.Anything, auth.GrantTypeAuthorizationCode).Return(authzCodeTokenSource, nil)
			mockClient.On("TokenSource", mock.Anything, auth.GrantTypeDeviceCode).Return(deviceCodeTokenSource, nil)
			mockClientFactory := &testutils.MockAuthClientFactory{}
			mockClientFactory.On("NewAuthClient").Return(&mockClient, nil)

			err := testutils.ExecuteCliLoginCommand(t, ctx, mockClientFactory, tokenStoreFactory, tt.args...)
			require.NoError(t, err, tt.description)

			// Verify the correct token source was used based on grant type
			storedSession, err := tokenStore.GetSession()
			require.NoError(t, err, "Should be able to retrieve stored session")

			assert.Equal(t, tt.expectedAccessToken, storedSession.AccessToken, "Access token should match expected grant type token source")
			assert.Equal(t, tt.expectedRefreshToken, storedSession.RefreshToken, "Refresh token should match expected grant type token source")
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestLoginCommand_Direct_TokenStoreFactoryError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedError := assert.AnError
	tokenStoreFactory := testutils.NewMockTokenStoreFactoryWithError(expectedError)
	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)

	err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory)
	require.Error(t, err, "Login should fail when token store factory returns error")
	assert.Contains(t, err.Error(), expectedError.Error(), "Error should contain the factory error")
	tokenStoreFactory.AssertExpectations(t)
}

func TestLoginCommand_Direct_StoreTypeSelection(t *testing.T) {
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
			description:       "Login should use keychain store type by default",
		},
		{
			name:              "explicit keychain store type",
			args:              []string{"--store-type", "keychain"},
			expectedStoreType: tokenstore.StoreTypeKeychain,
			description:       "Login should use keychain store type when explicitly specified",
		},
		{
			name:              "file store type",
			args:              []string{"--store-type", "file"},
			expectedStoreType: tokenstore.StoreTypeFile,
			description:       "Login should use file store type when specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			tokenStore := testutils.NewInMemoryTokenStore()
			tokenStoreFactory := testutils.NewMockTokenStoreFactory()
			tokenStoreFactory.On("NewTokenStore", tt.expectedStoreType).Return(tokenStore, nil)
			tokenSource := testutils.NewDefaultStaticTokenSource()
			clientFactory := testutils.NewMockAuthClientFactory(tokenSource)

			err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory, tt.args...)
			require.NoError(t, err, tt.description)
			// Verify token store was created with expected store type
			tokenStoreFactory.AssertExpectations(t)
		})
	}
}

func TestLoginCommand_Direct_InvalidStoreType(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenStoreFactory := testutils.NewMockTokenStoreFactory()
	tokenSource := testutils.NewDefaultStaticTokenSource()
	clientFactory := testutils.NewMockAuthClientFactory(tokenSource)

	err := testutils.ExecuteCliLoginCommand(t, ctx, clientFactory, tokenStoreFactory, "--store-type", "invalid")
	require.Error(t, err, "Login should fail with invalid store type")
	assert.Contains(t, err.Error(), "unable to parse store type from string: invalid", "Error should indicate invalid store type")
}
