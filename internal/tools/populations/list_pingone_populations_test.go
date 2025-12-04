// Copyright © 2025 Ping Identity Corporation

package populations_test

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
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func createMockPage(populations []management.Population) testutils.LegacySdkMockPage {
	return testutils.LegacySdkMockPage{
		EntityArray: &management.EntityArray{
			Embedded: &management.EntityArrayEmbedded{
				Populations: populations,
			},
		},
		HTTPResponse: &http.Response{StatusCode: 200},
		Error:        nil,
	}
}

func setupSuccessfulMock(mockClient *mockPingOneClientPopulationsWrapper, pages [][]management.Population) {
	mockPages := make([]testutils.LegacySdkMockPage, len(pages))
	for i, pagePops := range pages {
		mockPages[i] = createMockPage(pagePops)
	}

	mockClient.On("GetPopulations", mock.Anything, testEnvironmentId, mock.Anything).Return(
		testutils.MockLegacySdkPaginationIterator(mockPages), nil)
}

func setupErrorMock(mockClient *mockPingOneClientPopulationsWrapper, err error) {
	mockClient.On("GetPopulations", mock.Anything, testEnvironmentId, mock.Anything).Return(nil, err)
}

func TestListPopulationsHandler_MockClient(t *testing.T) {
	testCases := []struct {
		name               string
		filter             *string
		setupMock          func(*mockPingOneClientPopulationsWrapper)
		expectError        bool
		expectedError      error
		expectedPopResults []management.Population
	}{
		{
			name: "Success - Single page with two populations",
			setupMock: func(mockClient *mockPingOneClientPopulationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.Population{
					{testPop1, testPop2OnlyRequiredFields},
				})
			},
			expectError:        false,
			expectedPopResults: []management.Population{testPop1, testPop2OnlyRequiredFields},
		},
		{
			name: "Error - Client returns error",
			setupMock: func(mockClient *mockPingOneClientPopulationsWrapper) {
				setupErrorMock(mockClient, assert.AnError)
			},
			expectError:   true,
			expectedError: assert.AnError,
		},
		{
			name: "Success - Empty populations list",
			setupMock: func(mockClient *mockPingOneClientPopulationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.Population{
					{}, // Empty page
				})
			},
			expectError: false,
		},
		{
			name: "Success - Multiple pages with populations",
			setupMock: func(mockClient *mockPingOneClientPopulationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.Population{
					{testPop1, testPop2OnlyRequiredFields}, // Page 1: 2 populations
					{testPop3},                             // Page 2: 1 population
					{testPop4, testPop5AllFields},          // Page 3: 2 populations
				})
			},
			expectError: false,
			// For now, expect all pages to be returned as one group
			expectedPopResults: []management.Population{
				testPop1,
				testPop2OnlyRequiredFields,
				testPop3,
				testPop4,
				testPop5AllFields,
			},
		},
		{
			name:   "Success - With combined filter",
			filter: testutils.Pointer("name sw \"Test\" and id eq \"550e8400-e29b-41d4-a716-446655440001\""),
			setupMock: func(mockClient *mockPingOneClientPopulationsWrapper) {
				setupSuccessfulMock(mockClient, [][]management.Population{
					{testPop1, testPop2OnlyRequiredFields},
				})
			},
			expectError:        false,
			expectedPopResults: []management.Population{testPop1, testPop2OnlyRequiredFields},
		},
	}

	for _, tc := range testCases {
		// Test calling the handler directly
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tc.setupMock(mockClient)

			// Run the tool handler with the mock client
			req := &mcp.CallToolRequest{}
			handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			input := populations.ListPopulationsInput{
				EnvironmentId: testEnvironmentId,
				Filter:        tc.filter,
			}

			mcpResult, structuredResponse, err := handler(context.Background(), req, input)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError), "Expected error to match")
				}
				assert.Nil(t, mcpResult)
				assert.Nil(t, structuredResponse)
				mockClient.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			assert.Nil(t, mcpResult) // MCP result is typically nil for successful operations with structured output
			require.NotNil(t, structuredResponse)
			assert.Len(t, structuredResponse.Populations, len(tc.expectedPopResults))

			for i, expectedPopData := range tc.expectedPopResults {
				if i < len(structuredResponse.Populations) {
					actualPop := structuredResponse.Populations[i]
					assertPopulationMatches(t, expectedPopData, actualPop)
				}
			}

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tc.name+" via MCP", func(t *testing.T) {
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tc.setupMock(mockClient)

			handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, populations.ListPopulationsDef.McpTool, handler)

			// Execute over MCP
			input := populations.ListPopulationsInput{
				EnvironmentId: testEnvironmentId,
				Filter:        tc.filter,
			}
			output, err := mcptestutils.CallToolOverMcp(t, server, populations.ListPopulationsDef.McpTool.Name, input)

			require.NoError(t, err, "Expect no error calling tool")
			require.NotNil(t, output, "Expect non-nil output")

			if tc.expectError {
				testutils.AssertMcpCallError(t, output, tc.expectedError.Error())
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertMcpCallSuccess(t, err, output)

			// marshal the structured content into the expected output type
			outputPopulations := &populations.ListPopulationsOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputPopulations)
			require.NoError(t, err, "Failed to unmarshal structured content")

			assert.Len(t, outputPopulations.Populations, len(tc.expectedPopResults))

			for i, expectedPopData := range tc.expectedPopResults {
				if i < len(outputPopulations.Populations) {
					actualPop := outputPopulations.Populations[i]
					assertPopulationMatches(t, expectedPopData, actualPop)
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestListPopulationsHandler_PaginationErrorMidStream(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}

	// Create an iterator that succeeds on first page but fails on second page
	page1 := createMockPage([]management.Population{testPop1, testPop2OnlyRequiredFields})
	page2 := createMockPage([]management.Population{testPop3})
	page2.Error = assert.AnError // Error on second page

	pages := []testutils.LegacySdkMockPage{page1, page2}
	mockClient.On("GetPopulations", mock.Anything, mock.Anything, mock.Anything).
		Return(testutils.MockLegacySdkPaginationIterator(pages), nil)

	handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := populations.ListPopulationsInput{
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

func TestListPopulationsHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientPopulationsWrapper{}
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetPopulations", testutils.CancelledContextMatcher, mock.Anything, mock.Anything).Return(nil, context.Canceled)

	handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := populations.ListPopulationsInput{
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

func TestListPopulationsHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientPopulationsWrapper{}
			mockClient.On("GetPopulations", mock.Anything, mock.Anything, mock.Anything).Return(nil, tt.ApiError)
			handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, populations.ListPopulationsInput{
				EnvironmentId: testEnvironmentId,
			})

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestListPopulationsHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := populations.ListPopulationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListPopulationsHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := populations.ListPopulationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestListPopulationsHandler_InitializeAuthContext(t *testing.T) {
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
			mockClient := &mockPingOneClientPopulationsWrapper{}
			setupSuccessfulMock(mockClient, [][]management.Population{
				{testPop1},
			})

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := populations.ListPopulationsInput{
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

func TestListPopulationsHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skipf("Skipping TestListPopulationsHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	// Create the client wrapper
	clientWrapper := populations.NewPingOneClientPopulationsWrapper(client)

	req := &mcp.CallToolRequest{}
	handler := populations.ListPopulationsHandler(NewMockPingOneClientPopulationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())
	input := populations.ListPopulationsInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, structuredResponse, err := handler(context.Background(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, structuredResponse, "Structured response should not be nil")

	assert.NotNil(t, structuredResponse.Populations, "Populations list should not be nil")
	assert.GreaterOrEqual(t, len(structuredResponse.Populations), 1, "Populations list should have at least one entry")
}
