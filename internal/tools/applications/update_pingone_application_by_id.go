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
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:  "update_application_by_id",
		Title: "Update PingOne Application by ID",
		Description: `Update an existing application within a specified PingOne environment.

VERY IMPORTANT: Before updating, first get the latest application configuration using the 'get_application_by_id' tool to avoid overwriting pre-existing optional configuration values.`,
		InputSchema:  mustGenerateUpdateApplicationByIdSchema[UpdateApplicationByIdInput](),
		OutputSchema: mustGenerateUpdateApplicationByIdSchema[UpdateApplicationByIdOutput](),
	},
}

type UpdateApplicationByIdInput struct {
	EnvironmentId uuid.UUID              `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId uuid.UUID              `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
	Application   UpdateApplicationModel `json:"application" jsonschema:"REQUIRED. The updated application configuration details"`
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
