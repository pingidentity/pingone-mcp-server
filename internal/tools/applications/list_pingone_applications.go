// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var ListApplicationsDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "list_applications",
		Title:        "List PingOne Applications",
		Description:  "Lists all applications in an environment. Use to discover application IDs or review configurations before updates. Returns OIDC, SAML, External Link, and PingOne system applications.",
		InputSchema:  schema.MustGenerateSchema[ListApplicationsInput](),
		OutputSchema: schema.MustGenerateSchema[ListApplicationsOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type ListApplicationsInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
}

type ApplicationSummary struct {
	Id        *string                             `json:"id,omitempty" jsonschema:"The UUID of the application"`
	Name      string                              `json:"name" jsonschema:"The name of the application"`
	Protocol  *management.EnumApplicationProtocol `json:"protocol,omitempty" jsonschema:"The protocol type of the application"`
	Type      *management.EnumApplicationType     `json:"type,omitempty" jsonschema:"The type of the application"`
	CreatedAt *time.Time                          `json:"createdAt,omitempty" jsonschema:"The creation timestamp of the application"`
}

type ListApplicationsOutput struct {
	Applications []ApplicationSummary `json:"applications" jsonschema:"List of applications with their configuration details"`
}

// ListApplicationsHandler lists all PingOne applications using the provided client
func ListApplicationsHandler(clientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListApplicationsInput,
) (
	*mcp.CallToolResult,
	*ListApplicationsOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListApplicationsInput) (*mcp.CallToolResult, *ListApplicationsOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, ListApplicationsDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListApplicationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := clientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListApplicationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Listing applications", slog.String("environmentId", input.EnvironmentId.String()))

		// Call the API to list applications
		pagedIterator, err := client.GetApplications(ctx, input.EnvironmentId)
		if err != nil {
			toolErr := errs.NewToolError(ListApplicationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}
		// Aggregate all pages into one response
		result := ListApplicationsOutput{
			Applications: []ApplicationSummary{},
		}
		for next, err := range pagedIterator {
			logger.LogHttpResponse(ctx, next.HTTPResponse)
			if err != nil {
				apiErr := errs.NewApiError(next.HTTPResponse, err)
				errs.Log(ctx, apiErr)
				return nil, nil, apiErr
			}
			if next.EntityArray == nil || next.EntityArray.Embedded == nil {
				// This should never happen, err should be set if no data
				apiErr := errs.NewApiError(next.HTTPResponse, errors.New("no data in response"))
				errs.Log(ctx, apiErr)
				return nil, nil, apiErr
			}
			logger.FromContext(ctx).Debug("Retrieved applications page", slog.Int("count", len(next.EntityArray.Embedded.Applications)))
			for _, sdkApp := range next.EntityArray.Embedded.Applications {
				applicationSummary, err := getApplicationSummary(&sdkApp)
				if err != nil {
					toolErr := errs.NewToolError(GetApplicationByIdDef.McpTool.Name, err)
					errs.Log(ctx, toolErr)
					return nil, nil, toolErr
				}
				result.Applications = append(result.Applications, *applicationSummary)
			}
		}

		return nil, &result, nil
	}
}

const adminConsoleAppName = "PingOne Admin Console"

func getApplicationSummary(application *management.ReadOneApplication200Response) (*ApplicationSummary, error) {
	result := ApplicationSummary{}
	switch {
	case application.ApplicationExternalLink != nil:
		result.Id = application.ApplicationExternalLink.Id
		result.Name = application.ApplicationExternalLink.Name
		result.Protocol = &application.ApplicationExternalLink.Protocol
		result.Type = &application.ApplicationExternalLink.Type
		result.CreatedAt = application.ApplicationExternalLink.CreatedAt
	case application.ApplicationOIDC != nil:
		result.Id = application.ApplicationOIDC.Id
		result.Name = application.ApplicationOIDC.Name
		result.Protocol = &application.ApplicationOIDC.Protocol
		result.Type = &application.ApplicationOIDC.Type
		result.CreatedAt = application.ApplicationOIDC.CreatedAt
	case application.ApplicationPingOneAdminConsole != nil:
		result.Name = adminConsoleAppName
		adminConsoleType := management.ENUMAPPLICATIONTYPE_PING_ONE_ADMIN_CONSOLE
		result.Type = &adminConsoleType
	case application.ApplicationPingOnePortal != nil:
		result.Id = application.ApplicationPingOnePortal.Id
		result.Name = application.ApplicationPingOnePortal.Name
		result.Protocol = &application.ApplicationPingOnePortal.Protocol
		result.Type = &application.ApplicationPingOnePortal.Type
		result.CreatedAt = application.ApplicationPingOnePortal.CreatedAt
	case application.ApplicationPingOneSelfService != nil:
		result.Id = application.ApplicationPingOneSelfService.Id
		result.Name = application.ApplicationPingOneSelfService.Name
		result.Protocol = &application.ApplicationPingOneSelfService.Protocol
		result.Type = &application.ApplicationPingOneSelfService.Type
		result.CreatedAt = application.ApplicationPingOneSelfService.CreatedAt
	case application.ApplicationSAML != nil:
		result.Id = application.ApplicationSAML.Id
		result.Name = application.ApplicationSAML.Name
		result.Protocol = &application.ApplicationSAML.Protocol
		result.Type = &application.ApplicationSAML.Type
		result.CreatedAt = application.ApplicationSAML.CreatedAt
	case application.ApplicationWSFED != nil:
		result.Id = application.ApplicationWSFED.Id
		result.Name = application.ApplicationWSFED.Name
		result.Protocol = &application.ApplicationWSFED.Protocol
		result.Type = &application.ApplicationWSFED.Type
		result.CreatedAt = application.ApplicationWSFED.CreatedAt
	default:
		return nil, fmt.Errorf("unknown application type in response")
	}
	return &result, nil
}
