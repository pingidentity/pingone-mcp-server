// Copyright © 2025 Ping Identity Corporation

package environments_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateEnvironmentServicesByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.UpdateEnvironmentServicesByIdInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *environments.UpdateEnvironmentServicesByIdOutput)
	}{
		{
			name: "Success - Update environment services",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
						},
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					return len(req.Products) == 2 &&
						req.Products[0].Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE &&
						req.Products[1].Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA
				}
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
						},
					},
				}
				mockUpdateEnvironmentServicesByIdSetup(m, envID, matcher, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentServicesByIdOutput) {
				assert.NotNil(t, output.Services)
				require.Equal(t, 2, len(output.Services.Products))
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE, output.Services.Products[0].Type)
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA, output.Services.Products[1].Type)
			},
		},
		{
			name: "Success - Update environment services with single product",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					return len(req.Products) == 1
				}
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
					},
				}
				mockUpdateEnvironmentServicesByIdSetup(m, envID, matcher, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentServicesByIdOutput) {
				assert.NotNil(t, output.Services)
				require.Equal(t, 1, len(output.Services.Products))
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE, output.Services.Products[0].Type)
			},
		},
		{
			name: "Success - Update environment services with multiple products",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_RISK,
						},
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					return len(req.Products) == 3
				}
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_RISK,
						},
					},
				}
				mockUpdateEnvironmentServicesByIdSetup(m, envID, matcher, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentServicesByIdOutput) {
				assert.NotNil(t, output.Services)
				require.Equal(t, 3, len(output.Services.Products))
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE, output.Services.Products[0].Type)
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA, output.Services.Products[1].Type)
				assert.Equal(t, pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_RISK, output.Services.Products[2].Type)
			},
		},
		{
			name: "Error - Environment not found (404)",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockUpdateEnvironmentServicesByIdSetup(m, envID, nil, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper, envID uuid.UUID) {
				mockUpdateEnvironmentServicesByIdSetup(m, envID, nil, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no services data in response",
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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
			handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := mcptestutils.TestMcpServer(t)
			mcp.AddTool(server, environments.UpdateEnvironmentServicesByIdDef.McpTool, handler)

			// Execute over MCP
			output, err := mcptestutils.CallToolOverMcp(t, server, environments.UpdateEnvironmentServicesByIdDef.McpTool.Name, tt.input)

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
			outputServices := &environments.UpdateEnvironmentServicesByIdOutput{}
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

func TestUpdateEnvironmentServicesByIdHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	envID := testEnv1.id
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("UpdateEnvironmentServicesById", testutils.CancelledContextMatcher, envID, mock.Anything).Return(nil, nil, context.Canceled)

	handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
			Products: []pingone.EnvironmentBillOfMaterialsProduct{
				{
					Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
				},
			},
		},
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

func TestUpdateEnvironmentServicesByIdHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	envID := testEnv1.id
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
			Products: []pingone.EnvironmentBillOfMaterialsProduct{
				{
					Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			mockUpdateEnvironmentServicesByIdSetup(mockClient, envID, nil, nil, tt.StatusCode, tt.ApiError)
			handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentServicesByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
			Products: []pingone.EnvironmentBillOfMaterialsProduct{
				{Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE},
			},
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateEnvironmentServicesByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
			Products: []pingone.EnvironmentBillOfMaterialsProduct{
				{Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE},
			},
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateEnvironmentServicesByIdHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Enable when PingOne credentials are available")

	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client")

	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)

	// Note: Replace with a valid environment and application ID from your PingOne organization
	testEnvID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Get current services first
	getHandler := environments.GetEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())
	_, getOutput, err := getHandler(t.Context(), &mcp.CallToolRequest{}, environments.GetEnvironmentServicesByIdInput{
		EnvironmentId: testEnvID,
	})
	require.NoError(t, err, "Failed to get current environment services")
	require.NotNil(t, getOutput)

	// Update with the same services (no-op update)
	handler := environments.UpdateEnvironmentServicesByIdHandler(NewMockPingOneClientEnvironmentsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnvID,
		Services: pingone.EnvironmentBillOfMaterialsReplaceRequest{
			Products: getOutput.Services.Products,
		},
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
