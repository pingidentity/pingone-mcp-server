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

var GetEnvironmentServicesDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "get_environment_services",
		Title:        "Get PingOne Environment Services by ID",
		Description:  "Retrieve all the services assigned to a specified PingOne environment, by the environment's unique ID.",
		InputSchema:  schema.MustGenerateSchema[GetEnvironmentServicesInput](),
		OutputSchema: schema.MustGenerateSchema[GetEnvironmentServicesOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

// GetEnvironmentServicesInput defines the input parameters for retrieving environment services by ID
type GetEnvironmentServicesInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
}

// GetEnvironmentServicesOutput represents the result of retrieving environment services
type GetEnvironmentServicesOutput struct {
	Services pingone.EnvironmentBillOfMaterialsResponse `json:"services" jsonschema:"The bill of materials for the environment, including products and solution type"`
}

// GetEnvironmentServicesHandler retrieves PingOne environment services by ID using the provided client
func GetEnvironmentServicesHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetEnvironmentServicesInput,
) (
	*mcp.CallToolResult,
	*GetEnvironmentServicesOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetEnvironmentServicesInput) (*mcp.CallToolResult, *GetEnvironmentServicesOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetEnvironmentServicesDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentServicesDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetEnvironmentServicesDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving environment services",
			slog.String("environmentId", input.EnvironmentId.String()))

		// Call the API to retrieve the environment services
		services, httpResponse, err := client.GetEnvironmentServices(ctx, input.EnvironmentId)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if services == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no services data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Environment services retrieved successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(services.Products)))

		// Filter out _links field from response
		services.Links = nil

		result := &GetEnvironmentServicesOutput{
			Services: *services,
		}

		return nil, result, nil
	}
}
