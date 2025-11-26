// Copyright © 2025 Ping Identity Corporation

package environments_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestUpdateEnvironmentByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.UpdateEnvironmentByIdInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *environments.UpdateEnvironmentByIdOutput)
	}{
		{
			name: "Success - Update environment with required fields only",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment Name",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Name == "Updated Environment Name" &&
						req.Region == testEnv1.region &&
						req.Type == testEnv1.envType
				}
				expectedEnv := pingone.EnvironmentResponse{
					Id:     envID,
					Name:   "Updated Environment Name",
					Region: testEnv1.region,
					Type:   testEnv1.envType,
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, &expectedEnv, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentByIdOutput) {
				assert.Equal(t, "Updated Environment Name", output.Environment.Name)
				assert.Equal(t, testEnv1.region, output.Environment.Region)
				assert.Equal(t, testEnv1.envType, output.Environment.Type)
			},
		},
		{
			name: "Success - Update environment with all optional fields",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
				Description:   testutils.Pointer("Updated description"),
				Icon:          testutils.Pointer("updated-icon"),
				Status:        testutils.Pointer(pingone.ENVIRONMENTSTATUSVALUE_ACTIVE),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Name == "Updated Environment" &&
						req.Description != nil && *req.Description == "Updated description" &&
						req.Icon != nil && *req.Icon == "updated-icon" &&
						req.Status != nil && *req.Status == pingone.ENVIRONMENTSTATUSVALUE_ACTIVE
				}
				description := "Updated description"
				icon := "updated-icon"
				status := pingone.ENVIRONMENTSTATUSVALUE_ACTIVE
				expectedEnv := pingone.EnvironmentResponse{
					Id:          envID,
					Name:        "Updated Environment",
					Region:      testEnv1.region,
					Type:        testEnv1.envType,
					Description: &description,
					Icon:        &icon,
					Status:      &status,
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, &expectedEnv, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentByIdOutput) {
				assert.Equal(t, "Updated Environment", output.Environment.Name)
				require.NotNil(t, output.Environment.Description)
				assert.Equal(t, "Updated description", *output.Environment.Description)
				require.NotNil(t, output.Environment.Icon)
				assert.Equal(t, "updated-icon", *output.Environment.Icon)
				require.NotNil(t, output.Environment.Status)
				assert.Equal(t, pingone.ENVIRONMENTSTATUSVALUE_ACTIVE, *output.Environment.Status)
			},
		},
		{
			name: "Success - Update environment with different region",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        pingone.ENVIRONMENTREGIONCODE_EU,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Region == pingone.ENVIRONMENTREGIONCODE_EU
				}
				expectedEnv := pingone.EnvironmentResponse{
					Id:     envID,
					Name:   "Updated Environment",
					Region: pingone.ENVIRONMENTREGIONCODE_EU,
					Type:   testEnv1.envType,
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, &expectedEnv, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentByIdOutput) {
				assert.Equal(t, pingone.ENVIRONMENTREGIONCODE_EU, output.Environment.Region)
			},
		},
		{
			name: "Error - Environment not found (404)",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockUpdateEnvironmentByIdSetup(m, envID, nil, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockUpdateEnvironmentByIdSetup(m, envID, nil, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no environment data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			envID := tt.input.EnvironmentId
			tt.setupMock(mockClient, envID)
			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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

			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			envID := tt.input.EnvironmentId
			tt.setupMock(mockClient, envID)
			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, environments.UpdateEnvironmentByIdDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, environments.UpdateEnvironmentByIdDef.McpTool.Name, tt.input)

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
			outputEnvironment := &environments.UpdateEnvironmentByIdOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputEnvironment)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.validateOutput != nil {
				tt.validateOutput(t, outputEnvironment)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentByIdHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	envID := testEnv1.id
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("UpdateEnvironmentById", testutils.CancelledContextMatcher, envID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
		Name:          "Updated Environment",
		Region:        testEnv1.region,
		Type:          testEnv1.envType,
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

func TestUpdateEnvironmentByIdHandler_APIErrors(t *testing.T) {
	tests := append(testutils.CommonAPIErrorTestCases(), testutils.APIErrorTestCase{
		Name:            "400 Bad Request",
		StatusCode:      400,
		ApiError:        errors.New("bad request"),
		WantErrContains: "bad request",
	})

	envID := testEnv1.id
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
		Name:          "Updated Environment",
		Region:        testEnv1.region,
		Type:          testEnv1.envType,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			mockUpdateEnvironmentByIdSetup(mockClient, envID, nil, nil, tt.StatusCode, tt.ApiError)
			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

// TestUpdateEnvironmentByIdHandler_EdgeCaseInputs tests that the handler correctly passes edge case
// input values through to the API and properly handles API validation errors.
// Note: This does NOT test InputSchema validation (which happens in the MCP SDK layer before
// the handler is called). Instead, it verifies:
// 1. Handler passes all input values (including edge cases) to the API without modification
// 2. Handler correctly wraps and returns API-level validation errors
// 3. Handler doesn't crash or drop data on unexpected inputs
func TestUpdateEnvironmentByIdHandler_EdgeCaseInputs(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.UpdateEnvironmentByIdInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper, uuid.UUID)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Empty name is passed through to API",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "", // Empty name
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				// Verify handler passes empty name to API (not filtered/validated)
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Name == ""
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, nil, 400, errors.New("name is required"))
			},
			wantErr:         true,
			wantErrContains: "name is required",
		},
		{
			name: "Whitespace-only name is passed through to API",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "   ", // Whitespace only
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				// Verify handler passes whitespace to API (not trimmed/validated)
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Name == "   "
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, nil, 400, errors.New("name cannot be whitespace"))
			},
			wantErr:         true,
			wantErrContains: "name cannot be whitespace",
		},
		{
			name: "Empty description string is passed through and accepted",
			input: environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
				Description:   testutils.Pointer(""), // Empty description
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				// Verify handler passes empty description to API
				matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
					return req.Description != nil && *req.Description == ""
				}
				description := ""
				expectedEnv := pingone.EnvironmentResponse{
					Id:          envID,
					Name:        "Updated Environment",
					Region:      testEnv1.region,
					Type:        testEnv1.envType,
					Description: &description,
				}
				mockUpdateEnvironmentByIdSetup(m, envID, matcher, &expectedEnv, 200, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			req := &mcp.CallToolRequest{}

			// Execute
			mcpResult, output, err := handler(context.Background(), req, tt.input)

			// Assert error expectations
			if tt.wantErr {
				testutils.AssertHandlerError(t, err, mcpResult, output, tt.wantErrContains)
			} else {
				testutils.AssertHandlerSuccess(t, err, mcpResult, output)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentByIdHandler_AllStatusValues(t *testing.T) {
	// Test all valid status enum values to ensure proper handling
	// Based on AllowedEnvironmentStatusValueEnumValues from PingOne SDK
	statusTests := []struct {
		name          string
		statusValue   pingone.EnvironmentStatusValue
		expectedValue pingone.EnvironmentStatusValue
	}{
		{
			name:          "Status ACTIVE",
			statusValue:   pingone.ENVIRONMENTSTATUSVALUE_ACTIVE,
			expectedValue: pingone.ENVIRONMENTSTATUSVALUE_ACTIVE,
		},
		{
			name:          "Status DELETE_PENDING",
			statusValue:   pingone.ENVIRONMENTSTATUSVALUE_DELETE_PENDING,
			expectedValue: pingone.ENVIRONMENTSTATUSVALUE_DELETE_PENDING,
		},
	}

	for _, tt := range statusTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			envID := testEnv1.id

			// Matcher verifies status is correctly passed to API
			matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
				return req.Status != nil && *req.Status == tt.expectedValue
			}

			// Mock API response with the requested status
			expectedEnv := pingone.EnvironmentResponse{
				Id:     envID,
				Name:   "Updated Environment",
				Region: testEnv1.region,
				Type:   testEnv1.envType,
				Status: &tt.expectedValue,
			}
			mockUpdateEnvironmentByIdSetup(mockClient, envID, matcher, &expectedEnv, 200, nil)

			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
			req := &mcp.CallToolRequest{}
			input := environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
				Status:        &tt.statusValue,
			}

			// Execute
			mcpResult, output, err := handler(context.Background(), req, input)

			// Assert
			require.NoError(t, err)
			assert.Nil(t, mcpResult)
			require.NotNil(t, output)
			assert.NotNil(t, output.Environment.Status)
			assert.Equal(t, tt.expectedValue, *output.Environment.Status)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentByIdHandler_StatusOmitted(t *testing.T) {
	// Test that when status is nil/omitted, it's not sent to the API
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	envID := testEnv1.id

	// Matcher verifies status is NOT set when omitted
	matcher := func(req *pingone.EnvironmentReplaceRequest) bool {
		return req.Status == nil
	}

	expectedEnv := pingone.EnvironmentResponse{
		Id:     envID,
		Name:   "Updated Environment",
		Region: testEnv1.region,
		Type:   testEnv1.envType,
		Status: nil, // No status in response
	}
	mockUpdateEnvironmentByIdSetup(mockClient, envID, matcher, &expectedEnv, 200, nil)

	handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
		Name:          "Updated Environment",
		Region:        testEnv1.region,
		Type:          testEnv1.envType,
		Status:        nil, // Status omitted
	}

	// Execute
	mcpResult, output, err := handler(context.Background(), req, input)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, mcpResult)
	require.NotNil(t, output)

	mockClient.AssertExpectations(t)
}

func TestUpdateEnvironmentByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
		Name:          "Updated Environment",
		Region:        pingone.ENVIRONMENTREGIONCODE_NA,
		Type:          pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateEnvironmentByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
		Name:          "Updated Environment",
		Region:        pingone.ENVIRONMENTREGIONCODE_NA,
		Type:          pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateEnvironmentByIdHandler_InitializeAuthContext(t *testing.T) {
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
			// Set up a mock update response
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			expectedEnv := pingone.EnvironmentResponse{
				Id:     testEnv1.id,
				Name:   "Updated Environment",
				Region: testEnv1.region,
				Type:   testEnv1.envType,
			}
			mockUpdateEnvironmentByIdSetup(mockClient, testEnv1.id, nil, &expectedEnv, 200, nil)

			// Set up auth mocks
			tokenStore := tc.setupTokenStore()
			mockAuthClient, mockClientFactory := tc.setupAuthClient()
			authContextInitializer := initialize.AuthContextInitializer(mockClientFactory, tokenStore, auth.GrantTypeAuthorizationCode)

			// Create handler and execute
			handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), authContextInitializer)
			req := &mcp.CallToolRequest{}
			input := environments.UpdateEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
				Name:          "Updated Environment",
				Region:        testEnv1.region,
				Type:          testEnv1.envType,
			}

			_, _, err := handler(context.Background(), req, input)

			require.NoError(t, err)

			// Verify expectations
			mockClientFactory.AssertExpectations(t)
			mockAuthClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentByIdHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestUpdateEnvironmentByIdHandler_RealClient since it relies on real P1 client and modifies actual resources")

	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)
	handler := environments.UpdateEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Note: Replace with a valid environment ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentByIdInput{
		EnvironmentId: testEnvironmentId,
		Name:          "Updated Test Environment",
		Region:        pingone.ENVIRONMENTREGIONCODE_NA,
		Type:          pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
		Status:        testutils.Pointer(pingone.ENVIRONMENTSTATUSVALUE_ACTIVE),
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, testEnvironmentId, response.Environment.Id, "Environment ID should match")
	assert.Equal(t, input.Name, response.Environment.Name, "Environment name should be updated")
}
