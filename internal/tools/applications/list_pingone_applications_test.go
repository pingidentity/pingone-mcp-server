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
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func createMockPage(applications []management.ReadOneApplication200Response) testutils.LegacySdkMockPage {
	return testutils.LegacySdkMockPage{
		EntityArray: &management.EntityArray{
			Embedded: &management.EntityArrayEmbedded{
				Applications: applications,
			},
		},
		HTTPResponse: &http.Response{StatusCode: 200},
		Error:        nil,
	}
}

func setupSuccessfulMock(mockClient *mockPingOneClientApplicationsWrapper, pages [][]management.ReadOneApplication200Response) {
	mockPages := make([]testutils.LegacySdkMockPage, len(pages))
	for i, pageApplications := range pages {
		mockPages[i] = createMockPage(pageApplications)
	}

	mockClient.On("GetApplications", mock.Anything, testEnvironmentId).Return(
		testutils.MockLegacySdkPaginationIterator(mockPages), nil)
}

func setupErrorMock(mockClient *mockPingOneClientApplicationsWrapper, err error) {
	mockClient.On("GetApplications", mock.Anything, testEnvironmentId).Return(nil, err)
}

func assertListApplicationMatches(t *testing.T, expected management.ReadOneApplication200Response, actual applications.ApplicationSummary) {
	t.Helper()
	var expectedId, expectedName *string
	var expectedProtocol management.EnumApplicationProtocol
	var expectedType management.EnumApplicationType
	var expectedCreatedAt *time.Time
	switch {
	case expected.ApplicationExternalLink != nil:
		expectedId = expected.ApplicationExternalLink.Id
		expectedName = &expected.ApplicationExternalLink.Name
		expectedProtocol = expected.ApplicationExternalLink.Protocol
		expectedType = expected.ApplicationExternalLink.Type
		expectedCreatedAt = expected.ApplicationExternalLink.CreatedAt
	case expected.ApplicationOIDC != nil:
		expectedId = expected.ApplicationOIDC.Id
		expectedName = &expected.ApplicationOIDC.Name
		expectedProtocol = expected.ApplicationOIDC.Protocol
		expectedType = expected.ApplicationOIDC.Type
		expectedCreatedAt = expected.ApplicationOIDC.CreatedAt
	case expected.ApplicationPingOneAdminConsole != nil:
		// admin console does not define these fields
		name := "PingOne Admin Console"
		expectedName = &name
		expectedType = management.ENUMAPPLICATIONTYPE_PING_ONE_ADMIN_CONSOLE
	case expected.ApplicationPingOnePortal != nil:
		expectedId = expected.ApplicationPingOnePortal.Id
		expectedName = &expected.ApplicationPingOnePortal.Name
		expectedProtocol = expected.ApplicationPingOnePortal.Protocol
		expectedType = expected.ApplicationPingOnePortal.Type
		expectedCreatedAt = expected.ApplicationPingOnePortal.CreatedAt
	case expected.ApplicationPingOneSelfService != nil:
		expectedId = expected.ApplicationPingOneSelfService.Id
		expectedName = &expected.ApplicationPingOneSelfService.Name
		expectedProtocol = expected.ApplicationPingOneSelfService.Protocol
		expectedType = expected.ApplicationPingOneSelfService.Type
		expectedCreatedAt = expected.ApplicationPingOneSelfService.CreatedAt
	case expected.ApplicationSAML != nil:
		expectedId = expected.ApplicationSAML.Id
		expectedName = &expected.ApplicationSAML.Name
		expectedProtocol = expected.ApplicationSAML.Protocol
		expectedType = expected.ApplicationSAML.Type
		expectedCreatedAt = expected.ApplicationSAML.CreatedAt
	case expected.ApplicationWSFED != nil:
		expectedId = expected.ApplicationWSFED.Id
		expectedName = &expected.ApplicationWSFED.Name
		expectedProtocol = expected.ApplicationWSFED.Protocol
		expectedType = expected.ApplicationWSFED.Type
		expectedCreatedAt = expected.ApplicationWSFED.CreatedAt
	default:
		t.Error("Unknown application type in expected data")
	}

	require.Equal(t, expectedId == nil, actual.Id == nil, "Application ID presence should match")
	if expectedId != nil && actual.Id != nil {
		assert.Equal(t, *expectedId, *actual.Id, "Application ID should match")
	}
	require.NotNil(t, expectedName, "Expected application name should not be nil")
	assert.Equal(t, *expectedName, actual.Name, "Application Name should match")

	require.Equal(t, expectedProtocol == "", actual.Protocol == nil, "Application Protocol presence should match")
	if actual.Protocol != nil {
		assert.Equal(t, expectedProtocol, *actual.Protocol, "Application Protocol should match")
	}
	require.NotNil(t, actual.Type, "Actual application type should not be nil")
	assert.Equal(t, expectedType, *actual.Type, "Application Type should match")

	require.Equal(t, expectedCreatedAt == nil, actual.CreatedAt == nil, "Application CreatedAt presence should match")
	if expectedCreatedAt != nil && actual.CreatedAt != nil {
		assert.Equal(t, *expectedCreatedAt, *actual.CreatedAt, "Application CreatedAt should match")
	}
}

