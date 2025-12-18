// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

var CreateApplicationDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:         "create_oidc_application",
		Title:        "Create PingOne OIDC Application",
		Description:  "Create a new OIDC application within a specified PingOne environment.",
		InputSchema:  schema.MustGenerateSchema[CreateApplicationInput](),
		OutputSchema: schema.MustGenerateSchema[CreateApplicationOutput](),
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: func() *bool { b := false; return &b }(),
		},
	},
}

type CreateApplicationInput struct {
	EnvironmentId uuid.UUID                  `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	Application   management.ApplicationOIDC `json:"application" jsonschema:"REQUIRED. The OIDC application configuration details"`
}

type CreateApplicationOutput struct {
	Application management.ApplicationOIDC `json:"application" jsonschema:"The created application details"`
}

// CreateApplicationHandler creates a new PingOne application using the provided client
func CreateApplicationHandler(applicationsClientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreateApplicationInput,
) (
	*mcp.CallToolResult,
	*CreateApplicationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateApplicationInput) (*mcp.CallToolResult, *CreateApplicationOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, CreateApplicationDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(CreateApplicationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := applicationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(CreateApplicationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Creating application",
			slog.String("environmentId", input.EnvironmentId.String()),
		)

		createRequest := management.CreateApplicationRequest{
			ApplicationOIDC: &input.Application,
		}

		// Call the API to create the application
		applicationResponse, httpResponse, err := client.CreateApplication(ctx, input.EnvironmentId, createRequest)
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

		logger.FromContext(ctx).Debug("Application created successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
		)

		// Filter out the _links field
		applicationResponse.ApplicationOIDC.Links = nil
		result := &CreateApplicationOutput{
			Application: *applicationResponse.ApplicationOIDC,
		}

		return nil, result, nil
	}
}
