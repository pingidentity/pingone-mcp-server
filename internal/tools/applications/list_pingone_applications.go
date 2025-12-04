// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"errors"
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

var ListApplicationsDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "list_applications",
		Title:        "List PingOne Applications",
		Description:  "Lists all applications in an environment. Use to discover application IDs or review configurations before updates. Returns OIDC, SAML, External Link, and PingOne system applications.",
		InputSchema:  schema.MustGenerateSchema[ListApplicationsInput](),
		OutputSchema: MustGenerateListApplicationsOutputSchema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type ListApplicationsInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
}

type ListApplicationsOutput struct {
	Applications []ReadApplicationModel `json:"applications" jsonschema:"List of applications with their configuration details"`
}

func MustGenerateListApplicationsOutputSchema() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[ListApplicationsOutput]()
	// Modify the Applications property to use the ApplicationModel schema with oneOf constraint
	applicationModelSchema := MustGenerateReadApplicationModelSchema()
	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating ListApplicationsOutput schema")
	}
	appsProp, ok := baseSchema.Properties["applications"]
	if !ok {
		panic("applications property not found in ListApplicationsOutput schema")
	}
	appsProp.Items = applicationModelSchema
	return baseSchema
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
			Applications: []ReadApplicationModel{},
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
				result.Applications = append(result.Applications, ReadApplicationModelFromSDKReadResponse(sdkApp))
			}
		}

		return nil, &result, nil
	}
}
