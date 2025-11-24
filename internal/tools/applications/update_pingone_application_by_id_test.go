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
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mockUpdateApplicationByIdSetup(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID, response *management.ReadOneApplication200Response, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("UpdateApplicationById", mock.Anything, envID, appID, mock.Anything).Return(response, httpResp, err)
}

func TestUpdateApplicationByIdHandler_MockClient(t *testing.T) {
	// Create update models from existing test apps
	testOIDCUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testOIDCApp)
	testSAMLUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testSAMLApp)
	testExternalLinkUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testExternalLinkApp)
	testP1PortalUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testP1PortalApp)
	testWSFEDUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testWSFEDApp)
	testP1SelfServiceUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testP1SelfServiceApp)

	tests := []struct {
		name             string
		input            applications.UpdateApplicationByIdInput
		setupMock        func(*mockPingOneClientApplicationsWrapper, uuid.UUID, uuid.UUID)
		wantErr          bool
		wantErrContains  string
		expectedResponse *management.ReadOneApplication200Response
	}{
		{
			name: "Success - Update OIDC application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   testOIDCUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testOIDCApp, 200, nil)
			},
			expectedResponse: &testOIDCApp,
		},
		{
			name: "Success - Update SAML application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testSAMLApp.ApplicationSAML.Id),
				Application:   testSAMLUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testSAMLApp, 200, nil)
			},
			expectedResponse: &testSAMLApp,
		},
		{
			name: "Success - Update External Link application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testExternalLinkApp.ApplicationExternalLink.Id),
				Application:   testExternalLinkUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testExternalLinkApp, 200, nil)
			},
			expectedResponse: &testExternalLinkApp,
		},
		{
			name: "Success - Update PingOne Portal application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1PortalApp.ApplicationPingOnePortal.Id),
				Application:   testP1PortalUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testP1PortalApp, 200, nil)
			},
			expectedResponse: &testP1PortalApp,
		},
		{
			name: "Success - Update WS-FED application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testWSFEDApp.ApplicationWSFED.Id),
				Application:   testWSFEDUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testWSFEDApp, 200, nil)
			},
			expectedResponse: &testWSFEDApp,
		},
		{
			name: "Success - Update PingOne Self Service application by ID",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testP1SelfServiceApp.ApplicationPingOneSelfService.Id),
				Application:   testP1SelfServiceUpdateInput,
			},
			setupMock: func(m *mockPingOneClientApplicationsWrapper, envID uuid.UUID, appID uuid.UUID) {
				mockUpdateApplicationByIdSetup(m, envID, appID, &testP1SelfServiceApp, 200, nil)
			},
			expectedResponse: &testP1SelfServiceApp,
		},
		{
			name: "Error - Application not found (404)",
			input: applications.UpdateApplicationByIdInput{
				EnvironmentId: testEnvironmentId,
				ApplicationId: uuid.MustParse(*testOIDCApp.ApplicationOIDC.Id),
				Application:   testOIDCUpdateInput,
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
				Application:   testOIDCUpdateInput,
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
				assertUpdateApplicationMatches(t, *tt.expectedResponse, output.Application)
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
				assertUpdateApplicationMatches(t, *tt.expectedResponse, outputApplication.Application)
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
	testUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testOIDCApp)

	mockClient.On("UpdateApplicationById", testutils.CancelledContextMatcher, envID, appID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
		Application:   testUpdateInput,
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
	testUpdateInput := applications.UpdateApplicationModelFromSDKReadResponse(testOIDCApp)
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: envID,
		ApplicationId: appID,
		Application:   testUpdateInput,
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

func TestUpdateApplicationByIdHandler_JSONSchemaOneOfValidation(t *testing.T) {
	testCases := []struct {
		name             string
		malformedInput   applications.UpdateApplicationModel
		expectedErrorMsg string
		description      string
	}{
		{
			name:             "Multiple application types set in input",
			malformedInput:   applications.UpdateApplicationModelFromSDKReadResponse(testMalformedMultiTypeApp), // This app has both OIDC and SAML set
			expectedErrorMsg: "oneOf: validated against both",
			description:      "violates oneOf constraint by having multiple application types set simultaneously",
		},
		{
			name:             "No application type set in input",
			malformedInput:   applications.UpdateApplicationModelFromSDKReadResponse(testMalformedEmptyApp), // This app has no application configuration set
			expectedErrorMsg: "oneOf: did not validate against any of",
			description:      "violates oneOf constraint by having no application type set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test verifies that the MCP JSON schema validation properly fails
			// when an UpdateApplicationModel violates the oneOf constraint on input
			mockClient := &mockPingOneClientApplicationsWrapper{}
			envID := testEnvironmentId
			appID := testAppId

			server := testutils.TestMcpServer(t)
			handler := applications.UpdateApplicationByIdHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			mcp.AddTool(server, applications.UpdateApplicationByIdDef.McpTool, handler)

			input := applications.UpdateApplicationByIdInput{
				EnvironmentId: envID,
				ApplicationId: appID,
				Application:   tc.malformedInput,
			}
			_, err := testutils.CallToolOverMcp(t, server, applications.UpdateApplicationByIdDef.McpTool.Name, input)

			require.Error(t, err, "Expected MCP to reject request due to JSON schema validation failure that %s", tc.description)
			assert.Contains(t, err.Error(), tc.expectedErrorMsg, "Error should mention the oneOf validation issue")

			// Mock should not be called since validation should fail before API call
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
		Application: applications.UpdateApplicationModel{
			ApplicationOIDC: &management.ApplicationOIDC{
				Name:                    "Updated App",
				Enabled:                 true,
				Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
				Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
				GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
				TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
			},
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
		Application: applications.UpdateApplicationModel{
			ApplicationOIDC: &management.ApplicationOIDC{
				Name:                    "Updated App",
				Enabled:                 true,
				Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
				Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
				GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
				TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
			},
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
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
	testUpdateInput := applications.UpdateApplicationModel{
		ApplicationOIDC: &management.ApplicationOIDC{
			Name:                    "Updated Test App",
			Description:             testutils.Pointer("Updated description"),
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	req := &mcp.CallToolRequest{}
	input := applications.UpdateApplicationByIdInput{
		EnvironmentId: testEnvironmentId,
		ApplicationId: testApplicationId,
		Application:   testUpdateInput,
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")

	// Expected values to be compared here
}
