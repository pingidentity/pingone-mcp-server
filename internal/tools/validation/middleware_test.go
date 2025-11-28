// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockValidatorMiddleware struct {
	mock.Mock
}

func (m *mockValidatorMiddleware) ValidateEnvironment(ctx context.Context, environmentId uuid.UUID, operationType OperationType) error {
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

// Helper function to create CallToolRequest with proper arguments
func createCallToolRequest(toolName string, args map[string]any) *mcp.CallToolRequest {
	argsJSON, _ := json.Marshal(args)
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      toolName,
			Arguments: argsJSON,
		},
	}
}

// Test ToolRegistry

func TestNewToolRegistry(t *testing.T) {
	tools := []types.ToolDefinition{
		{
			IsReadOnly: true,
			McpTool: &mcp.Tool{
				Name:        "list_populations",
				Description: "List populations",
			},
		},
		{
			IsReadOnly: false,
			McpTool: &mcp.Tool{
				Name:        "create_population",
				Description: "Create population",
			},
		},
	}

	registry := NewToolRegistry(tools)

	assert.NotNil(t, registry)
	assert.Len(t, registry.tools, 2)
	assert.NotNil(t, registry.GetTool("list_populations"))
	assert.NotNil(t, registry.GetTool("create_population"))
	assert.Nil(t, registry.GetTool("nonexistent_tool"))
}

func TestToolRegistry_GetTool(t *testing.T) {
	readTool := types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name: "read_tool",
		},
	}

	registry := NewToolRegistry([]types.ToolDefinition{readTool})

	result := registry.GetTool("read_tool")
	require.NotNil(t, result)
	assert.Equal(t, "read_tool", result.McpTool.Name)
	assert.True(t, result.IsReadOnly)

	notFound := registry.GetTool("missing_tool")
	assert.Nil(t, notFound)
}

// Test extractEnvironmentId

func TestExtractEnvironmentId_Success(t *testing.T) {
	envId := uuid.New()

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "string UUID",
			args: map[string]any{
				"environmentId": envId.String(),
			},
		},
		{
			name: "with other fields",
			args: map[string]any{
				"environmentId": envId.String(),
				"name":          "Test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			result, found, err := extractEnvironmentId(argsJSON)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, envId, result)
		})
	}
}

func TestExtractEnvironmentId_NotFound(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "empty args",
			args: map[string]any{},
		},
		{
			name: "missing environmentId",
			args: map[string]any{
				"otherId": "some-value",
			},
		},
		{
			name: "invalid UUID string",
			args: map[string]any{
				"environmentId": "not-a-uuid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			result, found, err := extractEnvironmentId(argsJSON)
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Equal(t, uuid.UUID{}, result)
		})
	}
}

func TestExtractEnvironmentId_InvalidJSON(t *testing.T) {
	invalidJSON := json.RawMessage(`{invalid json}`)

	result, found, err := extractEnvironmentId(invalidJSON)
	assert.Error(t, err)
	assert.False(t, found)
	assert.Equal(t, uuid.UUID{}, result)
}

// Test determineOperationType

