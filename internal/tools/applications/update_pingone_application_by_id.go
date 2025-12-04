// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var UpdateApplicationByIdDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:  "update_application_by_id",
		Title: "Update PingOne Application by ID",
		Description: `Update application configuration using full replacement (HTTP PUT).

WORKFLOW - Required to avoid data loss:
1. Call 'get_application_by_id' to fetch current configuration
2. Modify only the fields you want to change
3. Pass the complete merged object to this tool

Omitted optional fields will be cleared.`,
		InputSchema:  mustGenerateUpdateApplicationByIdSchema[UpdateApplicationByIdInput](),
		OutputSchema: mustGenerateUpdateApplicationByIdSchema[UpdateApplicationByIdOutput](),
	},
}

type UpdateApplicationByIdInput struct {
	EnvironmentId uuid.UUID              `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	ApplicationId uuid.UUID              `json:"applicationId" jsonschema:"REQUIRED. Application UUID."`
	Application   UpdateApplicationModel `json:"application" jsonschema:"REQUIRED. Complete application config with modifications."`
}

type UpdateApplicationByIdOutput struct {
	Application UpdateApplicationModel `json:"application" jsonschema:"The updated application configuration details"`
}

func mustGenerateUpdateApplicationByIdSchema[T any]() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[T]()
	// Modify the Application property to use the UpdateApplicationModel schema with oneOf constraint
	applicationModelSchema := MustGenerateUpdateApplicationModelSchema()
	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating UpdateApplicationByIdInput schema")
	}
	baseSchema.Properties["application"] = applicationModelSchema
	return baseSchema
}

func UpdateApplicationByIdHandler(applicationsClientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateApplicationByIdInput,
) (
	*mcp.CallToolResult,
	*UpdateApplicationByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateApplicationByIdInput) (*mcp.CallToolResult, *UpdateApplicationByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdateApplicationByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := applicationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Updating application",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()),
		)

		updateRequest := UpdateApplicationModelToSDKUpdateRequest(input.Application)

		// Call the API to update the application
		applicationResponse, httpResponse, err := client.UpdateApplicationById(ctx, input.EnvironmentId, input.ApplicationId, updateRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if applicationResponse == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Application updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()),
		)

		result := &UpdateApplicationByIdOutput{
			Application: UpdateApplicationModelFromSDKReadResponse(*applicationResponse),
		}

		return nil, result, nil
	}
}
