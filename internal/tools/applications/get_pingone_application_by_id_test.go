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

func mockGetApplicationByIdSetup(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID, response *management.ReadOneApplication200Response, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("GetApplication", mock.Anything, envID, appID).Return(response, httpResp, err)
}

func TestGetApplicationByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name             string
		input            applications.GetApplicationByIdInput
		setupMock        func(*mockPingOneClientApplicationsWrapper, uuid.UUID, uuid.UUID)
		wantErr          bool
		wantErrContains  string
		expectedResponse *management.ReadOneApplication200Response
	}{
		{
			name: "Success - Get OIDC application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testOIDCApp, 200, nil)
			},
			expectedResponse: &testOIDCApp,
		},
		{
			name: "Success - Get SAML application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testSAMLApp.ApplicationSAML.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testSAMLApp, 200, nil)
			},
			expectedResponse: &testSAMLApp,
		},
		{
			name: "Success - Get External Link application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testExternalLinkApp.ApplicationExternalLink.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testExternalLinkApp, 200, nil)
			},
			expectedResponse: &testExternalLinkApp,
		},
		{
			name: "Success - Get PingOne Portal application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1PortalApp.ApplicationPingOnePortal.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testP1PortalApp, 200, nil)
			},
			expectedResponse: &testP1PortalApp,
		},
		{
			name: "Success - Get WS-FED application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testWSFEDApp.ApplicationWSFED.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testWSFEDApp, 200, nil)
			},
			expectedResponse: &testWSFEDApp,
		},
		{
			name: "Success - Get PingOne Self Service application by ID",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1SelfServiceApp.ApplicationPingOneSelfService.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, &testP1SelfServiceApp, 200, nil)
			},
			expectedResponse: &testP1SelfServiceApp,
		},
		{
			name: "Error - Application not found (404)",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, nil, 404, errors.New("application not found"))
			},
			wantErr:         true,
			wantErrContains: "application not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationByIdSetup(m, envID, appID, nil, 200, nil)
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
			handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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
				assertReadApplicationMatches(t, *tt.expectedResponse, output.Application)
			}

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.ApplicationId)
			handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, applications.GetApplicationByIdDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, applications.GetApplicationByIdDef.McpTool.Name, tt.input)

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
			outputApplication := &applications.GetApplicationByIdOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputApplication)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.expectedResponse != nil {
				assertReadApplicationMatches(t, *tt.expectedResponse, outputApplication.Application)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationByIdHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientApplicationsWrapper{}
	envID := testEnvironmentId
	appID := testAppId
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetApplication", testutils.CancelledContextMatcher, envID, appID).Return(nil, nil, context.Canceled)

	handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
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

func TestGetApplicationByIdHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvironmentId
	appID := testAppId
	input := applications.GetApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockGetApplicationByIdSetup(mockClient, envID, appID, nil, tt.StatusCode, tt.ApiError)
			handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationByIdHandler_JSONSchemaOneOfValidation(t *testing.T) {
	testCases := []struct {
		name             string
		malformedApp     management.ReadOneApplication200Response
		expectedErrorMsg string
		description      string
	}{
		{
			name:             "Multiple application types set",
			malformedApp:     testMalformedMultiTypeApp, // This app has both OIDC and SAML set
			expectedErrorMsg: "oneOf: validated against both",
			description:      "violates oneOf constraint by having multiple application types set simultaneously",
		},
		{
			name:             "No application type set",
			malformedApp:     testMalformedEmptyApp, // This app has no application types set
			expectedErrorMsg: "oneOf: did not validate against any of",
			description:      "violates oneOf constraint by having no application type set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test verifies that the MCP JSON schema validation properly fails
			// when a ReadApplicationModel violates the oneOf constraint
			mockClient := &mockPingOneClientApplicationsWrapper{}
			envID := testEnvironmentId
			appID := testAppId

			mockGetApplicationByIdSetup(mockClient, envID, appID, &tc.malformedApp, 200, nil)

			server := testutils.TestMcpServer(t)
			handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			mcp.AddTool(server, applications.GetApplicationByIdDef.McpTool, handler)

			input := applications.GetApplicationByIdInput{
				EnvironmentId: envID,
				ApplicationId: appID,
			}
			_, err := testutils.CallToolOverMcp(t, server, applications.GetApplicationByIdDef.McpTool.Name, input)

			require.Error(t, err, "Expected MCP to reject response due to JSON schema validation failure that %s", tc.description)
			assert.Contains(t, err.Error(), tc.expectedErrorMsg, "Error should mention the specific oneOf validation issue")

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testAppId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetApplicationByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testAppId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetApplicationByIdHandler_InitializeAuthContext(t *testing.T) {
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
				mockClientFactory := &testutils.MockAuthClientFactory{}
				mockClientFactory.On("NewAuthClient").Return(mockAuthClient, nil)
				return mockAuthClient, mockClientFactory
			},
			expectTokenSourceRetrieval: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a mock get response
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockGetApplicationByIdSetup(mockClient, testEnvironmentId, testAppId, &testOIDCApp, 200, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := applications.GetApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: testAppId,
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationByIdHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestGetApplicationByIdHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)
	handler := applications.GetApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	testApplicationId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testApplicationId,
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")

	// Expected values to be compared here
}
