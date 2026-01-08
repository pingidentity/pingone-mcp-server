// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mockGetApplicationSetup(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID, response *management.ReadOneApplication200Response, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("GetApplication", mock.Anything, envID, appID).Return(response, httpResp, err)
}

func TestGetApplicationHandler_MockClient(t *testing.T) {
	tests := []struct {
		name             string
		input            applications.GetApplicationInput
		setupMock        func(*mockPingOneClientApplicationsWrapper, uuid.UUID, uuid.UUID)
		wantErr          bool
		wantErrContains  string
		expectedResponse any
	}{
		{
			name: "Success - Get OIDC application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testOIDCApp, 200, nil)
			},
			expectedResponse: &testOIDCApp,
		},
		{
			name: "Success - Get SAML application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testSAMLApp.ApplicationSAML.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testSAMLApp, 200, nil)
			},
			expectedResponse: &testSAMLApp,
		},
		{
			name: "Success - Get External Link application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testExternalLinkApp.ApplicationExternalLink.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testExternalLinkApp, 200, nil)
			},
			expectedResponse: &testExternalLinkApp,
		},
		{
			name: "Success - Get PingOne Portal application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1PortalApp.ApplicationPingOnePortal.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testP1PortalApp, 200, nil)
			},
			expectedResponse: &testP1PortalApp,
		},
		{
			name: "Success - Get WS-FED application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testWSFEDApp.ApplicationWSFED.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testWSFEDApp, 200, nil)
			},
			expectedResponse: &testWSFEDApp,
		},
		{
			name: "Success - Get PingOne Self Service application by ID",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1SelfServiceApp.ApplicationPingOneSelfService.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, &testP1SelfServiceApp, 200, nil)
			},
			expectedResponse: &testP1SelfServiceApp,
		},
		{
			name: "Error - Application not found (404)",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, nil, 404, errors.New("application not found"))
			},
			wantErr:         true,
			wantErrContains: "application not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: applications.GetApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockGetApplicationSetup(m, envID, appID, nil, 200, nil)
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
			handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
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
			testutils.AssertUnstructuredHandlerSuccess(t, err, mcpResult, output)

			if tt.expectedResponse != nil {
				testutils.AssertUnstructuredOutputMatches(t, mcpResult, tt.expectedResponse)
			}

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.ApplicationId)
			handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, applications.GetApplicationDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, applications.GetApplicationDef.McpTool.Name, tt.input)
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

			if tt.expectedResponse != nil {
				testutils.AssertUnstructuredOutputMatches(t, output, tt.expectedResponse)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientApplicationsWrapper{}
	envID := testEnvironmentId
	appID := testAppId
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetApplication", testutils.CancelledContextMatcher, envID, appID).Return(nil, nil, context.Canceled)

	handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationInput{
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

func TestGetApplicationHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvironmentId
	appID := testAppId
	input := applications.GetApplicationInput{
		EnvironmentId: envID,
		ApplicationId: appID,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockGetApplicationSetup(mockClient, envID, appID, nil, tt.StatusCode, tt.ApiError)
			handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetApplicationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testAppId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetApplicationHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestGetApplicationHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)
	handler := applications.GetApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil))

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	testApplicationId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := applications.GetApplicationInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testApplicationId,
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")

	// Expected values to be compared here
}
