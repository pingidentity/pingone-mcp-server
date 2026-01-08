// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mockUpdateApplicationSetup(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID, response *management.ReadOneApplication200Response, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("UpdateApplication", mock.Anything, envID, appID, mock.Anything).Return(response, httpResp, err)
}

func TestUpdateApplicationHandler_MockClient(t *testing.T) {
	tests := []struct {
		name             string
		input            applications.UpdateApplicationInput
		setupMock        func(*mockPingOneClientApplicationsWrapper, uuid.UUID, uuid.UUID)
		wantErr          bool
		wantErrContains  string
		expectedResponse *management.ApplicationOIDC
	}{
		{
			name: "Success - Update OIDC Web application by ID",
			input: applications.UpdateApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationSetup(m, envID, appID, &testOIDCApp, 200, nil)
			},
			expectedResponse: testOIDCApp.ApplicationOIDC,
		},
		{
			name: "Success - Update OIDC SPA by ID",
			input: applications.UpdateApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testSinglePageApp.ApplicationOIDC.Id),
				Application:   *testSinglePageApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationSetup(m, envID, appID, &testSinglePageApp, 200, nil)
			},
			expectedResponse: testSinglePageApp.ApplicationOIDC,
		},
		{
			name: "Error - Application not found (404)",
			input: applications.UpdateApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationSetup(m, envID, appID, nil, 404, errors.New("application not found"))
			},
			wantErr:         true,
			wantErrContains: "application not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: applications.UpdateApplicationInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   *testOIDCApp.ApplicationOIDC,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationSetup(m, envID, appID, nil, 200, nil)
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
			handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
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
			testutils.AssertStructuredHandlerSuccess(t, err, mcpResult, output)

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
			handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, applications.UpdateApplicationDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, applications.UpdateApplicationDef.McpTool.Name, tt.input)

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
			outputApplication := &applications.UpdateApplicationOutput{}
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

func TestUpdateApplicationHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientApplicationsWrapper{}
	envID := testEnvironmentId
	appID := testAppId

	mockClient.On("UpdateApplication", testutils.CancelledContextMatcher, envID, appID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationInput{
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

func TestUpdateApplicationHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvironmentId
	appID := testAppId
	input := applications.UpdateApplicationInput{
		EnvironmentId: envID,
		ApplicationId: appID,
		Application:   *testOIDCApp.ApplicationOIDC,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockUpdateApplicationSetup(mockClient, envID, appID, nil, tt.StatusCode, tt.ApiError)
			handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateApplicationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationInput{
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

func TestUpdateApplicationHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestUpdateApplicationHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)
	handler := applications.UpdateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil))

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	testApplicationId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Create a simple update payload (e.g., updating description)
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationInput{
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
