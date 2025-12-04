// Copyright © 2025 Ping Identity Corporation

package environments_test

// Note: PingOne API has LIMITED SCIM filter support for environments endpoint.
//
// SUPPORTED operators:
//   - sw (starts with) - for 'name' attribute only
//   - eq (equal to) - for 'id', 'organization.id', 'license.id', 'status' attributes
//   - and (logical AND) - to connect multiple filters
//
// UNSUPPORTED operators (per PingOne API documentation):
//   - gt, lt, ge, le (comparison operators)
//   - in, ne, co, ew, pr (other SCIM operators)
//   - not, or (logical NOT and OR)

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	envtestutils "github.com/pingidentity/pingone-mcp-server/internal/tools/environments/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestListEnvironmentsHandler_MockClient(t *testing.T) {
	tests := []struct {
		name             string
		filter           *string
		setupMock        func(*envtestutils.MockEnvironmentsClient, *string)
		wantErr          bool
		wantErrContains  string
		wantEnvCount     int
		wantEnvironments []environmentTestData
	}{
		{
			name:             "single page with environments",
			setupMock:        mockListEnvironmentsSetup(t, nil, []environmentTestData{testEnv1, testEnv2}),
			wantEnvCount:     2,
			wantEnvironments: []environmentTestData{testEnv1, testEnv2},
		},
		{
			name: "multiple pages",
			setupMock: mockListEnvironmentsSetup(t, nil,
				[]environmentTestData{testEnv1, testEnv2},
				[]environmentTestData{testEnv3},
				[]environmentTestData{testEnv4}),
			wantEnvCount:     4,
			wantEnvironments: []environmentTestData{testEnv1, testEnv2, testEnv3, testEnv4},
		},
		{
			name:         "empty result",
			setupMock:    mockListEnvironmentsSetup(t, nil, []environmentTestData{}),
			wantEnvCount: 0,
		},
		{
			name:             "with filter",
			filter:           testutils.Pointer(`name sw "Test"`),
			setupMock:        mockListEnvironmentsSetup(t, nil, []environmentTestData{testEnv1}),
			wantEnvCount:     1,
			wantEnvironments: []environmentTestData{testEnv1},
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			tt.setupMock(mockClient, tt.filter)
			handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
			input := environments.ListEnvironmentsInput{Filter: tt.filter}

			// Execute handler directly
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert error expectations
			if tt.wantErr {
				testutils.AssertHandlerError(t, err, mcpResult, response, tt.wantErrContains)
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertStructuredHandlerSuccess(t, err, mcpResult, response)
			assert.Len(t, response.Environments, tt.wantEnvCount)

			// Verify environment details if specified
			for i, want := range tt.wantEnvironments {
				require.Less(t, i, len(response.Environments), "Not enough environments in response")
				assertEnvironmentMatches(t, want, response.Environments[i])
			}

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			tt.setupMock(mockClient, tt.filter)
			handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, environments.ListEnvironmentsDef.McpTool, handler)

			// Execute over MCP
			input := environments.ListEnvironmentsInput{Filter: tt.filter}
			output, err := mcptestutils.CallToolOverMcp(t, server, environments.ListEnvironmentsDef.McpTool.Name, input)

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
			outputEnvironments := &environments.ListEnvironmentsOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputEnvironments)
			require.NoError(t, err, "Failed to unmarshal structured content")
			assert.Len(t, outputEnvironments.Environments, tt.wantEnvCount)

			// Verify environment details if specified
			for i, want := range tt.wantEnvironments {
				require.Less(t, i, len(outputEnvironments.Environments), "Not enough environments in response")
				assertEnvironmentMatches(t, want, outputEnvironments.Environments[i])
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestListEnvironmentsHandler_PaginationErrorMidStream(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}

	// Create an iterator that succeeds on first page but fails on second page
	page1 := createMockPage(t, []environmentTestData{testEnv1, testEnv2})
	page2 := createMockPage(t, []environmentTestData{testEnv3})
	page2.Error = assert.AnError // Error on second page

	pages := []testutils.MockPage[pingone.EnvironmentsCollectionResponse]{page1, page2}
	mockClient.On("GetEnvironments", mock.Anything, mock.Anything).
		Return(testutils.MockPaginationIterator(pages), nil)

	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.ListEnvironmentsInput{}

	// Execute
	mcpResult, response, err := handler(context.Background(), req, input)

	// Assert - should fail with error from second page
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestListEnvironmentsHandler_EmptyEmbeddedInResponse(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}

	// Create a response with nil Embedded field
	page := testutils.MockPage[pingone.EnvironmentsCollectionResponse]{
		Data: &pingone.EnvironmentsCollectionResponse{
			Embedded: nil, // nil Embedded field
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
		Error:        nil,
	}

	pages := []testutils.MockPage[pingone.EnvironmentsCollectionResponse]{page}
	mockClient.On("GetEnvironments", mock.Anything, mock.Anything).
		Return(testutils.MockPaginationIterator(pages), nil)

	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.ListEnvironmentsInput{}

	// Execute
	mcpResult, response, err := handler(context.Background(), req, input)

	// Assert - should fail with error about missing data
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no data in response")
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestListEnvironmentsHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &envtestutils.MockEnvironmentsClient{}
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetEnvironments", testutils.CancelledContextMatcher, mock.Anything).Return(nil, context.Canceled)

	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.ListEnvironmentsInput{}

	// Execute
	mcpResult, response, err := handler(ctx, req, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, mcpResult)
	assert.Nil(t, response)

	mockClient.AssertExpectations(t)
}

func TestListEnvironmentsHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			mockClient.On("GetEnvironments", mock.Anything, mock.Anything).Return(nil, tt.ApiError)
			handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, environments.ListEnvironmentsInput{})

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestListEnvironmentsHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.ListEnvironmentsInput{}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListEnvironmentsHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.ListEnvironmentsInput{}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListEnvironmentsHandler_InitializeAuthContext(t *testing.T) {
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
			mockClient := &envtestutils.MockEnvironmentsClient{}
			setupFunc := mockListEnvironmentsSetup(t, nil, []environmentTestData{testEnv1})
			setupFunc(mockClient, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := environments.ListEnvironmentsInput{}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestListEnvironmentsHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Enable when PingOne credentials are available")

	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client")

	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)
	handler := environments.ListEnvironmentsHandler(envtestutils.NewMockEnvironmentsClientFactory(clientWrapper, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}

	// First, get all environments to use for filter tests
	_, allEnvsResp, err := handler(t.Context(), req, environments.ListEnvironmentsInput{})
	require.NoError(t, err, "Failed to list all environments")
	require.NotNil(t, allEnvsResp)
	require.NotEmpty(t, allEnvsResp.Environments, "No environments available for testing")

	allEnvs := allEnvsResp.Environments

	tests := []struct {
		name     string
		filter   *string
		validate func(t *testing.T, envs []pingone.EnvironmentResponse)
	}{
		{
			name: "no filter returns all environments",
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				assert.Len(t, envs, len(allEnvs), "Should return all environments")
			},
		},
		{
			name:   "filter by ID returns single environment",
			filter: testutils.Pointer(`id eq "` + allEnvs[0].Id.String() + `"`),
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				require.Len(t, envs, 1, "Should return exactly one environment")
				assert.Equal(t, allEnvs[0].Id, envs[0].Id, "Should return correct environment")
			},
		},
		{
			name: "filter by name prefix",
			filter: func() *string {
				if len(allEnvs[0].Name) < 3 {
					return testutils.Pointer(`name sw ""`)
				}
				return testutils.Pointer(`name sw "` + allEnvs[0].Name[:3] + `"`)
			}(),
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				assert.NotEmpty(t, envs, "Should find at least one environment")
			},
		},
		{
			name:   "filter by status",
			filter: testutils.Pointer(`status eq "ACTIVE"`),
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				assert.NotEmpty(t, envs, "Should find at least one environment")
			},
		},
		{
			name:   "complex AND filter",
			filter: testutils.Pointer(`id eq "` + allEnvs[0].Id.String() + `" and status eq "ACTIVE"`),
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				assert.LessOrEqual(t, len(envs), 1, "Should return at most one environment")
				if len(envs) == 1 {
					assert.Equal(t, allEnvs[0].Id, envs[0].Id)
				}
			},
		},
		{
			name:   "filter with no results",
			filter: testutils.Pointer(`id eq "00000000-0000-0000-0000-000000000000"`),
			validate: func(t *testing.T, envs []pingone.EnvironmentResponse) {
				assert.Empty(t, envs, "Should return no environments")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := environments.ListEnvironmentsInput{Filter: tt.filter}
			mcpResult, response, err := handler(t.Context(), req, input)

			require.NoError(t, err)
			assert.Nil(t, mcpResult)
			require.NotNil(t, response)

			tt.validate(t, response.Environments)
		})
	}
}
