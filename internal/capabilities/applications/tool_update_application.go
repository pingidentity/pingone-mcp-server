// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

var UpdateApplicationDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:  "update_oidc_application",
		Title: "Update PingOne OIDC Application by ID",
		Description: `Update OIDC application configuration using full replacement (HTTP PUT).

WORKFLOW - Required to avoid data loss:
1. Call 'get_application' to fetch current configuration
2. Modify only the fields you want to change
3. Pass the complete merged object to this tool

Omitted optional fields will be cleared.`,
		InputSchema:  schema.MustGenerateSchema[UpdateApplicationInput](),
		OutputSchema: schema.MustGenerateSchema[UpdateApplicationOutput](),
	},
}

type UpdateApplicationInput struct {
	EnvironmentId uuid.UUID                  `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	ApplicationId uuid.UUID                  `json:"applicationId" jsonschema:"REQUIRED. Application UUID."`
	Application   management.ApplicationOIDC `json:"application" jsonschema:"REQUIRED. The complete OIDC application config with modifications."`
}

type UpdateApplicationOutput struct {
	Application management.ApplicationOIDC `json:"application" jsonschema:"The updated application configuration details"`
}

func UpdateApplicationHandler(applicationsClientFactory ApplicationsClientFactory) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateApplicationInput,
) (
	*mcp.CallToolResult,
	*UpdateApplicationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateApplicationInput) (*mcp.CallToolResult, *UpdateApplicationOutput, error) {
		client, err := applicationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateApplicationDef.McpTool.Name, err)
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
		applicationResponse, httpResponse, err := client.UpdateApplication(ctx, input.EnvironmentId, input.ApplicationId, updateRequest)
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
		result := &UpdateApplicationOutput{
			Application: *applicationResponse.ApplicationOIDC,
		}

		return nil, result, nil
	}
}
