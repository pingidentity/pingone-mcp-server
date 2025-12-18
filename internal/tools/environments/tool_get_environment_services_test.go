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
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	envtestutils "github.com/pingidentity/pingone-mcp-server/internal/tools/environments/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvironmentServicesHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.GetEnvironmentServicesInput
		setupMock       func(*envtestutils.MockEnvironmentsClient, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *environments.GetEnvironmentServicesOutput)
	}{
		{
			name: "Success - Get environment services by ID",
			input: environments.GetEnvironmentServicesInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				expectedServices := createEnvironmentServicesResponse(t)
				mockGetEnvironmentServicesSetup(m, envID, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.GetEnvironmentServicesOutput) {
				assert.NotNil(t, output.Services)
				assert.NotEmpty(t, output.Services.Products)
				assert.Equal(t, 2, len(output.Services.Products))
			},
		},
		{
			name: "Success - Get environment services with complete data",
			input: environments.GetEnvironmentServicesInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				solutionType := "CUSTOMER"
				createdAt := time.Now().Add(-24 * time.Hour)
				updatedAt := time.Now()

				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
						},
					},
					SolutionType: &solutionType,
					CreatedAt:    &createdAt,
					UpdatedAt:    &updatedAt,
				}
				mockGetEnvironmentServicesSetup(m, envID, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.GetEnvironmentServicesOutput) {
				assert.NotNil(t, output.Services)
				assert.Equal(t, 2, len(output.Services.Products))
				assert.NotNil(t, output.Services.SolutionType)
				assert.Equal(t, "CUSTOMER", *output.Services.SolutionType)
				assert.NotNil(t, output.Services.CreatedAt)
				assert.NotNil(t, output.Services.UpdatedAt)
			},
		},
		{
			name: "Error - Environment not found (404)",
			input: environments.GetEnvironmentServicesInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				mockGetEnvironmentServicesSetup(m, envID, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: environments.GetEnvironmentServicesInput{
				EnvironmentId: testEnv1.id,
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				mockGetEnvironmentServicesSetup(m, envID, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no services data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil))
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
			mockClient := &envtestutils.MockEnvironmentsClient{}
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil))

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, environments.GetEnvironmentServicesDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, environments.GetEnvironmentServicesDef.McpTool.Name, tt.input)

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
			outputServices := &environments.GetEnvironmentServicesOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputServices)
			require.NoError(t, err, "Failed to unmarshal structured content")

			if tt.validateOutput != nil {
				tt.validateOutput(t, outputServices)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetEnvironmentServicesHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &envtestutils.MockEnvironmentsClient{}
	envID := testEnv1.id
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetEnvironmentServices", testutils.CancelledContextMatcher, envID).Return(nil, nil, context.Canceled)

	handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil))
	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentServicesInput{
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

func TestGetEnvironmentServicesHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnv1.id
	input := environments.GetEnvironmentServicesInput{
		EnvironmentId: testEnv1.id,
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			mockGetEnvironmentServicesSetup(mockClient, envID, nil, tt.StatusCode, tt.ApiError)
			handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil))

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestGetEnvironmentServicesHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, clientFactoryErr))
	req := &mcp.CallToolRequest{}
	input := environments.GetEnvironmentServicesInput{
		EnvironmentId: testEnv1.id,
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestGetEnvironmentServicesHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Enable when PingOne credentials are available")

	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client")

	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	handler := environments.GetEnvironmentServicesHandler(envtestutils.NewMockEnvironmentsClientFactory(clientWrapper, nil))
	input := environments.GetEnvironmentServicesInput{
		EnvironmentId: testEnvID,
	}

	// Execute
	mcpResult, output, err := handler(t.Context(), &mcp.CallToolRequest{}, input)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, mcpResult)
	require.NotNil(t, output)
	require.NotNil(t, output.Services)
	assert.NotEmpty(t, output.Services.Products, "Environment should have at least one product/service")
}
