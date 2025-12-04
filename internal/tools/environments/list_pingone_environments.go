// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"errors"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var ListEnvironmentsDef = types.ToolDefinition{
	IsReadOnly: true,
	McpTool: &mcp.Tool{
		Name:  "list_environments",
		Title: "List PingOne Environments",
		Description: `Lists all PingOne environments accessible to the authenticated user.

Use to discover environment IDs needed for other operations or to find environments by name/status.

Only filters that 'name' with 'sw' (starts with); 'id', 'organization.id', 'license.id', 'status' with 'eq' (equals); 'and' to combine are valid.

Filter examples:
- name sw "Prod"
- status eq "ACTIVE"
- name sw "Dev" and status eq "ACTIVE"

Returns: Array of environments with ID, name, type, region, license, and metadata.`,
		InputSchema:  schema.MustGenerateSchema[ListEnvironmentsInput](),
		OutputSchema: schema.MustGenerateSchema[ListEnvironmentsOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

// ListEnvironmentsInput defines the input parameters for listing environments
type ListEnvironmentsInput struct {
	Filter *string `json:"filter,omitempty" jsonschema:"OPTIONAL. SCIM filter. Only filters that 'name' with 'sw' (starts with); 'id', 'organization.id', 'license.id', 'status' with 'eq' (equals); 'and' to combine are valid."`
}

// ListEnvironmentsOutput represents the result of listing environments
type ListEnvironmentsOutput struct {
	Environments []pingone.EnvironmentResponse `json:"environments" jsonschema:"List of environments with their details including ID, name, type, region, and metadata"`
}

// ListEnvironmentsHandler lists all PingOne environments using the provided client
func ListEnvironmentsHandler(environmentsClientFactory EnvironmentsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListEnvironmentsInput,
) (
	*mcp.CallToolResult,
	*ListEnvironmentsOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListEnvironmentsInput) (*mcp.CallToolResult, *ListEnvironmentsOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, ListEnvironmentsDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListEnvironmentsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := environmentsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListEnvironmentsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		if input.Filter != nil {
			logger.FromContext(ctx).Debug("Using filter",
				slog.String("filter", *input.Filter))
		}

		// Call the API to list environments
		pagedIterator, err := client.GetEnvironments(ctx, input.Filter)
		if err != nil {
			toolErr := errs.NewToolError(ListEnvironmentsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		// Aggregate all pages into one response
		result := ListEnvironmentsOutput{
			Environments: []pingone.EnvironmentResponse{},
		}
		for next, err := range pagedIterator {
			logger.LogHttpResponse(ctx, next.HTTPResponse)

			if err != nil {
				apiErr := errs.NewApiError(next.HTTPResponse, err)
				errs.Log(ctx, apiErr)
				return nil, nil, apiErr
			}

			if next.Data == nil || next.Data.Embedded == nil {
				// This should never happen, err should be set if no data
				apiErr := errs.NewApiError(next.HTTPResponse, errors.New("no data in response"))
				errs.Log(ctx, apiErr)
				return nil, nil, apiErr
			}

			logger.FromContext(ctx).Debug("Retrieved environments page",
				slog.Int("count", len(next.Data.Embedded.Environments)))

			result.Environments = append(result.Environments, next.Data.Embedded.Environments...)
		}

		// Filter out _links field from all environment responses
		for i := range result.Environments {
			result.Environments[i].Links = nil
		}

		return nil, &result, nil
	}
}
