// Copyright © 2025 Ping Identity Corporation

package validation_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcptestutils "github.com/pingidentity/pingone-mcp-server/internal/testutils/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for integration testing

type mockValidatorMiddleware struct {
	mock.Mock
}

func (m *mockValidatorMiddleware) ValidateEnvironment(ctx context.Context, environmentId uuid.UUID, operationType validation.OperationType) error {
	args := m.Called(ctx, environmentId, operationType)
	return args.Error(0)
}

type mockToolRegistry struct {
	mock.Mock
}

func (m *mockToolRegistry) GetTool(name string) *types.ToolDefinition {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*types.ToolDefinition)
}

// Test input/output types for mock tools

type testToolInput struct {
	EnvironmentId uuid.UUID `json:"environmentId"`
	Name          string    `json:"name,omitempty"`
}

type testToolOutput struct {
	Message string `json:"message"`
	Id      string `json:"id,omitempty"`
}

// TestEnvironmentValidationMiddleware_ReadOperation_OverMcp tests read operations through MCP protocol.
func TestEnvironmentValidationMiddleware_ReadOperation_OverMcp(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name:         "list_test_resources",
			Description:  "List test resources in an environment",
			InputSchema:  schema.MustGenerateSchema[testToolInput](),
			OutputSchema: schema.MustGenerateSchema[testToolOutput](),
		},
	}

	mockReg.On("GetTool", "list_test_resources").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, validation.OperationTypeRead).Return(nil)

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a simple handler that returns success
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, *testToolOutput, error) {
		return nil, &testToolOutput{
			Message: "Read operation succeeded",
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP
	input := testToolInput{
		EnvironmentId: envId,
	}

	output, err := mcptestutils.CallToolOverMcp(t, server, "list_test_resources", input)

	// Assert success
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.IsError)
	require.NotNil(t, output.StructuredContent)

	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

// TestEnvironmentValidationMiddleware_WriteOperation_OverMcp tests write operations through MCP protocol.
func TestEnvironmentValidationMiddleware_WriteOperation_OverMcp(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: false,
		McpTool: &mcp.Tool{
			Name:         "create_test_resource",
			Description:  "Create a test resource in an environment",
			InputSchema:  schema.MustGenerateSchema[testToolInput](),
			OutputSchema: schema.MustGenerateSchema[testToolOutput](),
		},
	}

	mockReg.On("GetTool", "create_test_resource").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, validation.OperationTypeWrite).Return(nil)

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a simple handler that returns success
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, *testToolOutput, error) {
		return nil, &testToolOutput{
			Message: "Write operation succeeded",
			Id:      uuid.New().String(),
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP
	input := testToolInput{
		EnvironmentId: envId,
		Name:          "Test Resource",
	}

	output, err := mcptestutils.CallToolOverMcp(t, server, "create_test_resource", input)

	// Assert success
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.IsError)
	require.NotNil(t, output.StructuredContent)

	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

// TestEnvironmentValidationMiddleware_ValidationFailure_OverMcp tests validation failure scenarios through MCP protocol.
func TestEnvironmentValidationMiddleware_ValidationFailure_OverMcp(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: false,
		McpTool: &mcp.Tool{
			Name:         "create_test_resource",
			Description:  "Create a test resource in an environment",
			InputSchema:  schema.MustGenerateSchema[testToolInput](),
			OutputSchema: schema.MustGenerateSchema[testToolOutput](),
		},
	}

	validationErr := errors.New("to safeguard against unintended or breaking changes, this write operation is not allowed against PRODUCTION environments")

	mockReg.On("GetTool", "create_test_resource").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, validation.OperationTypeWrite).Return(validationErr)

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a handler that should not be called
	handlerCalled := false
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, *testToolOutput, error) {
		handlerCalled = true
		return nil, &testToolOutput{
			Message: "This should not be reached",
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP
	input := testToolInput{
		EnvironmentId: envId,
		Name:          "Test Resource",
	}

	output, err := mcptestutils.CallToolOverMcp(t, server, "create_test_resource", input)

	// Assert error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Contains(t, err.Error(), "PRODUCTION environments")
	assert.Nil(t, output)
	assert.False(t, handlerCalled, "Handler should not be called when validation fails")

	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

// TestEnvironmentValidationMiddleware_SkipValidation_ProductionNotApplicable_OverMcp tests skipping validation for tools that don't use environmentId.
func TestEnvironmentValidationMiddleware_SkipValidation_ProductionNotApplicable_OverMcp(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	// Input type without environmentId
	type listInput struct{}
	type listOutput struct {
		Count int `json:"count"`
	}

	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		ValidationPolicy: &types.ToolValidationPolicy{
			ProductionEnvironmentNotApplicable: true,
		},
		McpTool: &mcp.Tool{
			Name:         "list_all_resources",
			Description:  "List all resources across all environments",
			InputSchema:  schema.MustGenerateSchema[listInput](),
			OutputSchema: schema.MustGenerateSchema[listOutput](),
		},
	}

	mockReg.On("GetTool", "list_all_resources").Return(toolDef)
	// Validator should not be called since ProductionEnvironmentNotApplicable is true

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a simple handler that returns success
	handlerCalled := false
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input listInput) (*mcp.CallToolResult, *listOutput, error) {
		handlerCalled = true
		return nil, &listOutput{
			Count: 5,
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP
	input := listInput{}

	output, err := mcptestutils.CallToolOverMcp(t, server, "list_all_resources", input)

	// Assert success - validation should be skipped
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.IsError)
	assert.True(t, handlerCalled, "Handler should be called when validation is skipped")

	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertExpectations(t)
}

// TestEnvironmentValidationMiddleware_InvalidEnvironmentId_OverMcp tests invalid environment ID handling.
func TestEnvironmentValidationMiddleware_InvalidEnvironmentId_OverMcp(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name:         "list_test_resources",
			Description:  "List test resources",
			InputSchema:  schema.MustGenerateSchema[testToolInput](),
			OutputSchema: schema.MustGenerateSchema[testToolOutput](),
		},
	}

	mockReg.On("GetTool", "list_test_resources").Return(toolDef)
	// Validator should not be called for invalid UUID

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a handler that should not be called
	handlerCalled := false
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, *testToolOutput, error) {
		handlerCalled = true
		return nil, &testToolOutput{
			Message: "This should not be reached",
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP with invalid UUID string
	input := map[string]interface{}{
		"environmentId": "not-a-valid-uuid",
	}

	output, err := mcptestutils.CallToolOverMcp(t, server, "list_test_resources", input)

	// Assert error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Contains(t, err.Error(), "invalid environmentId format")
	assert.Nil(t, output)
	assert.False(t, handlerCalled, "Handler should not be called when validation fails")

	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertExpectations(t)
}

// TestEnvironmentValidationMiddleware_WithStructuredContent_OverMcp tests validation with structured content response.
func TestEnvironmentValidationMiddleware_WithStructuredContent_OverMcp(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	type resourceListOutput struct {
		Resources []map[string]interface{} `json:"resources"`
	}

	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name:         "list_test_resources",
			Description:  "List test resources in an environment",
			InputSchema:  schema.MustGenerateSchema[testToolInput](),
			OutputSchema: schema.MustGenerateSchema[resourceListOutput](),
		},
	}

	mockReg.On("GetTool", "list_test_resources").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, validation.OperationTypeRead).Return(nil)

	middleware := validation.NewEnvironmentValidationMiddleware(mockVal, mockReg)

	// Create a handler that returns structured content
	successHandler := func(ctx context.Context, req *mcp.CallToolRequest, input testToolInput) (*mcp.CallToolResult, *resourceListOutput, error) {
		return nil, &resourceListOutput{
			Resources: []map[string]interface{}{
				{
					"id":   uuid.New().String(),
					"name": "Resource 1",
				},
				{
					"id":   uuid.New().String(),
					"name": "Resource 2",
				},
			},
		}, nil
	}

	// Create MCP server and add tool with middleware
	server := mcptestutils.TestMcpServer(t)
	server.AddReceivingMiddleware(middleware.Handler)
	mcp.AddTool(server, toolDef.McpTool, successHandler)

	// Execute tool call over MCP
	input := testToolInput{
		EnvironmentId: envId,
	}

	output, err := mcptestutils.CallToolOverMcp(t, server, "list_test_resources", input)

	// Assert success with structured content
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.IsError)
	require.NotNil(t, output.StructuredContent)

	// Verify structured content can be unmarshaled
	jsonBytes, jsonErr := json.Marshal(output.StructuredContent)
	require.NoError(t, jsonErr)

	var structuredData resourceListOutput
	unmarshalErr := json.Unmarshal(jsonBytes, &structuredData)
	require.NoError(t, unmarshalErr)
	assert.Len(t, structuredData.Resources, 2)

	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}
