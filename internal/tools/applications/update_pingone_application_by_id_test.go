// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func mockUpdateApplicationByIdSetup(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID, response *management.ReadOneApplication200Response, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("UpdateApplicationById", mock.Anything, envID, appID, mock.Anything).Return(response, httpResp, err)
}

func TestUpdateApplicationByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name             string
		input            applications.UpdateApplicationByIdInput
		setupMock        func(*mockPingOneClientApplicationsWrapper, uuid.UUID, uuid.UUID)
		wantErr          bool
		wantErrContains  string
		expectedResponse *management.ApplicationOIDC
	}{
		{
			name: "Success - Update OIDC Web application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testOIDCApp, 200, nil)
			},
			expectedResponse: testOIDCApp.ApplicationOIDC,
		},
		{
			name: "Success - Update OIDC SPA by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testSinglePageApp.ApplicationOIDC.Id),
				Application:   *testSinglePageApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testSinglePageApp, 200, nil)
			},
			expectedResponse: testSinglePageApp.ApplicationOIDC,
		},
		{
			name: "Error - Application not found (404)",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, nil, 404, errors.New("application not found"))
			},
			wantErr:         true,
			wantErrContains: "application not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no application data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.ApplicationId)
			handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			req := &mcp.CallToolRequest{}

			// Execute
			mcpResult, output, err := handler(context.Background(), req, tt.input)

			// Assert error expectations
			if tt.wantErr {
				testutils.AssertHandlerError(t, err, mcpResult, output, tt.wantErrContains)
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertHandlerSuccess(t, err, mcpResult, output)

			if tt.expectedResponse != nil {
				assertOIDCApplicationMatches(t, tt.expectedResponse, &output.Application)
			}

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.ApplicationId)
			handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, applications.UpdateApplicationByIdDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, applications.UpdateApplicationByIdDef.McpTool.Name, tt.input)

			require.NoError(t, err, "Expect no error calling tool")
			require.NotNil(t, output, "Expect non-nil output")

			// Assert error expectations
			if tt.wantErr {
				testutils.AssertMcpCallError(t, output, tt.wantErrContains)
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertMcpCallSuccess(t, err, output)

			// marshal the structured content into the expected output type
			outputApplication := &applications.UpdateApplicationByIdOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputApplication)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.expectedResponse != nil {
				assertOIDCApplicationMatches(t, tt.expectedResponse, &outputApplication.Application)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateApplicationByIdHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientApplicationsWrapper{}
	envID := testEnvironmentId
	appID := testAppId

	mockClient.On("UpdateApplicationById", testutils.CancelledContextMatcher, envID, appID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
		Application:   *testOIDCApp.ApplicationOIDC,
	}

	// Execute
	mcpResult, output, err := handler(ctx, req, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)

	mockClient.AssertExpectations(t)
}

func TestUpdateApplicationByIdHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvironmentId
	appID := testAppId
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
		Application:   *testOIDCApp.ApplicationOIDC,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockUpdateApplicationByIdSetup(mockClient, envID, appID, nil, tt.StatusCode, tt.ApiError)
			handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateApplicationByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testAppId,
		Application: management.ApplicationOIDC{
			Name:                    "Updated App",
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateApplicationByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testAppId,
		Application: management.ApplicationOIDC{
			Name:                    "Updated App",
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateApplicationByIdHandler_InitializeAuthContext(t *testing.T) {
	testCases := []struct {
		name                       string
		setupTokenStore            func() *testutils.InMemoryTokenStore
		setupAuthClient            func() (*testutils.MockAuthClient, *testutils.MockAuthClientFactory)
		expectTokenSourceRetrieval bool
	}{
		{
			name: "Auto auth - no existing session",
			setupTokenStore: func() *testutils.InMemoryTokenStore {
				return testutils.NewInMemoryTokenStore()
			},
			setupAuthClient: func() (*testutils.MockAuthClient, *testutils.MockAuthClientFactory) {
				authzCodeTokenSource := testutils.NewStaticTokenSource(&oauth2.Token{
					AccessToken:  "authz-code-access-token",
					RefreshToken: "authz-code-refresh-token",
					Expiry:       time.Now().Add(time.Hour),
				})
				mockAuthClient := &testutils.MockAuthClient{}
				mockAuthClient.On("TokenSource", mock.Anything, auth.GrantTypeAuthorizationCode).Return(authzCodeTokenSource, nil)
				mockAuthClient.On("BrowserLoginAvailable", auth.GrantTypeAuthorizationCode).Return(true)
				mockClientFactory := &testutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: true,
		},
		{
			name: "Use existing auth session",
			setupTokenStore: func() *testutils.InMemoryTokenStore {
				return testutils.NewInMemoryTokenStoreWithDefaultSession()
			},
			setupAuthClient: func() (*testutils.MockAuthClient, *testutils.MockAuthClientFactory) {
				mockAuthClient := &testutils.MockAuthClient{}
				mockAuthClient.On("BrowserLoginAvailable", auth.GrantTypeAuthorizationCode).Return(true)
				mockClientFactory := &testutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a mock update response
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockUpdateApplicationByIdSetup(mockClient, testEnvironmentId, testAppId, &testOIDCApp, 200, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: testAppId,
				Application:   *testOIDCApp.ApplicationOIDC,
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestUpdateApplicationByIdHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestUpdateApplicationByIdHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)
	handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	testApplicationId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Create a simple update payload (e.g., updating description)
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testApplicationId,
		Application: management.ApplicationOIDC{
			Name:                    "Updated Test App",
			Description:             testutils.Pointer("Updated description"),
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")

	// Expected values to be compared here
}
