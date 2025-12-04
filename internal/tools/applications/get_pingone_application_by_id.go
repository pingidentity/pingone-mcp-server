// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"encoding/json"
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

var GetApplicationByIdDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:        "get_application_by_id",
		Title:       "Get PingOne Application by ID",
		Description: "Retrieve application configuration by ID. Use 'list_applications' first if you need to find the application ID. Call before 'update_application_by_id' to get current settings.",
		InputSchema: schema.MustGenerateSchema[GetApplicationByIdInput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type GetApplicationByIdInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId uuid.UUID `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
}

// GetApplicationByIdHandler retrieves a PingOne application by ID using the provided client
func GetApplicationByIdHandler(applicationsClientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetApplicationByIdInput,
) (
	*mcp.CallToolResult,
	any,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetApplicationByIdInput) (*mcp.CallToolResult, any, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetApplicationByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := applicationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving application",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()))

		// Call the API to retrieve the application
		application, httpResponse, err := client.GetApplication(ctx, input.EnvironmentId, input.ApplicationId)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if application == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Application retrieved successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("applicationId", input.ApplicationId.String()))

		// Serialize the application based on its type, filtering out the _links field
		formattedApplication, err := GetOutputFormattedApplication(application)
		if err != nil {
			toolErr := errs.NewToolError(GetApplicationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}
		applicationJsonBytes, err := json.Marshal(formattedApplication)
		if err != nil {
			toolErr := errs.NewToolError(GetApplicationByIdDef.McpTool.Name, fmt.Errorf("failed to marshal formatted application response: %w", err))
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(applicationJsonBytes),
				},
			},
		}, nil, nil
	}
}

func GetOutputFormattedApplication(application *management.ReadOneApplication200Response) (any, error) {
	// Return the configured application type, and filter out the _links field
	switch {
	case application.ApplicationExternalLink != nil:
		application.ApplicationExternalLink.Links = nil
		return application.ApplicationExternalLink, nil
	case application.ApplicationOIDC != nil:
		application.ApplicationOIDC.Links = nil
		return application.ApplicationOIDC, nil
	case application.ApplicationPingOneAdminConsole != nil:
		// No links field to remove
		return application.ApplicationPingOneAdminConsole, nil
	case application.ApplicationPingOnePortal != nil:
		application.ApplicationPingOnePortal.Links = nil
		return application.ApplicationPingOnePortal, nil
	case application.ApplicationPingOneSelfService != nil:
		application.ApplicationPingOneSelfService.Links = nil
		return application.ApplicationPingOneSelfService, nil
	case application.ApplicationSAML != nil:
		application.ApplicationSAML.Links = nil
		return application.ApplicationSAML, nil
	case application.ApplicationWSFED != nil:
		application.ApplicationWSFED.Links = nil
		return application.ApplicationWSFED, nil
	default:
		return nil, fmt.Errorf("unknown application type in response")
	}
}
