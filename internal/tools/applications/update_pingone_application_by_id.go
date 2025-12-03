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
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:         "update_application_by_id",
		Title:        "Update PingOne OIDC Application by ID",
		Description:  "Update an existing OIDC application within a specified PingOne environment.",
		InputSchema:  schema.MustGenerateSchema[UpdateApplicationByIdInput](),
		OutputSchema: schema.MustGenerateSchema[UpdateApplicationByIdOutput](),
	},
}

type UpdateApplicationByIdInput struct {
	EnvironmentId     uuid.UUID            `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId     uuid.UUID            `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
	ApplicationParent OIDCApplicationModel `json:"applicationOIDC" jsonschema:"REQUIRED. The OIDC application configuration details"`
}

type UpdateApplicationByIdOutput struct {
	ApplicationParent OIDCApplicationModel `json:"applicationOIDC" jsonschema:"The updated application configuration details"`
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
			ApplicationOIDC: input.ApplicationParent.ApplicationOIDC,
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

		result := &UpdateApplicationByIdOutput{
			ApplicationParent: OIDCApplicationModel{
				ApplicationOIDC: applicationResponse.ApplicationOIDC,
			},
		}

		return nil, result, nil
	}
}
