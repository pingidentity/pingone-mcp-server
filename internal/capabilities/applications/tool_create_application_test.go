// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

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

func TestCreateApplicationHandler_MockClient(t *testing.T) {
	testCases := []struct {
		name                string
		inputApplication    *management.ApplicationOIDC
		setupMock           func(*mockPingOneClientApplicationsWrapper, *management.ApplicationOIDC)
		expectError         bool
		expectedError       error
		expectedApplication *management.ApplicationOIDC
	}{
		{
			name:             "Success - Create OIDC application",
			inputApplication: testOIDCApp.ApplicationOIDC,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app *management.ApplicationOIDC) {
				expectedRequest := management.CreateApplicationRequest{
					ApplicationOIDC: app,
				}
				mockResponse := &management.CreateApplication201Response{
					ApplicationOIDC: app,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: testOIDCApp.ApplicationOIDC,
		},
		{
			name:             "Success - Create SPA application",
			inputApplication: testSinglePageApp.ApplicationOIDC,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app *management.ApplicationOIDC) {
				expectedRequest := management.CreateApplicationRequest{
					ApplicationOIDC: app,
				}
				mockResponse := &management.CreateApplication201Response{
					ApplicationOIDC: app,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: testSinglePageApp.ApplicationOIDC,
		},
		{
			name:             "Error - Client returns error",
			inputApplication: testOIDCApp.ApplicationOIDC,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app *management.ApplicationOIDC) {
				expectedRequest := management.CreateApplicationRequest{
					ApplicationOIDC: app,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(nil, &http.Response{StatusCode: 400}, assert.AnError)
			},
			expectError:   true,
			expectedError: assert.AnError,
		},
		{
			name:             "Error - Nil response",
			inputApplication: testOIDCApp.ApplicationOIDC,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app *management.ApplicationOIDC) {
				expectedRequest := management.CreateApplicationRequest{
					ApplicationOIDC: app,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(nil, &http.Response{StatusCode: 201}, nil)
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		// Test calling the handler directly
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tc.setupMock(mockClient, tc.inputApplication)
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   *tc.inputApplication,
			}

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert error handling
			if tc.expectError {
				require.Error(t, err)
				if tc.expectedError != nil {
					assert.ErrorIs(t, err, tc.expectedError)
				}
				assert.Nil(t, mcpResult)
				assert.Nil(t, output)
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success case
			require.NoError(t, err)
			assert.Nil(t, mcpResult)
			require.NotNil(t, output)

			outputApplication := &applications.CreateApplicationOutput{}
			jsonBytes, err := json.Marshal(output)
			require.NoError(t, err)
			err = json.Unmarshal(jsonBytes, outputApplication)
			require.NoError(t, err)

			assert.NotNil(t, outputApplication.Application)

			// Assert the returned application matches expected configuration
			assertOIDCApplicationMatches(t, tc.expectedApplication, &outputApplication.Application)

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tc.name+" via MCP", func(t *testing.T) {
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tc.setupMock(mockClient, tc.inputApplication)

			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, applications.CreateApplicationDef.McpTool, handler)

			// Execute over MCP
			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   *tc.inputApplication,
			}
			output, err := mcptestutils.CallToolOverMcp(t, server, applications.CreateApplicationDef.McpTool.Name, input)

			require.NoError(t, err, "Expect no error calling tool")
			require.NotNil(t, output, "Expect non-nil output")

			if tc.expectError {
				if tc.expectedError != nil {
					testutils.AssertMcpCallError(t, output, tc.expectedError.Error())
				} else {
					testutils.AssertMcpCallError(t, output, "")
				}
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertMcpCallSuccess(t, err, output)

			outputApplication := &applications.CreateApplicationOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err)
			err = json.Unmarshal(jsonBytes, outputApplication)
			require.NoError(t, err)

			assert.NotNil(t, outputApplication.Application)

			// Assert the returned application matches expected configuration
			assertOIDCApplicationMatches(t, tc.expectedApplication, &outputApplication.Application)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateApplicationHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testApp := management.CreateApplicationRequest{
		ApplicationOIDC: testOIDCApp.ApplicationOIDC,
	}

	mockClient := &mockPingOneClientApplicationsWrapper{}
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("CreateApplication", testutils.CancelledContextMatcher, testEnvironmentId, testApp).
		Return(nil, &http.Response{StatusCode: 400}, context.Canceled)

	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application:   *testApp.ApplicationOIDC,
	}

	// Execute
	mcpResult, response, err := handler(ctx, req, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestCreateApplicationHandler_APIErrors(t *testing.T) {
	testApp := management.CreateApplicationRequest{
		ApplicationOIDC: testOIDCApp.ApplicationOIDC,
	}

	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, testApp).
				Return(nil, &http.Response{StatusCode: 400}, tt.ApiError)
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil))

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   *testApp.ApplicationOIDC,
			})

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateApplicationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application: management.ApplicationOIDC{
			Name:                    "Test App",
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

func TestCreateApplicationHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skipf("Skipping TestCreateApplicationHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	// Create the client wrapper
	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)

	// Create a simple OIDC application for testing
	testApp := management.ApplicationOIDC{
		Name:                    "Test Real Client OIDC App",
		Description:             testutils.Pointer("A test OIDC application created by real client test"),
		Enabled:                 true,
		Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
		Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
		GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
		RedirectUris:            []string{"https://test.example.com/callback"},
		TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
	}

	req := &mcp.CallToolRequest{}
	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil))
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application:   testApp,
	}

	mcpResult, structuredResponse, err := handler(context.Background(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, structuredResponse, "Structured response should not be nil")

	// Marshal and unmarshal to validate the structured response
	outputApplication := &applications.CreateApplicationOutput{}
	jsonBytes, err := json.Marshal(structuredResponse)
	require.NoError(t, err, "Failed to marshal structured response")
	err = json.Unmarshal(jsonBytes, outputApplication)
	require.NoError(t, err, "Failed to unmarshal structured response")

	// Verify the created application matches the input
	assert.NotNil(t, outputApplication.Application, "Application should not be nil")
	assertOIDCApplicationMatches(t, &testApp, &outputApplication.Application)
}
