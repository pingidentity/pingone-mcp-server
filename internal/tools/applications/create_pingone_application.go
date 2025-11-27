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

var CreateApplicationDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:         "create_application",
		Title:        "Create PingOne Application",
		Description:  "Create a new OAuth 2.0, SAML, or external link application within a specified PingOne environment.",
		InputSchema:  mustGenerateCreateApplicationSchema[CreateApplicationInput](),
		OutputSchema: mustGenerateCreateApplicationSchema[CreateApplicationOutput](),
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: func() *bool { b := false; return &b }(),
		},
	},
}

type CreateApplicationInput struct {
	EnvironmentId uuid.UUID              `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Application   CreateApplicationModel `json:"application" jsonschema:"REQUIRED. The application configuration details"`
}

type CreateApplicationOutput struct {
	Application CreateApplicationModel `json:"application" jsonschema:"The created application details"`
}

func mustGenerateCreateApplicationSchema[T any]() *jsonschema.Schema {
	baseSchema := schema.MustGenerateSchema[T]()
	// Modify the Application property to use the CreateApplicationModel schema with oneOf constraint
	applicationModelSchema := MustGenerateCreateApplicationModelSchema()
	if baseSchema.Properties == nil {
		panic("baseSchema.Properties is nil when generating CreateApplicationInput schema")
	}
	baseSchema.Properties["application"] = applicationModelSchema
	return baseSchema
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

		createRequest := CreateApplicationModelToSDKCreateRequest(input.Application)

		// Call the API to create the application
		applicationResponse, httpResponse, err := client.CreateApplication(ctx, input.EnvironmentId, createRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if applicationResponse == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no application data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Application created successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
		)

		result := &CreateApplicationOutput{
			Application: CreateApplicationModelFromSDKCreateResponse(*applicationResponse),
		}

		return nil, result, nil
	}
}
