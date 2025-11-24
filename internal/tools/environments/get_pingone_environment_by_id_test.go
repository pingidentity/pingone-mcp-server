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
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvironmentByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.GetEnvironmentByIdInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *environments.GetEnvironmentByIdOutput)
	}{
		{
			name: "Success - Get environment by ID",
			input: environments.GetEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				expectedEnv := createEnvironmentResponse(t, testEnv1)
				mockGetEnvironmentByIdSetup(m, envID, &expectedEnv, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.GetEnvironmentByIdOutput) {
				assertEnvironmentMatches(t, testEnv1, output.Environment)
			},
		},
		{
			name: "Success - Get environment with complete data",
			input: environments.GetEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				description := "Test environment description"
				status := pingone.ENVIRONMENTSTATUSVALUE_ACTIVE
				createdAt := time.Now().Add(-24 * time.Hour)
				updatedAt := time.Now()

				expectedEnv := pingone.EnvironmentResponse{
					Id:          envID,
					Name:        testEnv1.name,
					Region:      testEnv1.region,
					Type:        testEnv1.envType,
					Description: &description,
					Status:      &status,
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				}
				mockGetEnvironmentByIdSetup(m, envID, &expectedEnv, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.GetEnvironmentByIdOutput) {
				assert.Equal(t, testEnv1.name, output.Environment.Name)
				require.NotNil(t, output.Environment.Description)
				assert.Equal(t, "Test environment description", *output.Environment.Description)
				require.NotNil(t, output.Environment.Status)
				assert.Equal(t, pingone.ENVIRONMENTSTATUSVALUE_ACTIVE, *output.Environment.Status)
			},
		},
		{
			name: "Error - Environment not found (404)",
			input: environments.GetEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockGetEnvironmentByIdSetup(m, envID, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: environments.GetEnvironmentByIdInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockGetEnvironmentByIdSetup(m, envID, nil, 200, nil)
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
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, environments.GetEnvironmentByIdDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, environments.GetEnvironmentByIdDef.McpTool.Name, tt.input)

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
			outputEnvironment := &environments.GetEnvironmentByIdOutput{}
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

func TestGetEnvironmentByIdHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	envID := testEnv1.id
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetEnvironmentById", testutils.CancelledContextMatcher, envID).Return(nil, nil, context.Canceled)

	handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
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

func TestGetEnvironmentByIdHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnv1.id
	input := environments.GetEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			mockGetEnvironmentByIdSetup(mockClient, envID, nil, tt.StatusCode, tt.ApiError)
			handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetEnvironmentByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetEnvironmentByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentByIdInput{
		EnvironmentId: testEnv1.id,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetEnvironmentByIdHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestGetEnvironmentByIdHandler_RealClient since it relies on real P1 client")

	ctx := context.Background()
	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)
	handler := environments.GetEnvironmentByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Note: Replace with a valid environment ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentByIdInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, response, err := handler(ctx, req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, testEnvironmentId, response.Environment.Id, "Environment ID should match")
}
