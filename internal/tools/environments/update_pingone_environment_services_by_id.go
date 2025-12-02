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

var UpdateEnvironmentServicesByIdDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:         "update_environment_services_by_id",
		Title:        "Update PingOne Environment Services by ID",
		Description:  "Update the services assigned to a PingOne environment (update's the environment's Bill of Materials) by the environment's unique ID.",
		InputSchema:  schema.MustGenerateSchema[UpdateEnvironmentServicesByIdInput](),
		OutputSchema: schema.MustGenerateSchema[UpdateEnvironmentServicesByIdOutput](),
	},
}

type UpdateEnvironmentServicesByIdInput struct {
	EnvironmentId uuid.UUID                                        `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Services      pingone.EnvironmentBillOfMaterialsReplaceRequest `json:"services" jsonschema:"REQUIRED. The bill of materials for the environment, including products and solution type."`
}

type UpdateEnvironmentServicesByIdOutput struct {
	Services pingone.EnvironmentBillOfMaterialsResponse `json:"services" jsonschema:"The updated bill of materials for the environment, including products and solution type"`
}

// UpdateEnvironmentServicesByIdHandler updates PingOne environment services by ID using the provided client
func UpdateEnvironmentServicesByIdHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateEnvironmentServicesByIdInput,
) (
	*mcp.CallToolResult,
	*UpdateEnvironmentServicesByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateEnvironmentServicesByIdInput) (*mcp.CallToolResult, *UpdateEnvironmentServicesByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdateEnvironmentServicesByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentServicesByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Updating environment services",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(input.Services.Products)))

		// Call the API to update the environment services
		services, httpResponse, err := client.UpdateEnvironmentServicesById(ctx, input.EnvironmentId, &input.Services)
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

		logger.FromContext(ctx).Debug("Environment services updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.Int("productCount", len(services.Products)))

		// Filter out _links field from response
		services.Links = nil

		result := &UpdateEnvironmentServicesByIdOutput{
			Services: *services,
		}

		return nil, result, nil
	}
}
