// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

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

func TestCreateApplicationHandler_MockClient(t *testing.T) {
	// Test application configurations based on test helpers
	oidcAppConfig := applications.CreateApplicationModel{
		ApplicationOIDC: testOIDCApp.ApplicationOIDC,
	}

	samlAppConfig := applications.CreateApplicationModel{
		ApplicationSAML: testSAMLApp.ApplicationSAML,
	}

	externalLinkAppConfig := applications.CreateApplicationModel{
		ApplicationExternalLink: testExternalLinkApp.ApplicationExternalLink,
	}

	wsfedAppConfig := applications.CreateApplicationModel{
		ApplicationWSFED: testWSFEDApp.ApplicationWSFED,
	}

	testCases := []struct {
		name                string
		inputApplication    applications.CreateApplicationModel
		setupMock           func(*mockPingOneClientApplicationsWrapper, applications.CreateApplicationModel)
		expectError         bool
		expectedError       error
		expectedApplication *applications.CreateApplicationModel
	}{
		{
			name:             "Success - Create OIDC application",
			inputApplication: oidcAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
				mockResponse := &management.CreateApplication201Response{
					ApplicationOIDC: app.ApplicationOIDC,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: &oidcAppConfig,
		},
		{
			name:             "Success - Create SAML application",
			inputApplication: samlAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
				mockResponse := &management.CreateApplication201Response{
					ApplicationSAML: app.ApplicationSAML,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: &samlAppConfig,
		},
		{
			name:             "Success - Create External Link application",
			inputApplication: externalLinkAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
				mockResponse := &management.CreateApplication201Response{
					ApplicationExternalLink: app.ApplicationExternalLink,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: &externalLinkAppConfig,
		},
		{
			name:             "Success - Create WS-FED application",
			inputApplication: wsfedAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
				mockResponse := &management.CreateApplication201Response{
					ApplicationWSFED: app.ApplicationWSFED,
				}
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(mockResponse, &http.Response{StatusCode: 201}, nil)
			},
			expectError:         false,
			expectedApplication: &wsfedAppConfig,
		},
		{
			name:             "Error - Client returns error",
			inputApplication: oidcAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
				mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
					Return(nil, &http.Response{StatusCode: 400}, assert.AnError)
			},
			expectError:   true,
			expectedError: assert.AnError,
		},
		{
			name:             "Error - Nil response",
			inputApplication: oidcAppConfig,
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper, app applications.CreateApplicationModel) {
				expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(app)
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
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   tc.inputApplication,
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
			assertCreateApplicationMatches(t, *tc.expectedApplication, outputApplication.Application)

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tc.name+" via MCP", func(t *testing.T) {
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tc.setupMock(mockClient, tc.inputApplication)

			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, applications.CreateApplicationDef.McpTool, handler)

			// Execute over MCP
			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   tc.inputApplication,
			}
			output, err := testutils.CallToolOverMcp(t, server, applications.CreateApplicationDef.McpTool.Name, input)

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
			assertCreateApplicationMatches(t, *tc.expectedApplication, outputApplication.Application)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateApplicationHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testApp := applications.CreateApplicationModel{
		ApplicationOIDC: testOIDCApp.ApplicationOIDC,
	}

	mockClient := &mockPingOneClientApplicationsWrapper{}
	// Mock should return context.Canceled error when context is already cancelled
	expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(testApp)
	mockClient.On("CreateApplication", testutils.CancelledContextMatcher, testEnvironmentId, expectedRequest).
		Return(nil, &http.Response{StatusCode: 400}, context.Canceled)

	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application:   testApp,
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
	testApp := applications.CreateApplicationModel{
		ApplicationOIDC: testOIDCApp.ApplicationOIDC,
	}
	expectedRequest := applications.CreateApplicationModelToSDKCreateRequest(testApp)

	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockClient.On("CreateApplication", mock.Anything, testEnvironmentId, expectedRequest).
				Return(nil, &http.Response{StatusCode: 400}, tt.ApiError)
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   testApp,
			})

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateApplicationHandler_JSONSchemaValidationFailure(t *testing.T) {
	testCases := []struct {
		name             string
		malformedApp     applications.CreateApplicationModel
		expectedErrorMsg string
		description      string
	}{
		{
			name: "Multiple application types set",
			malformedApp: applications.CreateApplicationModel{
				ApplicationOIDC: testOIDCApp.ApplicationOIDC,
				ApplicationSAML: testSAMLApp.ApplicationSAML, // Having both OIDC and SAML violates oneOf
			},
			expectedErrorMsg: "oneOf: validated against both",
			description:      "violates oneOf constraint by having multiple application types set simultaneously",
		},
		{
			name:         "No application type set",
			malformedApp: applications.CreateApplicationModel{
				// All fields are nil/empty, violating the oneOf constraint requiring exactly one type
			},
			expectedErrorMsg: "oneOf: did not validate against any of",
			description:      "violates oneOf constraint by having no application type set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This test verifies that the MCP JSON schema validation properly fails
			// when a CreateApplicationModel violates the oneOf constraint

			// The mock won't be called because MCP should reject the input before reaching the handler
			mockClient := &mockPingOneClientApplicationsWrapper{}

			server := testutils.TestMcpServer(t)
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			mcp.AddTool(server, applications.CreateApplicationDef.McpTool, handler)

			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application:   tc.malformedApp,
			}
			_, err := testutils.CallToolOverMcp(t, server, applications.CreateApplicationDef.McpTool.Name, input)

			require.Error(t, err, "Expected MCP to reject request due to JSON schema validation failure that %s", tc.description)
			assert.Contains(t, err.Error(), tc.expectedErrorMsg, "Error should mention the specific oneOf validation issue")

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateApplicationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application: applications.CreateApplicationModel{
			ApplicationOIDC: &management.ApplicationOIDC{
				Name:                    "Test App",
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

func TestCreateApplicationHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := applications.CreateApplicationInput{
		EnvironmentId: testEnvironmentId,
		Application: applications.CreateApplicationModel{
			ApplicationOIDC: &management.ApplicationOIDC{
				Name:                    "Test App",
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

func TestCreateApplicationHandler_InitializeAuthContext(t *testing.T) {
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
			// Set up a mock create response
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockResponse := &management.CreateApplication201Response{
				ApplicationOIDC: testOIDCApp.ApplicationOIDC,
			}
			mockClient.On("CreateApplication", mock.Anything, mock.Anything, mock.Anything).
				Return(mockResponse, &http.Response{StatusCode: 201}, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := applications.CreateApplicationInput{
				EnvironmentId: testEnvironmentId,
				Application: applications.CreateApplicationModel{
					ApplicationOIDC: testOIDCApp.ApplicationOIDC,
				},
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations - validate whether or not the token source was retrieved
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
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
	testApp := applications.CreateApplicationModel{
		ApplicationOIDC: &management.ApplicationOIDC{
			Name:                    "Test Real Client OIDC App",
			Description:             testutils.Pointer("A test OIDC application created by real client test"),
			Enabled:                 true,
			Protocol:                management.ENUMAPPLICATIONPROTOCOL_OPENID_CONNECT,
			Type:                    management.ENUMAPPLICATIONTYPE_WEB_APP,
			GrantTypes:              []management.EnumApplicationOIDCGrantType{management.ENUMAPPLICATIONOIDCGRANTTYPE_AUTHORIZATION_CODE},
			RedirectUris:            []string{"https://test.example.com/callback"},
			TokenEndpointAuthMethod: management.ENUMAPPLICATIONOIDCTOKENAUTHMETHOD_CLIENT_SECRET_BASIC,
		},
	}

	req := &mcp.CallToolRequest{}
	handler := applications.CreateApplicationHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())
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
	assertCreateApplicationMatches(t, testApp, outputApplication.Application)
}
