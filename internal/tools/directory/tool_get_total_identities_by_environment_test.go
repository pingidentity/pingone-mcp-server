// Copyright © 2025 Ping Identity Corporation

package directory_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/directory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTotalIdentitiesByEnvironmentHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           directory.GetTotalIdentitiesByEnvironmentInput
		setupMock       func(*mockPingOneClientDirectoryWrapper, uuid.UUID, string)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *directory.GetTotalIdentitiesByEnvironmentOutput)
	}{
		{
			name: "Success - Get total identities with default date range",
			input: directory.GetTotalIdentitiesByEnvironmentInput{
				EnvironmentId: testEnvId,
			},
			setupMock: func(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, filter string) {
				expectedReport := createTotalIdentitiesResponse(t)
				mockGetTotalIdentitiesByEnvironmentSetup(m, envID, &expectedReport, 200, nil)
			},
			validateOutput: func(t *testing.T, output *directory.GetTotalIdentitiesByEnvironmentOutput) {
				require.NotNil(t, output)
				assert.NotNil(t, output.TotalIdentitiesReport)
			},
		},
		{
			name: "Success - Get total identities with custom date range",
			input: directory.GetTotalIdentitiesByEnvironmentInput{
				EnvironmentId: testEnvId,
				StartDate:     &testStartDate,
				EndDate:       &testEndDate,
			},
			setupMock: func(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, filter string) {
				expectedReport := createTotalIdentitiesResponse(t)
				mockGetTotalIdentitiesByEnvironmentSetup(m, envID, &expectedReport, 200, nil)
			},
			validateOutput: func(t *testing.T, output *directory.GetTotalIdentitiesByEnvironmentOutput) {
				require.NotNil(t, output)
				assert.NotNil(t, output.TotalIdentitiesReport)
			},
		},
		{
			name: "Error - Environment not found (404)",
			input: directory.GetTotalIdentitiesByEnvironmentInput{
				EnvironmentId: testEnvId,
			},
			setupMock: func(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, filter string) {
				mockGetTotalIdentitiesByEnvironmentSetup(m, envID, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: directory.GetTotalIdentitiesByEnvironmentInput{
				EnvironmentId: testEnvId,
			},
			setupMock: func(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, filter string) {
				mockGetTotalIdentitiesByEnvironmentSetup(m, envID, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no total identities report data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientDirectoryWrapper{}
			// Calculate filter string based on input dates
			filter := calculateFilter(tt.input)
			tt.setupMock(mockClient, tt.input.EnvironmentId, filter)
			handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(mockClient, nil))
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

			if tt.validateOutput != nil {
				tt.validateOutput(t, output)
			}

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientDirectoryWrapper{}
			filter := calculateFilter(tt.input)
			tt.setupMock(mockClient, tt.input.EnvironmentId, filter)
			handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, directory.GetTotalIdentitiesByEnvironmentDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, directory.GetTotalIdentitiesByEnvironmentDef.McpTool.Name, tt.input)

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
			outputReport := &directory.GetTotalIdentitiesByEnvironmentOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputReport)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.validateOutput != nil {
				tt.validateOutput(t, outputReport)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetTotalIdentitiesByEnvironmentHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientDirectoryWrapper{}
	envID := testEnvId
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetTotalIdentitiesByEnvironmentId", testutils.CancelledContextMatcher, envID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := directory.GetTotalIdentitiesByEnvironmentInput{
		EnvironmentId: testEnvId,
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

func TestGetTotalIdentitiesByEnvironmentHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvId
	input := directory.GetTotalIdentitiesByEnvironmentInput{
		EnvironmentId: testEnvId,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientDirectoryWrapper{}
			mockGetTotalIdentitiesByEnvironmentSetup(mockClient, envID, nil, tt.StatusCode, tt.ApiError)
			handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(mockClient, nil))

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetTotalIdentitiesByEnvironmentHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientDirectoryWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := directory.GetTotalIdentitiesByEnvironmentInput{
		EnvironmentId: testEnvId,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetTotalIdentitiesByEnvironmentHandler_RealClient(t *testing.T) {
	//TODO enable test when we can run against a real P1 client
	t.Skip("Skipping TestGetTotalIdentitiesByEnvironmentHandler_RealClient since it relies on real P1 client")

	ctx := context.Background()
	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := directory.NewPingOneClientDirectoryWrapper(client)
	handler := directory.GetTotalIdentitiesByEnvironmentHandler(NewMockPingOneClientDirectoryWrapperFactory(clientWrapper, nil))

	// Note: Replace with a valid environment ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := directory.GetTotalIdentitiesByEnvironmentInput{
		EnvironmentId: testEnvironmentId,
	}

	mcpResult, response, err := handler(ctx, req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")
	assert.NotNil(t, response.TotalIdentitiesReport, "Total identities report should not be nil")
}
