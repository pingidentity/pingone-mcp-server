// Copyright © 2025 Ping Identity Corporation

package populations_test

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
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to set up GetPopulation mock
func mockGetPopulationSetup(m *mockPingOneClientPopulationsWrapper, envID uuid.UUID, popID uuid.UUID, response *management.Population, statusCode int, err error) {
	httpResp := &http.Response{StatusCode: statusCode}
	m.On("GetPopulation", mock.Anything, envID, popID).Return(response, httpResp, err)
}

func TestGetPopulationHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           populations.GetPopulationInput
		setupMock       func(*mockPingOneClientPopulationsWrapper, uuid.UUID, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *populations.GetPopulationOutput)
	}{
		{
			name:  "Success - Get population by ID",
			input: getPopulationInputFromPopulation(testPop1, testEnvironmentId),
			setupMock: func(m *mockPingOneClientPopulationsWrapper, envID uuid.UUID, popID uuid.UUID) {
				mockGetPopulationSetup(m, envID, popID, &testPop1, 200, nil)
			},
			validateOutput: func(t *testing.T, output *populations.GetPopulationOutput) {
				assertPopulationMatches(t, testPop1, output.Population)
			},
		},
		{
			name:  "Error - Population not found (404)",
			input: getPopulationInputFromPopulation(testPop1, testEnvironmentId),
			setupMock: func(m *mockPingOneClientPopulationsWrapper, envID uuid.UUID, popID uuid.UUID) {
				mockGetPopulationSetup(m, envID, popID, nil, 404, errors.New("population not found"))
			},
			wantErr:         true,
			wantErrContains: "population not found",
		},
		{
			name:  "Error - API returns nil response with no error",
			input: getPopulationInputFromPopulation(testPop1, testEnvironmentId),
			setupMock: func(m *mockPingOneClientPopulationsWrapper, envID uuid.UUID, popID uuid.UUID) {
				mockGetPopulationSetup(m, envID, popID, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no population data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.PopulationId)
			handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil))
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
			mockClient := &mockPingOneClientPopulationsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId, tt.input.PopulationId)
			handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, populations.GetPopulationDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, populations.GetPopulationDef.McpTool.Name, tt.input)
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
			outputPopulation := &populations.GetPopulationOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputPopulation)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.validateOutput != nil {
				tt.validateOutput(t, outputPopulation)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetPopulationHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientPopulationsWrapper{}
	envID := testEnvironmentId
	popID := uuid.MustParse(*testPop1.Id)
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetPopulation", testutils.CancelledContextMatcher, envID, popID).Return(nil, nil, context.Canceled)

	handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := getPopulationInputFromPopulation(testPop1, testEnvironmentId)

	// Execute
	mcpResult, output, err := handler(ctx, req, input)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)

	mockClient.AssertExpectations(t)
}

func TestGetPopulationHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnvironmentId
	popID := uuid.MustParse(*testPop1.Id)
	input := getPopulationInputFromPopulation(testPop1, testEnvironmentId)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientPopulationsWrapper{}
			mockGetPopulationSetup(mockClient, envID, popID, nil, tt.StatusCode, tt.ApiError)
			handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, nil))

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetPopulationHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientPopulationsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := populations.GetPopulationInput{
		EnvironmentId: testEnvironmentId,
		PopulationId:  uuid.MustParse(*testPop1.Id),
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetPopulationHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestGetPopulationHandler_RealClient since it relies on real P1 client")

	var emptyToken string
	client, err := legacy.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(t.Context(), emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	clientWrapper := populations.NewPingOneClientPopulationsWrapper(client)
	handler := populations.GetPopulationHandler(NewMockPingOneClientPopulationsWrapperFactory(clientWrapper, nil))

	// Note: Replace with a valid environment and population ID from your PingOne organization
	testEnvironmentId := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	testPopulationId := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &mcp.CallToolRequest{}
	input := populations.GetPopulationInput{
		EnvironmentId: testEnvironmentId,
		PopulationId:  testPopulationId,
	}

	mcpResult, response, err := handler(t.Context(), req, input)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, testPopulationId.String(), *response.Population.Id, "Population ID should match")
}
