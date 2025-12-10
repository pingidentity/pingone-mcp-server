// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var UpdateApplicationByIdDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:  "update_oidc_application",
		Title: "Update PingOne OIDC Application by ID",
		Description: `Update OIDC application configuration using full replacement (HTTP PUT).

WORKFLOW - Required to avoid data loss:
1. Call 'get_application' to fetch current configuration
2. Modify only the fields you want to change
3. Pass the complete merged object to this tool

Omitted optional fields will be cleared.`,
		InputSchema:  schema.MustGenerateSchema[UpdateApplicationByIdInput](),
		OutputSchema: schema.MustGenerateSchema[UpdateApplicationByIdOutput](),
	},
}

type UpdateApplicationByIdInput struct {
	EnvironmentId uuid.UUID                  `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	ApplicationId uuid.UUID                  `json:"applicationId" jsonschema:"REQUIRED. Application UUID."`
	Application   management.ApplicationOIDC `json:"application" jsonschema:"REQUIRED. The complete OIDC application config with modifications."`
}

type UpdateApplicationByIdOutput struct {
	Application management.ApplicationOIDC `json:"application" jsonschema:"The updated application configuration details"`
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

		updateRequest := management.UpdateApplicationRequest{
			ApplicationOIDC: &input.Application,
		}

		// Call the API to update the application
		applicationResponse, httpResponse, err := client.UpdateApplicationById(ctx, input.EnvironmentId, input.ApplicationId, updateRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if applicationResponse == nil || applicationResponse.ApplicationOIDC == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Application updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()),
		)

		// Filter out the _links field
		applicationResponse.ApplicationOIDC.Links = nil
		result := &UpdateApplicationByIdOutput{
			Application: *applicationResponse.ApplicationOIDC,
		}

		return nil, result, nil
	}
}