func TestListApplicationsHandler_MockClient(t *testing.T) {
	testCases := []struct {
		name                       string
		setupMock                  func(*mockPingOneClientApplicationsWrapper)
		expectError                bool
		expectedError              error
		expectedApplicationResults []management.ReadOneApplication200Response
	}{
		{
			name: "Success - Single page with two applications",
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.ReadOneApplication200Response{
					{testOIDCApp, testOIDCAppOnlyRequiredFields},
				})
			},
			expectError:                false,
			expectedApplicationResults: []management.ReadOneApplication200Response{testOIDCApp, testOIDCAppOnlyRequiredFields},
		},
		{
			name: "Error - Client returns error",
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper) {
				setupErrorMock(mockClient, assert.AnError)
			},
			expectError:   true,
			expectedError: assert.AnError,
		},
		{
			name: "Success - Empty applications list",
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.ReadOneApplication200Response{
					{}, // Empty page
				})
			},
			expectError: false,
		},
		{
			name: "Success - Multiple pages with applications",
			setupMock: func(mockClient *mockPingOneClientApplicationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.ReadOneApplication200Response{
					{testOIDCApp, testOIDCAppOnlyRequiredFields}, // Page 1: 2 applications
					{testSAMLApp}, // Page 2: 1 application
					{testSinglePageApp, testExternalLinkApp, testP1PortalApp, testWSFEDApp, testP1SelfServiceApp, testP1AdminConsoleApp}, // Page 3: 6 applications
				})
			},
			expectError: false,
			// For now, expect all pages to be returned as one group
			expectedApplicationResults: []management.ReadOneApplication200Response{
				testOIDCApp,
				testOIDCAppOnlyRequiredFields,
				testSAMLApp,
				testSinglePageApp,
				testExternalLinkApp,
				testP1PortalApp,
				testWSFEDApp,
				testP1SelfServiceApp,
				testP1AdminConsoleApp,
			},
		},
	}

	for _, tc := range testCases {
		// Test calling the handler directly
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tc.setupMock(mockClient)
			handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			input := applications.ListApplicationsInput{
				EnvironmentId: testEnvironmentId,
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
			testutils.AssertStructuredHandlerSuccess(t, err, mcpResult, output)

			assert.Len(t, output.Applications, len(tc.expectedApplicationResults))

			for i, expectedAppData := range tc.expectedApplicationResults {
				if i < len(output.Applications) {
					actualApp := output.Applications[i]
					assertListApplicationMatches(t, expectedAppData, actualApp)
				}
			}

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tc.name+" via MCP", func(t *testing.T) {
			mockClient := &mockPingOneClientApplicationsWrapper{}
			tc.setupMock(mockClient)

			handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, applications.ListApplicationsDef.McpTool, handler)

			// Execute over MCP
			input := applications.ListApplicationsInput{
				EnvironmentId: testEnvironmentId,
			}
			output, err := mcptestutils.CallToolOverMcp(t, server, applications.ListApplicationsDef.McpTool.Name, input)

			require.NoError(t, err, "Expect no error calling tool")
			require.NotNil(t, output, "Expect non-nil output")

			if tc.expectError {
				testutils.AssertMcpCallError(t, output, tc.expectedError.Error())
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertMcpCallSuccess(t, err, output)

			outputApplications := &applications.ListApplicationsOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputApplications)
			require.NoError(t, err, "Failed to unmarshal structured content")

			assert.Len(t, outputApplications.Applications, len(tc.expectedApplicationResults))

			for i, expectedAppData := range tc.expectedApplicationResults {
				if i < len(outputApplications.Applications) {
					actualApp := outputApplications.Applications[i]
					assertListApplicationMatches(t, expectedAppData, actualApp)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestListApplicationsHandler_PaginationErrorMidStream(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}

	// Create an iterator that succeeds on first page but fails on second page
	page1 := createMockPage([]management.ReadOneApplication200Response{testOIDCApp, testOIDCAppOnlyRequiredFields})
	page2 := createMockPage([]management.ReadOneApplication200Response{testSAMLApp})
	page2.Error = assert.AnError // Error on second page

	pages := []testutils.LegacySdkMockPage{page1, page2}
	mockClient.On("GetApplications", mock.Anything, mock.Anything).
		Return(testutils.MockLegacySdkPaginationIterator(pages), nil)

	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, response, err := handler(context.Background(), req, input)

	// Should fail with error from second page
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestListApplicationsHandler_EmptyEmbeddedInResponse(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}

	// Create a page with nil embedded data
	page := testutils.LegacySdkMockPage{
		EntityArray:  nil, // This triggers the "no data in response" error
		HTTPResponse: &http.Response{StatusCode: 200},
		Error:        nil,
	}

	pages := []testutils.LegacySdkMockPage{page}
	mockClient.On("GetApplications", mock.Anything, mock.Anything).
		Return(testutils.MockLegacySdkPaginationIterator(pages), nil)

	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, response, err := handler(context.Background(), req, input)

	// Should fail with "no data in response" error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no data in response")
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestListApplicationsHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientApplicationsWrapper{}
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetApplications", testutils.CancelledContextMatcher, mock.Anything).Return(nil, context.Canceled)

	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
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

func TestListApplicationsHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientApplicationsWrapper{}
			mockClient.On("GetApplications", mock.Anything, mock.Anything).Return(nil, tt.ApiError)
			handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, applications.ListApplicationsInput{
				EnvironmentId: testEnvironmentId,
			})

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestListApplicationsHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListApplicationsHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientApplicationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListApplicationsHandler_InitializeAuthContext(t *testing.T) {
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
			// Set up a mock list response
			mockClient := &mockPingOneClientApplicationsWrapper{}
			setupSuccessfulMock(mockClient, [][]management.ReadOneApplication200Response{
				{testOIDCApp},
			})

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := applications.ListApplicationsInput{
				EnvironmentId: testEnvironmentId,
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestListApplicationsHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skipf("Skipping TestListApplicationsHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	// Create the client wrapper
	clientWrapper := applications.NewPingOneClientApplicationsWrapper(client)

	req := &mcp.CallToolRequest{}
	handler := applications.ListApplicationsHandler(NewMockPingOneClientApplicationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())
	input := applications.ListApplicationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, structuredResponse, err := handler(context.Background(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, structuredResponse, "Structured response should not be nil")

	// Marshal and unmarshal to validate the structured response
	outputApplications := &applications.ListApplicationsOutput{}
	jsonBytes, err := json.Marshal(structuredResponse)
	require.NoError(t, err, "Failed to marshal structured response")
	err = json.Unmarshal(jsonBytes, outputApplications)
	require.NoError(t, err, "Failed to unmarshal structured response")

	// The applications list might be empty or contain applications
	assert.NotNil(t, outputApplications.Applications, "Applications should not be nil")
}
