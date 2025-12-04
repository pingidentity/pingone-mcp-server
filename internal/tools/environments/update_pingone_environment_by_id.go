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

var UpdateEnvironmentByIdDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:  "update_environment_by_id",
		Title: "Update PingOne Environment by ID",
		Description: `Update environment configuration using full replacement (HTTP PUT).

WORKFLOW - Required to avoid data loss:
1. Call 'get_environment_by_id' to fetch current configuration
2. Modify only the fields you want to change
3. Pass the complete merged object to this tool

Omitted optional fields will be cleared. Common updates: name, description, type (SANDBOX→PRODUCTION is permanent). Cannot change: region, ID.`,
		InputSchema:  schema.MustGenerateSchema[UpdateEnvironmentByIdInput](),
		OutputSchema: schema.MustGenerateSchema[UpdateEnvironmentByIdOutput](),
	},
}

// UpdateEnvironmentByIdInput defines the input parameters for updating an environment
type UpdateEnvironmentByIdInput struct {
	BillOfMaterials *pingone.EnvironmentBillOfMaterialsReplaceRequest `json:"billOfMaterials,omitempty" jsonschema:"OPTIONAL. The Bill of Materials for the environment. Specifies the PingOne and non-PingOne products and services associated with this environment deployment."`
	Description     *string                                           `json:"description,omitempty" jsonschema:"OPTIONAL. The description of the environment."`
	EnvironmentId   uuid.UUID                                         `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment to update."`
	Icon            *string                                           `json:"icon,omitempty" jsonschema:"OPTIONAL. The URL referencing the image to use for the environment icon. The supported image types are JPEG/JPG, PNG, and GIF."`
	License         *pingone.EnvironmentLicense                       `json:"license,omitempty" jsonschema:"OPTIONAL. The active license associated with this environment. Required only if your organization has more than one active license."`
	Name            string                                            `json:"name" jsonschema:"REQUIRED. Environment name, must be unique within organization."`
	Region          pingone.EnvironmentRegionCode                     `json:"region" jsonschema:"REQUIRED. Region code (NA/CA/EU/AU/SG/AP). Set at creation, immutable."`
	Status          *pingone.EnvironmentStatusValue                   `json:"status,omitempty" jsonschema:"OPTIONAL. ACTIVE or DELETE_PENDING. For PRODUCTION environments, use Update Environment Status endpoint instead."`
	Type            pingone.EnvironmentTypeValue                      `json:"type" jsonschema:"REQUIRED. PRODUCTION or SANDBOX. SANDBOX can be promoted to PRODUCTION (permanent, cannot revert)."`
}

// UpdateEnvironmentByIdOutput represents the result of updating an environment
type UpdateEnvironmentByIdOutput struct {
	Environment pingone.EnvironmentResponse `json:"environment" jsonschema:"The updated environment details including ID, name, type, region, and metadata"`
}

// UpdateEnvironmentByIdHandler updates a PingOne environment by ID using the provided client
func UpdateEnvironmentByIdHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateEnvironmentByIdInput,
) (
	*mcp.CallToolResult,
	*UpdateEnvironmentByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateEnvironmentByIdInput) (*mcp.CallToolResult, *UpdateEnvironmentByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdateEnvironmentByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdateEnvironmentByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Updating environment",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("name", input.Name),
			slog.String("region", string(input.Region)),
			slog.String("type", string(input.Type)))

		// Build the environment replace request
		replaceRequest := pingone.NewEnvironmentReplaceRequest(
			input.Name,
			input.Region,
			input.Type,
		)

		// Set optional fields if provided
		if input.Description != nil {
			replaceRequest.SetDescription(*input.Description)
		}
		if input.Icon != nil {
			replaceRequest.SetIcon(*input.Icon)
		}
		if input.BillOfMaterials != nil {
			replaceRequest.SetBillOfMaterials(*input.BillOfMaterials)
		}
		if input.License != nil {
			replaceRequest.SetLicense(*input.License)
		}
		if input.Status != nil {
			replaceRequest.SetStatus(*input.Status)
		}

		// Call the API to update the environment
		environment, httpResponse, err := client.UpdateEnvironmentById(ctx, input.EnvironmentId, replaceRequest)
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

		logger.FromContext(ctx).Debug("Environment updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("name", input.Name),
			slog.String("region", string(input.Region)),
			slog.String("type", string(input.Type)),
		)

		// Filter out _links field from response
		environment.Links = nil

		result := &UpdateEnvironmentByIdOutput{
			Environment: *environment,
		}

		return nil, result, nil
	}
}
