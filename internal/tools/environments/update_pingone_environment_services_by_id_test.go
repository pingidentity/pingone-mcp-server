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
	envtestutils "github.com/pingidentity/pingone-mcp-server/internal/tools/environments/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEnvironmentServicesByIdHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.UpdateEnvironmentServicesByIdInput
		setupMock       func(*envtestutils.MockEnvironmentsClient, uuid.UUID)
		wantErr         bool
		wantErrContains string
		validateOutput  func(*testing.T, *environments.UpdateEnvironmentServicesByIdOutput)
	}{
		{
			name: "Success - Update environment services",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call to retrieve current services (empty initially)
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

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
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call to retrieve current services (empty initially)
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

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
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA)},
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_RISK)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call to retrieve current services (empty initially)
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

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
			name: "Success - Update environment services with NEO expands to Verify and Credentials",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
					{Type: environments.NeoServiceValue},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call to retrieve current services (empty initially)
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					// Should have 3 products: BASE, VERIFY, and CREDENTIALS
					if len(req.Products) != 3 {
						return false
					}
					// Check that all three expected types are present
					hasBase := false
					hasVerify := false
					hasCredentials := false
					for _, product := range req.Products {
						switch product.Type {
						case pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE:
							hasBase = true
						case pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY:
							hasVerify = true
						case pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS:
							hasCredentials = true
						}
					}
					return hasBase && hasVerify && hasCredentials
				}
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS,
						},
					},
				}
				mockUpdateEnvironmentServicesByIdSetup(m, envID, matcher, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentServicesByIdOutput) {
				assert.NotNil(t, output.Services)
				require.Equal(t, 3, len(output.Services.Products))
				// Verify all three expected types are present
				productTypes := make(map[pingone.EnvironmentBillOfMaterialsProductType]bool)
				for _, product := range output.Services.Products {
					productTypes[product.Type] = true
				}
				assert.True(t, productTypes[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE])
				assert.True(t, productTypes[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_VERIFY])
				assert.True(t, productTypes[pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_CREDENTIALS])
			},
		},
		{
			name: "Error - Environment not found (404) on GET",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call fails with 404
				mockGetEnvironmentServicesByIdSetup(m, envID, nil, 404, errors.New("environment not found"))
			},
			wantErr:         true,
			wantErrContains: "environment not found",
		},
		{
			name: "Error - API returns nil response with no error on GET",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call returns nil response
				mockGetEnvironmentServicesByIdSetup(m, envID, nil, 200, nil)
			},
			wantErr:         true,
			wantErrContains: "no services data in response from get",
		},
		{
			name: "Success - Preserves existing product configuration when service already enabled",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
					{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA)},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call returns existing services with BASE already configured
				baseId := uuid.New()
				desc := "Existing BASE product description"
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type:        pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
							Id:          &baseId,
							Description: &desc,
						},
					},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

				// Verify that UPDATE preserves the existing BASE product configuration
				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					if len(req.Products) != 2 {
						return false
					}
					// Find BASE product and verify it has the same ID (preserved)
					hasPreservedBase := false
					hasNewMFA := false
					for _, product := range req.Products {
						if product.Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE {
							hasPreservedBase = product.Id != nil && *product.Id == baseId && product.Description != nil && *product.Description == desc
						}
						if product.Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA {
							hasNewMFA = true
						}
					}
					return hasPreservedBase && hasNewMFA
				}
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type:        pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
							Id:          &baseId,
							Description: &desc,
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
				// Verify BASE product has an ID (preserved)
				for _, product := range output.Services.Products {
					if product.Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE {
						assert.NotNil(t, product.Id, "BASE product should have preserved ID")
						assert.NotNil(t, product.Description, "BASE product should have preserved Description")
					}
				}
			},
		},
		{
			name: "Success - Update with all optional fields (Bookmarks, Console, Tags)",
			input: environments.UpdateEnvironmentServicesByIdInput{
				EnvironmentId: testEnv1.id,
				Services: []environments.EnvironmentServiceInput{
					{
						Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_DAVINCI),
						Bookmarks: []pingone.EnvironmentBillOfMaterialsProductBookmark{
							{
								Name: "Custom Dashboard",
								Href: "https://example.com/dashboard",
							},
							{
								Name: "Documentation",
								Href: "https://example.com/docs",
							},
						},
						Console: &pingone.EnvironmentBillOfMaterialsProductConsole{
							Href: "https://console.example.com/davinci",
						},
						Tags: []string{"DAVINCI_MINIMAL"},
					},
					{
						Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE),
						Bookmarks: []pingone.EnvironmentBillOfMaterialsProductBookmark{
							{
								Name: "Admin Portal",
								Href: "https://example.com/admin",
							},
						},
						Console: &pingone.EnvironmentBillOfMaterialsProductConsole{
							Href: "https://console.example.com/base",
						},
					},
				},
			},
			setupMock: func(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID) {
				// Mock GET call to retrieve current services (empty initially)
				currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{},
				}
				mockGetEnvironmentServicesByIdSetup(m, envID, currentServices, 200, nil)

				matcher := func(req *pingone.EnvironmentBillOfMaterialsReplaceRequest) bool {
					if len(req.Products) != 2 {
						return false
					}

					// Verify all optional fields are passed to the client
					hasDaVinciWithOptionalFields := false
					hasBaseWithOptionalFields := false

					for _, product := range req.Products {
						if product.Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_DAVINCI {
							hasDaVinciWithOptionalFields =
								len(product.Bookmarks) == 2 &&
									product.Bookmarks[0].Name == "Custom Dashboard" &&
									product.Bookmarks[0].Href == "https://example.com/dashboard" &&
									product.Bookmarks[1].Name == "Documentation" &&
									product.Bookmarks[1].Href == "https://example.com/docs" &&
									product.Console != nil &&
									product.Console.Href == "https://console.example.com/davinci" &&
									len(product.Tags) == 1 &&
									product.Tags[0] == "DAVINCI_MINIMAL"
						}
						if product.Type == pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE {
							hasBaseWithOptionalFields =
								len(product.Bookmarks) == 1 &&
									product.Bookmarks[0].Name == "Admin Portal" &&
									product.Bookmarks[0].Href == "https://example.com/admin" &&
									product.Console != nil &&
									product.Console.Href == "https://console.example.com/base" &&
									len(product.Tags) == 0
						}
					}

					return hasDaVinciWithOptionalFields && hasBaseWithOptionalFields
				}

				// What the mock returns doesn't matter for this test, just verifying the input to the client
				expectedServices := pingone.EnvironmentBillOfMaterialsResponse{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_DAVINCI,
						},
						{
							Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
						},
					},
				}
				mockUpdateEnvironmentServicesByIdSetup(m, envID, matcher, &expectedServices, 200, nil)
			},
			validateOutput: func(t *testing.T, output *environments.UpdateEnvironmentServicesByIdOutput) {
				assert.NotNil(t, output.Services)
				require.Equal(t, 2, len(output.Services.Products))
				// Mainly validating that the client is called, rather than what we set the mock to return
			},
		},
	}

	for _, tt := range tests {
		// Test calling the handler directly
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			tt.setupMock(mockClient, tt.input.EnvironmentId)
			handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
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
			handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())

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

	mockClient := &envtestutils.MockEnvironmentsClient{}
	envID := testEnv1.id
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("GetEnvironmentServicesById", testutils.CancelledContextMatcher, envID).Return(nil, nil, context.Canceled)

	handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: []environments.EnvironmentServiceInput{
			{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
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
		Services: []environments.EnvironmentServiceInput{
			{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &envtestutils.MockEnvironmentsClient{}
			// Mock GET call returns empty services successfully
			currentServices := &pingone.EnvironmentBillOfMaterialsResponse{
				Products: []pingone.EnvironmentBillOfMaterialsProduct{},
			}
			mockGetEnvironmentServicesByIdSetup(mockClient, envID, currentServices, 200, nil)
			// Mock UPDATE call returns the API error
			mockUpdateEnvironmentServicesByIdSetup(mockClient, envID, nil, nil, tt.StatusCode, tt.ApiError)
			handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestUpdateEnvironmentServicesByIdHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: []environments.EnvironmentServiceInput{
			{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
		},
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestUpdateEnvironmentServicesByIdHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &envtestutils.MockEnvironmentsClient{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnv1.id,
		Services: []environments.EnvironmentServiceInput{
			{Type: string(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE)},
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
	getHandler := environments.GetEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(clientWrapper, nil), testutils.MockContextInitializer())
	_, getOutput, err := getHandler(t.Context(), &mcp.CallToolRequest{}, environments.GetEnvironmentServicesByIdInput{
		EnvironmentId: testEnvID,
	})
	require.NoError(t, err, "Failed to get current environment services")
	require.NotNil(t, getOutput)

	// Update with the same services (no-op update)
	handler := environments.UpdateEnvironmentServicesByIdHandler(envtestutils.NewMockEnvironmentsClientFactory(clientWrapper, nil), testutils.MockContextInitializer())

	// Convert products to EnvironmentServiceInput slice
	var services []environments.EnvironmentServiceInput
	for _, product := range getOutput.Services.Products {
		services = append(services, environments.EnvironmentServiceInput{Type: string(product.Type)})
	}

	input := environments.UpdateEnvironmentServicesByIdInput{
		EnvironmentId: testEnvID,
		Services:      services,
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
