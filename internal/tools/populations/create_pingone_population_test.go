// Copyright © 2025 Ping Identity Corporation

package populations_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestCreatePopulationHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           populations.CreatePopulationInput
		setupMock       func(*mockPingOneClientPopulationsWrapper)
		wantErr         bool
		wantErrContains string
	}{
		{
			name:  "Success - Create population with required fields only",
			input: createPopulationInputFromPopulation(testPop2OnlyRequiredFields, testEnvironmentId),
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, testEnvironmentId, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == testPop2OnlyRequiredFields.Name && req.Description == nil
				})).Return(&testPop2OnlyRequiredFields, &http.Response{StatusCode: 201}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Success - Create population with optional fields",
			input: createPopulationInputFromPopulation(testPop1, testEnvironmentId), // testPop1 has name + description
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, testEnvironmentId, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == testPop1.Name &&
						req.Description != nil &&
						*req.Description == *testPop1.Description
				})).Return(&testPop1, &http.Response{StatusCode: 201}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Success - Create population with all optional fields",
			input: createPopulationInputFromPopulation(testPop5AllFields, testEnvironmentId),
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, testEnvironmentId, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == testPop5AllFields.Name &&
						slices.Equal(req.AlternativeIdentifiers, testPop5AllFields.AlternativeIdentifiers) &&
						req.Description != nil &&
						*req.Description == *testPop5AllFields.Description &&
						req.PreferredLanguage != nil &&
						*req.PreferredLanguage == *testPop5AllFields.PreferredLanguage &&
						req.PasswordPolicy != nil &&
						req.PasswordPolicy.Id == testPop5AllFields.PasswordPolicy.Id &&
						req.Theme != nil &&
						req.Theme.Id != nil &&
						*req.Theme.Id == *testPop5AllFields.Theme.Id
				})).Return(&testPop5AllFields, &http.Response{StatusCode: 201}, nil)
			},
			wantErr: false,
		},
		{
			name: "Error - API returns error",
			input: populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "Failed Population",
			},
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, mock.Anything, mock.Anything).Return(
					nil, &http.Response{StatusCode: 400}, errors.New("bad request"))
			},
			wantErr:         true,
			wantErrContains: "bad request",
		},
		{
			name: "Error - API returns nil response",
			input: populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "Nil Response Population",
			},
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, mock.Anything, mock.Anything).Return(
					nil, &http.Response{StatusCode: 201}, nil)
			},
			wantErr:         true,
			wantErrContains: "no population data in response",
		},
	}

	for _, tc := range tests {
		// Test calling the handler directly
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tc.setupMock(mockClient)

			// Run the tool handler with the mock client
			req := &mcp.CallToolRequest{}
			handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			mcpResult, structuredResponse, err := handler(context.Background(), req, tc.input)

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrContains != "" {
					assert.Contains(t, err.Error(), tc.wantErrContains)
				}
				assert.Nil(t, mcpResult)
				assert.Nil(t, structuredResponse)
				mockClient.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			assert.Nil(t, mcpResult) // MCP result is typically nil for successful operations with structured output
			assertCreatePopulationOutput(t, tc.input, structuredResponse)

			mockClient.AssertExpectations(t)
		})

		// Test via call over MCP
		t.Run(tc.name+" via MCP", func(t *testing.T) {
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tc.setupMock(mockClient)

			handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, populations.CreatePopulationDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, populations.CreatePopulationDef.McpTool.Name, tc.input)

			require.NoError(t, err, "Expect no error calling tool")
			require.NotNil(t, output, "Expect non-nil output")

			if tc.wantErr {
				testutils.AssertMcpCallError(t, output, tc.wantErrContains)
				mockClient.AssertExpectations(t)
				return
			}

			// Assert success expectations
			testutils.AssertMcpCallSuccess(t, err, output)

			// marshal the structured content into the expected output type
			outputPopulation := &populations.CreatePopulationOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputPopulation)
			require.NoError(t, err, "Failed to unmarshal structured content")

			assertCreatePopulationOutput(t, tc.input, outputPopulation)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreatePopulationHandler_ContextCancellation(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	mockClient.On("CreatePopulation", mock.Anything, mock.Anything, mock.Anything).Return(
		nil, &http.Response{StatusCode: 0}, context.Canceled)

	req := &mcp.CallToolRequest{}
	handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	input := populations.CreatePopulationInput{
		EnvironmentId: testEnvironmentId,
		Name:          "Test Population",
	}

	mcpResult, structuredResponse, err := handler(ctx, req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, mcpResult)
	assert.Nil(t, structuredResponse)
	mockClient.AssertExpectations(t)
}

func TestCreatePopulationHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientPopulationsWrapper{}
			mockClient.On("CreatePopulation", mock.Anything, mock.Anything, mock.Anything).Return(
				nil, &http.Response{StatusCode: tt.StatusCode}, tt.ApiError)

			handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			input := createPopulationInputFromPopulation(testPop2OnlyRequiredFields, testEnvironmentId)

			// Execute
			mcpResult, response, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, response, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreatePopulationHandler_EdgeCaseInputs(t *testing.T) {
	testCases := []struct {
		name        string
		input       populations.CreatePopulationInput
		setupMock   func(*mockPingOneClientPopulationsWrapper)
		expectedErr string
	}{
		{
			name: "Empty name is passed through to API",
			input: populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "",
			},
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, mock.Anything, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == ""
				})).Return(nil, &http.Response{StatusCode: 400}, errors.New("name is required"))
			},
			expectedErr: "name is required",
		},
		{
			name: "Whitespace-only name is passed through to API",
			input: populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "   ",
			},
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				m.On("CreatePopulation", mock.Anything, mock.Anything, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == "   "
				})).Return(nil, &http.Response{StatusCode: 400}, errors.New("name cannot be whitespace"))
			},
			expectedErr: "name cannot be whitespace",
		},
		{
			name: "Empty description string is passed through and accepted",
			input: populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "Test Population",
				Description:   testutils.Pointer(""),
			},
			setupMock: func(m *mockPingOneClientPopulationsWrapper) {
				createdPopID := "550e8400-e29b-41d4-a716-446655441003"
				mockPopulation := &management.Population{
					Id:          &createdPopID,
					Name:        "Test Population",
					Description: testutils.Pointer(""),
				}
				m.On("CreatePopulation", mock.Anything, mock.Anything, mock.MatchedBy(func(req management.Population) bool {
					return req.Name == "Test Population" &&
						req.Description != nil &&
						*req.Description == ""
				})).Return(mockPopulation, &http.Response{StatusCode: 201}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tc.setupMock(mockClient)

			req := &mcp.CallToolRequest{}
			handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			mcpResult, structuredResponse, err := handler(context.Background(), req, tc.input)

			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
				assert.Nil(t, mcpResult)
				assert.Nil(t, structuredResponse)
			} else {
				require.NoError(t, err)
				assert.Nil(t, mcpResult)
				assertCreatePopulationOutput(t, tc.input, structuredResponse)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreatePopulationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := populations.CreatePopulationInput{
		EnvironmentId: testEnvironmentId,
		Name:          "Test Population",
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestCreatePopulationHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := populations.CreatePopulationInput{
		EnvironmentId: testEnvironmentId,
		Name:          "Test Population",
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestCreatePopulationHandler_InitializeAuthContext(t *testing.T) {
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
			// Set up a mock create response
			mockClient := &mockPingOneClientPopulationsWrapper{}
			mockClient.On("CreatePopulation", mock.Anything, testEnvironmentId, mock.Anything).Return(
				&testPop1, &http.Response{StatusCode: 201}, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := populations.CreatePopulationInput{
				EnvironmentId: testEnvironmentId,
				Name:          "Test Population",
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestCreatePopulationHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skipf("Skipping TestCreatePopulationHandler_RealClient since it relies on real P1 client and creates actual resources")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := populations.NewPingOneClientPopulationsWrapper(client)
	handler := populations.CreatePopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Note: Replace with a valid environment ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := populations.CreatePopulationInput{
		EnvironmentId: testEnvironmentId,
		Name:          "Test Population from Real Client",
		Description:   testutils.Pointer("Created by automated test for real client validation"),
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	assertCreatePopulationOutput(t, input, response)
}
