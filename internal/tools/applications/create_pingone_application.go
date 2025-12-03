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

var CreateApplicationDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:         "create_application",
		Title:        "Create PingOne OIDC Application",
		Description:  "Create a new OIDC application within a specified PingOne environment.",
		InputSchema:  schema.MustGenerateSchema[CreateApplicationInput](),
		OutputSchema: schema.MustGenerateSchema[CreateApplicationOutput](),
	},
}

type CreateApplicationInput struct {
	EnvironmentId uuid.UUID                  `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
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

		result := &CreateApplicationOutput{
			Application: *applicationResponse.ApplicationOIDC,
		}

		return nil, result, nil
	}
}
