// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

// ToolRegistry provides access to tool definitions for middleware.
// This allows the middleware to determine tool characteristics like read-only status.
type ToolRegistry interface {
	// GetTool returns the tool definition for the given tool name.
	// Returns nil if the tool is not found.
	GetTool(name string) *types.ToolDefinition
}

// DefaultToolRegistry implements ToolRegistry using a map for fast lookups.
type DefaultToolRegistry struct {
	tools map[string]*types.ToolDefinition
}

// NewToolRegistry creates a new tool registry from a slice of tool definitions.
// The registry builds an index for fast lookups by tool name.
func NewToolRegistry(tools []types.ToolDefinition) *DefaultToolRegistry {
	registry := &DefaultToolRegistry{
		tools: make(map[string]*types.ToolDefinition),
	}
	for i := range tools {
		registry.tools[tools[i].McpTool.Name] = &tools[i]
	}
	return registry
}

// GetTool returns the tool definition for the given tool name.
// Returns nil if the tool is not found in the registry.
func (r *DefaultToolRegistry) GetTool(name string) *types.ToolDefinition {
	return r.tools[name]
}

// EnvironmentValidationMiddleware validates environment access for all tool calls.
// It intercepts tool call requests, extracts the environmentId parameter, and validates:
// 1. Environment exists and is accessible
// 2. For write operations, environment is not PRODUCTION type
//
// This middleware should be added to the MCP server via AddReceivingMiddleware.
// Tools without an environmentId parameter are not validated (e.g., list_environments).
type EnvironmentValidationMiddleware struct {
	validator    EnvironmentValidator
	toolRegistry ToolRegistry
}

// NewEnvironmentValidationMiddleware creates middleware with validator and tool registry.
// The validator is used to check environment access and type.
// The toolRegistry is used to determine if a tool is read-only or performs write operations.
func NewEnvironmentValidationMiddleware(
	validator EnvironmentValidator,
	toolRegistry ToolRegistry,
) *EnvironmentValidationMiddleware {
	return &EnvironmentValidationMiddleware{
		validator:    validator,
		toolRegistry: toolRegistry,
	}
}

// Handler implements the middleware pattern by returning a MethodHandler that wraps the next handler.
// This handler intercepts all MCP method calls and validates tool calls that operate on environments.
func (m *EnvironmentValidationMiddleware) Handler(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Only validate tool calls, not other MCP methods (initialize, list_tools, etc.)
		if method != "tools/call" {
			return next(ctx, method, req)
		}

		// Extract tool call details
		callToolReq, ok := req.(*mcp.CallToolRequest)
		if !ok {
			// Should never happen for tools/call method, but validation is mandatory, so we fail the call
			return nil, fmt.Errorf("environment validation failed: %w", fmt.Errorf("invalid tool call request"))
		}

		toolName := callToolReq.Params.Name

		// Lookup tool definition
		toolDef := m.toolRegistry.GetTool(toolName)

		// Determine operation type from tool definition
		operationType := determineOperationType(toolDef)

		// Check if validation should be skipped based on tool policy and operation type
		if shouldSkipEnvironmentValidation(toolDef, operationType) {
			// Skip environment validation for this tool as per its validation policy
			logger.FromContext(ctx).Debug("Skipping environment validation for tool",
				slog.String("tool", toolName),
				slog.String("operationType", string(operationType)),
				slog.String("reason", "validation policy allows operation"))
			return next(ctx, method, req)
		}

		// Extract environmentId from parameters
		environmentId, hasEnvId, err := extractEnvironmentId(callToolReq.Params.Arguments)
		if err != nil {
			// Failed to parse arguments, validation is mandatory, so we fail the call
			logger.FromContext(ctx).Error("Failed to parse tool arguments",
				slog.String("tool", toolName),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("environment validation failed: %w", err)
		}
		if !hasEnvId {
			// Tool doesn't use environmentId, validation is mandatory, so we fail the call
			logger.FromContext(ctx).Error("Tool requires environment validation, but no environmentId was found to validate",
				slog.String("tool", toolName))
			return nil, fmt.Errorf("environment validation failed: %w", err)
		}
		if environmentId == nil {
			// Tool does use environmentId, but for whatever reason wasn't returned, validation is mandatory, so we fail the call
			logger.FromContext(ctx).Error("Tool requires environment validation, environmentId is expected, but no environmentId was found",
				slog.String("tool", toolName))
			return nil, fmt.Errorf("environment validation failed: %w", err)
		}

		logger.FromContext(ctx).Debug("Validating environment for tool",
			slog.String("tool", toolName),
			slog.String("environmentId", environmentId.String()),
			slog.String("operationType", string(operationType)))

		// Validate environment
		if err := m.validator.ValidateEnvironment(ctx, *environmentId, operationType); err != nil {
			logger.FromContext(ctx).Error("Environment validation failed",
				slog.String("tool", toolName),
				slog.String("environmentId", environmentId.String()),
				slog.String("operationType", string(operationType)),
				slog.String("error", err.Error()))
			return nil, fmt.Errorf("environment validation failed: %w", err)
		}

		logger.FromContext(ctx).Debug("Environment validation passed",
			slog.String("tool", toolName),
			slog.String("environmentId", environmentId.String()))

		// Validation passed, continue to tool handler
		return next(ctx, method, req)
	}
}

