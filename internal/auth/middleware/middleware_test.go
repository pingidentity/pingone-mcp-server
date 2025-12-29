// Copyright © 2025 Ping Identity Corporation

package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/middleware"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	authtestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// Mock next handler for testing middleware behavior
type mockNextHandler struct {
	mock.Mock
}

func (m *mockNextHandler) Handle(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
	args := m.Called(ctx, method, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(mcp.Result), args.Error(1)
}

// TestAuthMiddleware_NonToolCallPassThrough verifies that non-tool calls bypass authentication
func TestAuthMiddleware_NonToolCallPassThrough(t *testing.T) {
	testCases := []struct {
		name   string
		method string
		req    mcp.Request
	}{
		{
			name:   "initialize method",
			method: "initialize",
			req:    &mcp.InitializeRequest{},
		},
		{
			name:   "list_tools method",
			method: "tools/list",
			req:    &mcp.ListToolsRequest{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up middleware with mock dependencies
			tokenStore := testutils.NewInMemoryTokenStore()
			authClientFactory := authtestutils.NewEmptyMockAuthClientFactory()
			middleware := middleware.NewAuthMiddleware(authClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Set up mock next handler
			nextHandler := &mockNextHandler{}
			expectedResult := &mcp.CallToolResult{}
			nextHandler.On("Handle", mock.Anything, tc.method, tc.req).Return(expectedResult, nil)

			// Create wrapped handler
			handler := middleware.Handler(nextHandler.Handle)

			// Execute
			result, err := handler(context.Background(), tc.method, tc.req)

			// Verify next handler was called directly without auth
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)
			nextHandler.AssertExpectations(t)

			// Verify auth client factory was never called
			authClientFactory.AssertNotCalled(t, "NewAuthClient")
		})
	}
}

// TestAuthMiddleware_InvalidRequestType verifies error handling for invalid request types
func TestAuthMiddleware_InvalidRequestType(t *testing.T) {
	// Set up middleware
	tokenStore := testutils.NewInMemoryTokenStore()
	authClientFactory := authtestutils.NewEmptyMockAuthClientFactory()
	middleware := middleware.NewAuthMiddleware(authClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

	// Set up mock next handler (should never be called)
	nextHandler := &mockNextHandler{}

	// Create wrapped handler
	handler := middleware.Handler(nextHandler.Handle)

	// Execute with wrong request type for tools/call
	result, err := handler(context.Background(), "tools/call", &mcp.InitializeRequest{})

	// Verify error is returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed: invalid tool call request")
	assert.Nil(t, result)

	// Verify next handler was never called
	nextHandler.AssertNotCalled(t, "Handle")
}

// TestAuthMiddleware_InitializeAuthContextError verifies error handling when auth initialization fails
func TestAuthMiddleware_InitializeAuthContextError(t *testing.T) {
	// Set up middleware with error-inducing factory
	tokenStore := testutils.NewInMemoryTokenStore()
	authClientFactory := authtestutils.NewEmptyMockAuthClientFactory()
	initContextErr := errors.New("failed to initialize auth context")
	authClientFactory.On("NewAuthClient").Return(nil, initContextErr)

	middleware := middleware.NewAuthMiddleware(authClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

	// Set up mock next handler (should never be called)
	nextHandler := &mockNextHandler{}

	// Create wrapped handler
	handler := middleware.Handler(nextHandler.Handle)

	// Create tool call request
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "list_populations",
		},
	}

	// Execute
	result, err := handler(context.Background(), "tools/call", req)

	// Verify error is returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Contains(t, err.Error(), "failed to create auth client")
	assert.Nil(t, result)

	// Verify next handler was never called
	nextHandler.AssertNotCalled(t, "Handle")

	// Verify auth client factory was called
	authClientFactory.AssertExpectations(t)
}

// TestAuthMiddleware_InitializeAuthContext verifies successful auth context initialization
func TestAuthMiddleware_InitializeAuthContext(t *testing.T) {
	testCases := []struct {
		name                       string
		grantType                  auth.GrantType
		setupTokenStore            func() *testutils.InMemoryTokenStore
		setupAuthClient            func(auth.GrantType) (*authtestutils.MockAuthClient, *authtestutils.MockAuthClientFactory)
		expectTokenSourceRetrieval bool
	}{
		{
			name:      "Auto auth - no existing session",
			grantType: auth.GrantTypeAuthorizationCode,
			setupTokenStore: func() *testutils.InMemoryTokenStore {
				return testutils.NewInMemoryTokenStore()
			},
			setupAuthClient: func(grantType auth.GrantType) (*authtestutils.MockAuthClient, *authtestutils.MockAuthClientFactory) {
				authzCodeTokenSource := testutils.NewStaticTokenSource(&oauth2.Token{
					AccessToken:  "authz-code-access-token",
					RefreshToken: "authz-code-refresh-token",
					Expiry:       time.Now().Add(time.Hour),
				})
				mockAuthClient := &authtestutils.MockAuthClient{}
				mockAuthClient.On("TokenSource", mock.Anything, grantType, (*mcp.ServerSession)(nil)).Return(authzCodeTokenSource, nil)
				mockAuthClient.On("BrowserLoginAvailable", grantType).Return(true)
				mockClientFactory := &authtestutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: true,
		},
		{
			name:      "Use existing auth session - authorization code",
			grantType: auth.GrantTypeAuthorizationCode,
			setupTokenStore: func() *testutils.InMemoryTokenStore {
				return testutils.NewInMemoryTokenStoreWithDefaultSession()
			},
			setupAuthClient: func(grantType auth.GrantType) (*authtestutils.MockAuthClient, *authtestutils.MockAuthClientFactory) {
				mockAuthClient := &authtestutils.MockAuthClient{}
				mockAuthClient.On("BrowserLoginAvailable", grantType).Return(true)
				mockClientFactory := &authtestutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: false,
		},
		{
			name:      "Use existing auth session - device code",
			grantType: auth.GrantTypeDeviceCode,
			setupTokenStore: func() *testutils.InMemoryTokenStore {
				return testutils.NewInMemoryTokenStoreWithDefaultSession()
			},
			setupAuthClient: func(grantType auth.GrantType) (*authtestutils.MockAuthClient, *authtestutils.MockAuthClientFactory) {
				mockAuthClient := &authtestutils.MockAuthClient{}
				mockAuthClient.On("BrowserLoginAvailable", grantType).Return(false)
				mockClientFactory := &authtestutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient(tc.grantType)

			// Create middleware
			middleware := middleware.NewAuthMiddleware(mockClientFactory, tokenStore, tc.grantType)

			// Set up mock next handler
			nextHandler := &mockNextHandler{}
			expectedResult := &mcp.CallToolResult{}
			nextHandler.On("Handle", mock.Anything, "tools/call", mock.Anything).Return(expectedResult, nil)

			// Create wrapped handler
			handler := middleware.Handler(nextHandler.Handle)

			// Create tool call request
			req := &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name: "list_populations",
				},
			}

			// Execute
			result, err := handler(context.Background(), "tools/call", req)

			// Verify successful execution
			require.NoError(t, err)
			assert.Equal(t, expectedResult, result)

			// Verify expectations - validate whether or not the token source was retrieved
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
			nextHandler.AssertExpectations(t)
		})
	}
}

// TestAuthMiddleware_ToolCallContextPropagation verifies that authenticated context is properly passed
func TestAuthMiddleware_ToolCallContextPropagation(t *testing.T) {
	// Set up with existing session
	tokenStore := testutils.NewInMemoryTokenStoreWithDefaultSession()

	// Set up auth client
	mockAuthClient := &authtestutils.MockAuthClient{}
	mockAuthClient.On("BrowserLoginAvailable", auth.GrantTypeAuthorizationCode).Return(true)
	mockClientFactory := &authtestutils.MockAuthClientFactory{}
	mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)

	// Create middleware
	middleware := middleware.NewAuthMiddleware(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

	// Set up mock next handler that captures context
	nextHandler := &mockNextHandler{}
	var capturedCtx context.Context
	nextHandler.On("Handle", mock.Anything, "tools/call", mock.Anything).
		Run(func(args mock.Arguments) {
			capturedCtx = args.Get(0).(context.Context)
		}).
		Return(&mcp.CallToolResult{}, nil)

	// Create wrapped handler
	handler := middleware.Handler(nextHandler.Handle)

	// Create tool call request
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name: "list_populations",
		},
	}

	// Execute
	originalCtx := context.Background()
	result, err := handler(originalCtx, "tools/call", req)

	// Verify successful execution
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify context was modified (not the same as original)
	assert.NotEqual(t, originalCtx, capturedCtx, "Context should be modified with auth information")

	// Verify next handler was called with authenticated context
	nextHandler.AssertExpectations(t)
}

