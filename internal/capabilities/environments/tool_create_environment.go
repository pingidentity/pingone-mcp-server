// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

var CreateEnvironmentDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		ProductionEnvironmentNotApplicable: true, // Tool does not act on an existing environment
	},
	McpTool: &mcp.Tool{
		Name:         "create_environment",
		Title:        "Create PingOne Environment",
		Description:  "Create a new sandbox PingOne environment. Only SANDBOX type supported via API (PRODUCTION must be created via admin console). Requires license quota. Environment becomes available immediately but services may take 10-30 seconds to initialize.",
		InputSchema:  schema.MustGenerateSchema[CreateEnvironmentInput](),
		OutputSchema: schema.MustGenerateSchema[CreateEnvironmentOutput](),
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: func() *bool { b := false; return &b }(),
		},
	},
}

// CreateEnvironmentInput defines the input parameters for creating an environment
type CreateEnvironmentInput struct {
	BillOfMaterials *pingone.EnvironmentBillOfMaterials `json:"billOfMaterials,omitempty" jsonschema:"OPTIONAL. The Bill of Materials for the environment. Create requests that do not specify this property receive a default PingOne Bill of Materials on creation. Specifies the PingOne and non-PingOne products and services associated with this environment deployment."`
	Description     *string                             `json:"description,omitempty" jsonschema:"OPTIONAL. The description of the environment."`
	Icon            *string                             `json:"icon,omitempty" jsonschema:"OPTIONAL. The URL referencing the image to use for the environment icon. The supported image types are JPEG/JPG, PNG, and GIF."`
	License         pingone.EnvironmentLicense          `json:"license" jsonschema:"REQUIRED. The active license associated with this environment. Required only if your organization has more than one active license."`
	Name            string                              `json:"name" jsonschema:"REQUIRED. Environment name, must be unique within organization."`
	Region          pingone.EnvironmentRegionCode       `json:"region" jsonschema:"REQUIRED. Region code: NA, CA, EU, AU, SG, or AP. Cannot be changed after creation."`
}

// CreateEnvironmentOutput represents the result of creating an environment
type CreateEnvironmentOutput struct {
	Environment pingone.EnvironmentResponse `json:"environment" jsonschema:"The created environment details including ID, name, type, region, and metadata"`
}

// CreateEnvironmentHandler creates a new PingOne environment using the provided client
func CreateEnvironmentHandler(environmentsClientFactory EnvironmentsClientFactory) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreateEnvironmentInput,
) (
	*mcp.CallToolResult,
	*CreateEnvironmentOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateEnvironmentInput) (*mcp.CallToolResult, *CreateEnvironmentOutput, error) {
		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(CreateEnvironmentDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Creating environment",
			slog.String("name", input.Name),
			slog.String("region", string(input.Region)),
			slog.String("type", "SANDBOX"),
		)

		// SANDBOX is hardcoded as PRODUCTION environments are not supported via this MCP tool
		createRequest := pingone.NewEnvironmentCreateRequest(
			input.Name,
			input.Region,
			pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
			input.License,
		)

		// Set optional fields if provided
		if input.Description != nil {
			createRequest.SetDescription(*input.Description)
		}
		if input.Icon != nil {
			createRequest.SetIcon(*input.Icon)
		}
		if input.BillOfMaterials != nil {
			createRequest.SetBillOfMaterials(*input.BillOfMaterials)
		}

		// Call the API to create the environment
		envResponse, httpResponse, err := client.CreateEnvironment(ctx, createRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if envResponse == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no environment data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Environment created successfully",
			slog.String("environmentId", envResponse.Id.String()),
			slog.String("name", envResponse.Name))

		// Filter out _links field from response
		envResponse.Links = nil

		result := &CreateEnvironmentOutput{
			Environment: *envResponse,
		}

		return nil, result, nil
	}
}
