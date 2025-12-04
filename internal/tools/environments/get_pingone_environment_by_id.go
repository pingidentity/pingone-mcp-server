// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var GetEnvironmentByIdDef = types.ToolDefinition{
	IsReadOnly: true,
	McpTool: &mcp.Tool{
		Name:         "get_environment_by_id",
		Title:        "Get PingOne Environment by ID",
		Description:  "Retrieve an environment's full configuration by ID. Use 'list_environments' first if you need to find the environment ID. Call this before 'update_environment_by_id' to get current configuration.",
		InputSchema:  schema.MustGenerateSchema[GetEnvironmentByIdInput](),
		OutputSchema: schema.MustGenerateSchema[GetEnvironmentByIdOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

// GetEnvironmentByIdInput defines the input parameters for retrieving an environment by ID
type GetEnvironmentByIdInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. UUID format (e.g., '123e4567-e89b-12d3-a456-426614174000')."`
}

// GetEnvironmentByIdOutput represents the result of retrieving an environment
type GetEnvironmentByIdOutput struct {
	Environment pingone.EnvironmentResponse `json:"environment" jsonschema:"The environment details including ID, name, type, region, and metadata"`
}

// GetEnvironmentByIdHandler retrieves a PingOne environment by ID using the provided client
func GetEnvironmentByIdHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetEnvironmentByIdInput,
) (
	*mcp.CallToolResult,
	*GetEnvironmentByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEnvironmentByIdInput) (*mcp.CallToolResult, *GetEnvironmentByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetEnvironmentByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving environment", slog.String("environmentId", input.EnvironmentId.String()))

		// Call the API to retrieve the environment
		environment, httpResponse, err := client.GetEnvironmentById(ctx, input.EnvironmentId)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if environment == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no environment data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Environment retrieved successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("environmentName", environment.Name),
		)

		// Filter out _links field from response
		environment.Links = nil

		result := &GetEnvironmentByIdOutput{
			Environment: *environment,
		}

		return nil, result, nil
	}
}