// TestAuthMiddleware_MultipleToolCalls verifies middleware handles sequential tool calls
func TestAuthMiddleware_MultipleToolCalls(t *testing.T) {
	// Set up with existing session
	tokenStore := testutils.NewInMemoryTokenStoreWithDefaultSession()

	// Set up auth client
	mockAuthClient := &authtestutils.MockAuthClient{}
	mockAuthClient.On("BrowserLoginAvailable", auth.GrantTypeAuthorizationCode).Return(true)
	mockClientFactory := &authtestutils.MockAuthClientFactory{}
	mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)

	// Create middleware
	middleware := middleware.NewAuthMiddleware(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

	// Set up mock next handler
	nextHandler := &mockNextHandler{}
	expectedResult := &mcp.CallToolResult{}
	nextHandler.On("Handle", mock.Anything, "tools/call", mock.Anything).Return(expectedResult, nil)

	// Create wrapped handler
	handler := middleware.Handler(nextHandler.Handle)

	// Execute multiple tool calls
	tools := []string{"list_populations", "create_population", "get_population"}
	for _, toolName := range tools {
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Name: toolName,
			},
		}

		result, err := handler(context.Background(), "tools/call", req)

		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	}

	// Verify expectations - auth should be initialized for each call
	mockClientFactory.AssertExpectations(t)
	mockAuthClient.AssertExpectations(t)
	assert.Equal(t, len(tools), len(nextHandler.Calls))
}
