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
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateEnvironmentHandler_MockClient(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.CreateEnvironmentInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Success - Create SANDBOX environment with required fields only",
			input: environments.CreateEnvironmentInput{
				Name:    "New Test Environment",
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				createdEnvID := uuid.MustParse("550e8400-e29b-41d4-a716-446655441001")
				mockCreateEnvironmentSetup(m,
					func(req *pingone.EnvironmentCreateRequest) bool {
						return req.Name == "New Test Environment" &&
							req.Region == pingone.ENVIRONMENTREGIONCODE_NA &&
							req.Type == pingone.ENVIRONMENTTYPEVALUE_SANDBOX &&
							req.License.Id == testLicenseID
					},
					&pingone.EnvironmentResponse{
						Id:     createdEnvID,
						Name:   "New Test Environment",
						Region: pingone.ENVIRONMENTREGIONCODE_NA,
						Type:   pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
					},
					201,
					nil,
				)
			},
		},
		{
			name: "Success - Create environment with optional fields",
			input: environments.CreateEnvironmentInput{
				Name:        "Full Featured Environment",
				Region:      pingone.ENVIRONMENTREGIONCODE_EU,
				License:     *pingone.NewEnvironmentLicense(testLicenseID),
				Description: testutils.Pointer("A test environment with description"),
				Icon:        testutils.Pointer("test-icon"),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				createdEnvID := uuid.MustParse("550e8400-e29b-41d4-a716-446655441002")
				mockCreateEnvironmentSetup(m,
					func(req *pingone.EnvironmentCreateRequest) bool {
						hasDesc := req.Description != nil && *req.Description == "A test environment with description"
						hasIcon := req.Icon != nil && *req.Icon == "test-icon"
						return req.Name == "Full Featured Environment" &&
							req.Region == pingone.ENVIRONMENTREGIONCODE_EU &&
							hasDesc && hasIcon
					},
					&pingone.EnvironmentResponse{
						Id:          createdEnvID,
						Name:        "Full Featured Environment",
						Region:      pingone.ENVIRONMENTREGIONCODE_EU,
						Type:        pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
						Description: testutils.Pointer("A test environment with description"),
						Icon:        testutils.Pointer("test-icon"),
					},
					201,
					nil,
				)
			},
		},
		{
			name: "Success - Create environment with bill of materials",
			input: environments.CreateEnvironmentInput{
				Name:    "Environment with BOM",
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
				BillOfMaterials: &pingone.EnvironmentBillOfMaterials{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{
						*pingone.NewEnvironmentBillOfMaterialsProduct(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE),
						*pingone.NewEnvironmentBillOfMaterialsProduct(pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA),
					},
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				createdEnvID := uuid.MustParse("550e8400-e29b-41d4-a716-446655441004")
				mockCreateEnvironmentSetup(m,
					func(req *pingone.EnvironmentCreateRequest) bool {
						hasBOM := req.BillOfMaterials != nil && len(req.BillOfMaterials.Products) == 2
						return req.Name == "Environment with BOM" &&
							req.Region == pingone.ENVIRONMENTREGIONCODE_NA &&
							hasBOM
					},
					&pingone.EnvironmentResponse{
						Id:     createdEnvID,
						Name:   "Environment with BOM",
						Region: pingone.ENVIRONMENTREGIONCODE_NA,
						Type:   pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
					},
					201,
					nil,
				)
			},
		},
		{
			name: "Success - Create environment in different regions",
			input: environments.CreateEnvironmentInput{
				Name:    "AP Environment",
				Region:  pingone.ENVIRONMENTREGIONCODE_AP,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				createdEnvID := uuid.MustParse("550e8400-e29b-41d4-a716-446655441003")
				mockCreateEnvironmentSetup(m, nil,
					&pingone.EnvironmentResponse{
						Id:     createdEnvID,
						Name:   "AP Environment",
						Region: pingone.ENVIRONMENTREGIONCODE_AP,
						Type:   pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
					},
					201,
					nil,
				)
			},
		},
		{
			name: "Error - API returns nil response",
			input: environments.CreateEnvironmentInput{
				Name:    "Test Environment",
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				mockCreateEnvironmentSetup(m, nil, nil, 201, nil)
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
			tt.setupMock(mockClient)
			handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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
			assert.NotEqual(t, uuid.Nil, output.Environment.Id)
			assert.Equal(t, pingone.ENVIRONMENTTYPEVALUE_SANDBOX, output.Environment.Type, "Environment type should always be SANDBOX")

			// Validate that output matches input using shared helper
			assertCreateEnvironmentOutput(t, tt.input, output)

			mockClient.AssertExpectations(t)
		})
		// Test via call over MCP
		t.Run(tt.name+" via MCP", func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			tt.setupMock(mockClient)
			handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			server := testutils.TestMcpServer(t)
			mcp.AddTool(server, environments.CreateEnvironmentDef.McpTool, handler)

			// Execute over MCP
			output, err := testutils.CallToolOverMcp(t, server, environments.CreateEnvironmentDef.McpTool.Name, tt.input)

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
			outputEnvironment := &environments.CreateEnvironmentOutput{}
			jsonBytes, err := json.Marshal(output.StructuredContent)
			require.NoError(t, err, "Failed to marshal structured content")
			err = json.Unmarshal(jsonBytes, outputEnvironment)
			require.NoError(t, err, "Failed to unmarshal structured content")

			assert.NotEqual(t, uuid.Nil, outputEnvironment.Environment.Id)
			assert.Equal(t, pingone.ENVIRONMENTTYPEVALUE_SANDBOX, outputEnvironment.Environment.Type, "Environment type should always be SANDBOX")

			// Validate that output matches input using shared helper
			assertCreateEnvironmentOutput(t, tt.input, outputEnvironment)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateEnvironmentHandler_ContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	// Mock should return context.Canceled error when context is already cancelled
	mockClient.On("CreateEnvironment", testutils.CancelledContextMatcher, mock.Anything).Return(nil, nil, context.Canceled)

	handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.CreateEnvironmentInput{
		Name:    "Test Environment",
		Region:  pingone.ENVIRONMENTREGIONCODE_NA,
		License: *pingone.NewEnvironmentLicense(testLicenseID),
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

// TestCreateEnvironmentHandler_EdgeCaseInputs tests that the handler correctly passes edge case
// input values through to the API and properly handles API validation errors.
// Note: This does NOT test InputSchema validation (which happens in the MCP SDK layer before
// the handler is called). Instead, it verifies:
// 1. Handler passes all input values (including edge cases) to the API without modification
// 2. Handler correctly wraps and returns API-level validation errors
// 3. Handler doesn't crash or drop data on unexpected inputs
func TestCreateEnvironmentHandler_EdgeCaseInputs(t *testing.T) {
	tests := []struct {
		name            string
		input           environments.CreateEnvironmentInput
		setupMock       func(*mockPingOneClientEnvironmentsWrapper)
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "Empty name is passed through to API",
			input: environments.CreateEnvironmentInput{
				Name:    "", // Empty name
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				// Verify handler passes empty name to API (not filtered/validated)
				matcher := func(req *pingone.EnvironmentCreateRequest) bool {
					return req.Name == ""
				}
				mockCreateEnvironmentSetup(m, matcher, nil, 400, errors.New("name is required"))
			},
			wantErr:         true,
			wantErrContains: "name is required",
		},
		{
			name: "Whitespace-only name is passed through to API",
			input: environments.CreateEnvironmentInput{
				Name:    "   ", // Whitespace only
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				// Verify handler passes whitespace to API (not trimmed/validated)
				matcher := func(req *pingone.EnvironmentCreateRequest) bool {
					return req.Name == "   "
				}
				mockCreateEnvironmentSetup(m, matcher, nil, 400, errors.New("name cannot be whitespace"))
			},
			wantErr:         true,
			wantErrContains: "name cannot be whitespace",
		},
		{
			name: "Empty BillOfMaterials products array is passed through",
			input: environments.CreateEnvironmentInput{
				Name:    "Test Environment",
				Region:  pingone.ENVIRONMENTREGIONCODE_NA,
				License: *pingone.NewEnvironmentLicense(testLicenseID),
				BillOfMaterials: &pingone.EnvironmentBillOfMaterials{
					Products: []pingone.EnvironmentBillOfMaterialsProduct{}, // Empty array
				},
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				// Verify handler passes empty products array to API
				matcher := func(req *pingone.EnvironmentCreateRequest) bool {
					return req.BillOfMaterials != nil && len(req.BillOfMaterials.Products) == 0
				}
				mockCreateEnvironmentSetup(m, matcher, nil, 400, errors.New("at least one product is required"))
			},
			wantErr:         true,
			wantErrContains: "at least one product is required",
		},
		{
			name: "Empty description string is passed through and accepted",
			input: environments.CreateEnvironmentInput{
				Name:        "Test Environment",
				Region:      pingone.ENVIRONMENTREGIONCODE_NA,
				License:     *pingone.NewEnvironmentLicense(testLicenseID),
				Description: testutils.Pointer(""), // Empty description
			},
			setupMock: func(m *mockPingOneClientEnvironmentsWrapper) {
				createdEnvID := uuid.MustParse("550e8400-e29b-41d4-a716-446655441005")
				// Verify handler passes empty description to API
				matcher := func(req *pingone.EnvironmentCreateRequest) bool {
					return req.Description != nil && *req.Description == ""
				}
				mockCreateEnvironmentSetup(m, matcher,
					&pingone.EnvironmentResponse{
						Id:          createdEnvID,
						Name:        "Test Environment",
						Region:      pingone.ENVIRONMENTREGIONCODE_NA,
						Type:        pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
						Description: testutils.Pointer(""),
					},
					201,
					nil,
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			tt.setupMock(mockClient)
			handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())
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

func TestCreateEnvironmentHandler_APIErrors(t *testing.T) {
	tests := testutils.CommonAPIErrorTestCases()

	input := environments.CreateEnvironmentInput{
		Name:    "Test Environment",
		Region:  pingone.ENVIRONMENTREGIONCODE_NA,
		License: *pingone.NewEnvironmentLicense(testLicenseID),
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Setup
			mockClient := &mockPingOneClientEnvironmentsWrapper{}
			mockCreateEnvironmentSetup(mockClient, nil, nil, tt.StatusCode, tt.ApiError)
			handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializer())

			// Execute
			mcpResult, output, err := handler(context.Background(), &mcp.CallToolRequest{}, input)

			// Assert
			testutils.AssertHandlerError(t, err, mcpResult, output, tt.WantErrContains)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCreateEnvironmentHandler_GetAuthenticatedClientError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	clientFactoryErr := errors.New("failed to get authenticated client")
	handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, clientFactoryErr), testutils.MockContextInitializer())
	req := &mcp.CallToolRequest{}
	input := environments.CreateEnvironmentInput{
		Name:    "Test Environment",
		Region:  pingone.ENVIRONMENTREGIONCODE_NA,
		License: *pingone.NewEnvironmentLicense(testLicenseID),
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

func TestCreateEnvironmentHandler_InitializeAuthContextError(t *testing.T) {
	mockClient := &mockPingOneClientEnvironmentsWrapper{}
	initContextErr := errors.New("failed to initialize auth context")
	handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(mockClient, nil), testutils.MockContextInitializerWithError(initContextErr))
	req := &mcp.CallToolRequest{}
	input := environments.CreateEnvironmentInput{
		Name:    "Test Environment",
		Region:  pingone.ENVIRONMENTREGIONCODE_NA,
		License: *pingone.NewEnvironmentLicense(testLicenseID),
	}

	mcpResult, output, err := handler(context.Background(), req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize auth context")
	assert.Nil(t, mcpResult)
	assert.Nil(t, output)
}

// TestCreateEnvironmentHandler_RealClient tests the handler with a real PingOne client.
// This test is skipped by default as it creates actual resources.
func TestCreateEnvironmentHandler_RealClient(t *testing.T) {
	//TODO enable test when we have can run against a real P1 client
	t.Skip("Skipping TestCreateEnvironmentHandler_RealClient since it relies on real P1 client and creates actual resources")

	var emptyToken string
	client, err := sdk.NewDefaultClientFactory(testutils.TestServerVersion).NewClient(emptyToken)
	require.NoError(t, err, "Failed to create PingOne client - check your credentials")

	// Create the client wrapper
	clientWrapper := environments.NewPingOneClientEnvironmentsWrapper(client)

	// Note: You would need a valid license ID from your PingOne environment
	licenseUUID := uuid.MustParse("00000000-0000-0000-0000-000000000000") // Replace with valid UUID
	args := environments.CreateEnvironmentInput{
		Name:        "Test Environment from SDK",
		Region:      pingone.ENVIRONMENTREGIONCODE_NA,
		License:     *pingone.NewEnvironmentLicense(licenseUUID),
		Description: testutils.Pointer("Created by automated test"),
	}

	req := &mcp.CallToolRequest{}
	handler := environments.CreateEnvironmentHandler(NewMockPingOneClientEnvironmentsWrapperFactory(clientWrapper, nil), testutils.MockContextInitializer())

	mcpResult, structuredResponse, err := handler(context.Background(), req, args)

	require.NoError(t, err, "Handler should not return error with valid credentials")
	assert.Nil(t, mcpResult, "MCP result should be nil for successful operations")
	require.NotNil(t, structuredResponse, "Structured response should not be nil")

	assert.NotEqual(t, uuid.Nil, structuredResponse.Environment.Id, "Created environment should have an ID")
	assert.Equal(t, args.Name, structuredResponse.Environment.Name)
	assert.Equal(t, pingone.ENVIRONMENTTYPEVALUE_SANDBOX, structuredResponse.Environment.Type)

	// Clean up: Delete the created environment
	// Note: You would need to implement deletion or manually clean up
}
