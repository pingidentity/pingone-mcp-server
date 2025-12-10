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

var GetEnvironmentDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "get_environment",
		Title:        "Get PingOne Environment by ID",
		Description:  "Retrieve an environment's full configuration by ID. Use 'list_environments' first if you need to find the environment ID. Call this before 'update_environment' to get current configuration.",
		InputSchema:  schema.MustGenerateSchema[GetEnvironmentInput](),
		OutputSchema: schema.MustGenerateSchema[GetEnvironmentOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

// GetEnvironmentInput defines the input parameters for retrieving an environment by ID
type GetEnvironmentInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. UUID format (e.g., '123e4567-e89b-12d3-a456-426614174000')."`
}

// GetEnvironmentOutput represents the result of retrieving an environment
type GetEnvironmentOutput struct {
	Environment pingone.EnvironmentResponse `json:"environment" jsonschema:"The environment details including ID, name, type, region, and metadata"`
}

// GetEnvironmentHandler retrieves a PingOne environment by ID using the provided client
func GetEnvironmentHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetEnvironmentInput,
) (
	*mcp.CallToolResult,
	*GetEnvironmentOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEnvironmentInput) (*mcp.CallToolResult, *GetEnvironmentOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetEnvironmentDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving environment", slog.String("environmentId", input.EnvironmentId.String()))

		// Call the API to retrieve the environment
		environment, httpResponse, err := client.GetEnvironment(ctx, input.EnvironmentId)
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

		result := &GetEnvironmentOutput{
			Environment: *environment,
		}

		return nil, result, nil
	}
}
