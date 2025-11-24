// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"fmt"
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

var GetApplicationByIdDef = types.ToolDefinition{
	IsReadOnly: true,
	McpTool: &mcp.Tool{
		Name:         "get_application_by_id",
		Title:        "Get PingOne Application by ID",
		Description:  "Retrieve an application's configuration by its unique ID within a specified PingOne environment.",
		InputSchema:  schema.MustGenerateSchema[GetApplicationByIdInput](),
		OutputSchema: MustGenerateGetApplicationByIdOutputSchema(),
	},
}

type GetApplicationByIdInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	ApplicationId uuid.UUID `json:"applicationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne application"`
}

type GetApplicationByIdOutput struct {
	Application ReadApplicationModel `json:"application" jsonschema:"The configuration details of the retrieved PingOne application"`
}

func MustGenerateGetApplicationByIdOutputSchema() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[GetApplicationByIdOutput]()
	// Modify the Application property to use the ApplicationModel schema with oneOf constraint
	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating GetApplicationByIdOutput schema")
	}
	baseSchema.Properties["application"] = MustGenerateReadApplicationModelSchema()
	return baseSchema
}

// GetApplicationByIdHandler retrieves a PingOne application by ID using the provided client
func GetApplicationByIdHandler(applicationsClientFactory ApplicationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetApplicationByIdInput,
) (
	*mcp.CallToolResult,
	*GetApplicationByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetApplicationByIdInput) (*mcp.CallToolResult, *GetApplicationByIdOutput, error) {
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

		result := &GetApplicationByIdOutput{
			Application: ReadApplicationModelFromSDKReadResponse(*application),
		}

		return nil, result, nil
	}
}
