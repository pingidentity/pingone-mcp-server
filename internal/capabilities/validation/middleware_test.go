// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
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
			McpTool: &mcp.Tool{
				Name:        "list_populations",
				Description: "List populations",
				Annotations: &mcp.ToolAnnotations{
					ReadOnlyHint: true,
				},
			},
		},
		{
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
			require.NotNil(t, result)
			assert.Equal(t, envId, *result)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			result, found, err := extractEnvironmentId(argsJSON)
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Nil(t, result)
		})
	}
}

func TestExtractEnvironmentId_InvalidUUID(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "invalid UUID string",
			args: map[string]any{
				"environmentId": "not-a-uuid",
			},
		},
		{
			name: "empty string",
			args: map[string]any{
				"environmentId": "",
			},
		},
		{
			name: "malformed UUID",
			args: map[string]any{
				"environmentId": "123-456-789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			result, found, err := extractEnvironmentId(argsJSON)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid environmentId format")
			assert.True(t, found)
			assert.Nil(t, result)
		})
	}
}

func TestExtractEnvironmentId_InvalidJSON(t *testing.T) {
	invalidJSON := json.RawMessage(`{invalid json}`)

	result, found, err := extractEnvironmentId(invalidJSON)
	assert.Error(t, err)
	assert.False(t, found)
	assert.Nil(t, result)
}

// Test shouldSkipEnvironmentValidation

func TestShouldSkipEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name          string
		toolDef       *types.ToolDefinition
		operationType OperationType
		expected      bool
	}{
		{
			name:          "nil tool definition",
			toolDef:       nil,
			operationType: OperationTypeRead,
			expected:      true,
		},
		{
			name: "ProductionEnvironmentNotApplicable true - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					ProductionEnvironmentNotApplicable: true,
				},
				McpTool: &mcp.Tool{Name: "no_env_tool"},
			},
			operationType: OperationTypeRead,
			expected:      true,
		},
		{
			name: "ProductionEnvironmentNotApplicable true - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					ProductionEnvironmentNotApplicable: true,
				},
				McpTool: &mcp.Tool{Name: "no_env_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      true,
		},
		{
			name: "AllowProductionEnvironmentWrite true - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentWrite: true,
				},
				McpTool: &mcp.Tool{Name: "trusted_write_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      true,
		},
		{
			name: "AllowProductionEnvironmentWrite true - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentWrite: true,
				},
				McpTool: &mcp.Tool{Name: "trusted_write_tool"},
			},
			operationType: OperationTypeRead,
			expected:      false, // Write permission doesn't grant read permission
		},
		{
			name: "AllowProductionEnvironmentRead true - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead: true,
				},
				McpTool: &mcp.Tool{Name: "trusted_read_tool"},
			},
			operationType: OperationTypeRead,
			expected:      true,
		},
		{
			name: "AllowProductionEnvironmentRead true - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead: true,
				},
				McpTool: &mcp.Tool{Name: "trusted_read_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      false, // Read permission doesn't grant write permission
		},
		{
			name: "Both read and write permissions - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead:  true,
					AllowProductionEnvironmentWrite: true,
				},
				McpTool: &mcp.Tool{Name: "full_access_tool"},
			},
			operationType: OperationTypeRead,
			expected:      true,
		},
		{
			name: "Both read and write permissions - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead:  true,
					AllowProductionEnvironmentWrite: true,
				},
				McpTool: &mcp.Tool{Name: "full_access_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      true,
		},
		{
			name: "No permissions - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead:  false,
					AllowProductionEnvironmentWrite: false,
				},
				McpTool: &mcp.Tool{Name: "restricted_tool"},
			},
			operationType: OperationTypeRead,
			expected:      false,
		},
		{
			name: "No permissions - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					AllowProductionEnvironmentRead:  false,
					AllowProductionEnvironmentWrite: false,
				},
				McpTool: &mcp.Tool{Name: "restricted_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      false,
		},
		{
			name: "Nil validation policy - READ operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: nil,
				McpTool:          &mcp.Tool{Name: "no_policy_tool"},
			},
			operationType: OperationTypeRead,
			expected:      false,
		},
		{
			name: "Nil validation policy - WRITE operation",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: nil,
				McpTool:          &mcp.Tool{Name: "no_policy_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      false,
		},
		{
			name: "ProductionEnvironmentNotApplicable overrides other settings - READ",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					ProductionEnvironmentNotApplicable: true,
					AllowProductionEnvironmentRead:     false,
					AllowProductionEnvironmentWrite:    false,
				},
				McpTool: &mcp.Tool{Name: "not_applicable_tool"},
			},
			operationType: OperationTypeRead,
			expected:      true, // ProductionEnvironmentNotApplicable takes precedence
		},
		{
			name: "ProductionEnvironmentNotApplicable overrides other settings - WRITE",
			toolDef: &types.ToolDefinition{
				ValidationPolicy: &types.ToolValidationPolicy{
					ProductionEnvironmentNotApplicable: true,
					AllowProductionEnvironmentRead:     false,
					AllowProductionEnvironmentWrite:    false,
				},
				McpTool: &mcp.Tool{Name: "not_applicable_tool"},
			},
			operationType: OperationTypeWrite,
			expected:      true, // ProductionEnvironmentNotApplicable takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipEnvironmentValidation(tt.toolDef, tt.operationType)
			assert.Equal(t, tt.expected, result)
		})
	}
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
				McpTool: &mcp.Tool{
					Name: "list_tool",
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: true,
					},
				},
			},
			expected: OperationTypeRead,
		},
		{
			name: "write tool",
			toolDef: &types.ToolDefinition{
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
		McpTool: &mcp.Tool{
			Name: "list_environments",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
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
		McpTool: &mcp.Tool{
			Name: "list_populations",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
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

func TestEnvironmentValidationMiddleware_WriteOperation_Success(t *testing.T) {
	envId := uuid.New()
	mockVal := new(mockValidatorMiddleware)
	mockReg := new(mockToolRegistry)

	toolDef := &types.ToolDefinition{
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
		McpTool: &mcp.Tool{
			Name: "test_tool",
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint: true,
			},
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