// extractEnvironmentId extracts the environmentId from tool call arguments.
// Returns the UUID, true if found, and any parsing error.
// Supports both string and direct UUID representations in JSON.
func extractEnvironmentId(argsJSON json.RawMessage) (*uuid.UUID, bool, error) {
	// Parse JSON into map
	var args map[string]any
	if err := json.Unmarshal(argsJSON, &args); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	// Try direct field access with key "environmentId"
	if envIdRaw, ok := args["environmentId"]; ok {
		// Handle string representation
		if envIdStr, ok := envIdRaw.(string); ok {
			parsed, err := uuid.Parse(envIdStr)
			if err == nil {
				return &parsed, true, nil
			} else {
				return nil, true, fmt.Errorf("invalid environmentId format: %w", err)
			}
		}
		// Handle direct UUID type (less common but possible)
		if envIdUUID, ok := envIdRaw.(uuid.UUID); ok {
			return &envIdUUID, true, nil
		}
	}

	// Environment ID not found or invalid format
	return nil, false, nil
}

// determineOperationType determines if a tool performs read or write operations.
// Read-only tools can operate on PRODUCTION environments.
// Write tools are blocked from operating on PRODUCTION environments.
func determineOperationType(toolDef *types.ToolDefinition) OperationType {
	if toolDef == nil || toolDef.IsReadOnly || (toolDef.McpTool.Annotations != nil && toolDef.McpTool.Annotations.ReadOnlyHint) {
		return OperationTypeRead
	}
	return OperationTypeWrite
}

// shouldSkipEnvironmentValidation determines if environment validation should be skipped for a tool.
// Returns true if the tool definition is nil or if the tool's validation policy allows the operation.
// The logic follows this priority:
// 1. If toolDef is nil, skip validation (unknown tool)
// 2. If ProductionEnvironmentNotApplicable is true, skip validation (tool doesn't use environmentId)
// 3. If operation is WRITE and AllowProductionEnvironmentWrite is true, skip validation
// 4. If operation is READ and AllowProductionEnvironmentRead is true, skip validation
// 5. Otherwise, perform validation (default restrictive behavior)
func shouldSkipEnvironmentValidation(toolDef *types.ToolDefinition, operationType OperationType) bool {
	if toolDef == nil {
		return true
	}

	if toolDef.ValidationPolicy != nil {
		// If tool doesn't operate on environments, skip validation entirely
		if toolDef.ValidationPolicy.ProductionEnvironmentNotApplicable {
			return true
		}

		// Check operation-specific permissions
		if operationType == OperationTypeWrite && toolDef.ValidationPolicy.AllowProductionEnvironmentWrite {
			return true
		}

		if operationType == OperationTypeRead && toolDef.ValidationPolicy.AllowProductionEnvironmentRead {
			return true
		}
	}

	// Default: require validation
	return false
}