func TestDetermineOperationType(t *testing.T) {
	tests := []struct {
		name     string
		toolDef  *types.ToolDefinition
		expected OperationType
	}{
		{
			name:     "nil tool definition",
			toolDef:  nil,
			expected: OperationTypeRead,
		},
		{
			name: "read-only tool",
			toolDef: &types.ToolDefinition{
				IsReadOnly: true,
				McpTool:    &mcp.Tool{Name: "list_tool"},
			},
			expected: OperationTypeRead,
		},
		{
			name: "write tool",
			toolDef: &types.ToolDefinition{
				IsReadOnly: false,
				McpTool: &mcp.Tool{
					Name:        "create_tool",
					Annotations: &mcp.ToolAnnotations{},
				},
			},
			expected: OperationTypeWrite,
		},
		{
			name: "read-only hint annotation",
			toolDef: &types.ToolDefinition{
				IsReadOnly: false,
				McpTool: &mcp.Tool{
					Name: "annotated_read_tool",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: true,
					},
				},
			},
			expected: OperationTypeRead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineOperationType(tt.toolDef)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test Middleware

func TestEnvironmentValidationMiddleware_NonToolCall(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	// Test with initialize method (not tools/call)
	req := &mcp.InitializeRequest{}
	_, err := handler(context.Background(), "initialize", req)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertNotCalled(t, "GetTool")
}

func TestEnvironmentValidationMiddleware_InvalidRequestType(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	// Test with wrong request type for tools/call method
	req := &mcp.InitializeRequest{}
	result, err := handler(context.Background(), "tools/call", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Contains(t, err.Error(), "invalid tool call request")
	assert.Nil(t, result)
	assert.False(t, nextCalled)
	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertNotCalled(t, "GetTool")
}

func TestEnvironmentValidationMiddleware_ToolWithoutEnvironmentId(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	// Mock tool definition that doesn't skip validation
	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name: "list_environments",
		},
	}
	mockReg.On("GetTool", "list_environments").Return(toolDef)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	// Tool call without environmentId parameter should fail as validation is mandatory
	req := createCallToolRequest("list_environments", map[string]any{
		"filter": "name sw \"Test\"",
	})

	result, err := handler(context.Background(), "tools/call", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Nil(t, result)
	assert.False(t, nextCalled, "next handler should not be called when validation fails")
	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_ReadOperation_Success(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name:        "list_populations",
			Annotations: &mcp.ToolAnnotations{},
		},
	}

	mockReg.On("GetTool", "list_populations").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, OperationTypeRead).Return(nil)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	req := createCallToolRequest("list_populations", map[string]any{
		"environmentId": envId.String(),
	})

	_, err := handler(context.Background(), "tools/call", req)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_WriteOperation_Production_Blocked(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: false,
		McpTool: &mcp.Tool{
			Name:        "create_population",
			Annotations: &mcp.ToolAnnotations{},
		},
	}

	mockReg.On("GetTool", "create_population").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, OperationTypeWrite).Return(nil)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	req := createCallToolRequest("create_population", map[string]any{
		"environmentId": envId.String(),
		"name":          "Test Population",
	})

	_, err := handler(context.Background(), "tools/call", req)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_ValidationError(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: false,
		McpTool: &mcp.Tool{
			Name:        "create_population",
			Annotations: &mcp.ToolAnnotations{},
		},
	}

	validationErr := errors.New("environment not found")

	mockReg.On("GetTool", "create_population").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, OperationTypeWrite).Return(validationErr)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	req := createCallToolRequest("create_population", map[string]any{
		"environmentId": envId.String(),
	})

	result, err := handler(context.Background(), "tools/call", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Contains(t, err.Error(), "environment not found")
	assert.Nil(t, result)
	assert.False(t, nextCalled, "next handler should not be called when validation fails")
	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_ProductionProtection(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
		IsReadOnly: false,
		McpTool: &mcp.Tool{
			Name: "delete_population",
			// No tool annotations so we can test nil pointer behaviour
		},
	}

	productionErr := errors.New("to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed")

	mockReg.On("GetTool", "delete_population").Return(toolDef)
	mockVal.On("ValidateEnvironment", mock.Anything, envId, OperationTypeWrite).Return(productionErr)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	req := createCallToolRequest("delete_population", map[string]any{
		"environmentId": envId.String(),
		"populationId":  uuid.New().String(),
	})

	result, err := handler(context.Background(), "tools/call", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Contains(t, err.Error(), "PRODUCTION environments")
	assert.Nil(t, result)
	assert.False(t, nextCalled)
	mockVal.AssertExpectations(t)
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_UnknownTool(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	// Tool not found in registry - when nil, validation is skipped
	mockReg.On("GetTool", "unknown_tool").Return((*types.ToolDefinition)(nil))

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	req := createCallToolRequest("unknown_tool", map[string]any{
		"environmentId": envId.String(),
	})

	_, err := handler(context.Background(), "tools/call", req)

	// When tool definition is nil, validation is skipped and request proceeds
	assert.NoError(t, err)
	assert.True(t, nextCalled)
	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertExpectations(t)
}

func TestEnvironmentValidationMiddleware_InvalidJSON(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	// Mock tool definition
	toolDef := &types.ToolDefinition{
		IsReadOnly: true,
		McpTool: &mcp.Tool{
			Name: "test_tool",
		},
	}
	mockReg.On("GetTool", "test_tool").Return(toolDef)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return nil, nil
	}

	handler := middleware.Handler(next)

	// Create request with invalid JSON arguments
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test_tool",
			Arguments: json.RawMessage(`{invalid json}`),
		},
	}

	result, err := handler(context.Background(), "tools/call", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment validation failed")
	assert.Nil(t, result)
	assert.False(t, nextCalled)
	mockVal.AssertNotCalled(t, "ValidateEnvironment")
	mockReg.AssertExpectations(t)
}

func TestNewEnvironmentValidationMiddleware(t *testing.T) {
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	middleware := NewEnvironmentValidationMiddleware(mockVal, mockReg)

	assert.NotNil(t, middleware)
	assert.Equal(t, mockVal, middleware.validator)
	assert.Equal(t, mockReg, middleware.toolRegistry)
}
