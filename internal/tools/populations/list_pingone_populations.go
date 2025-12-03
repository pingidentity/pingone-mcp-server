// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"errors"
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

var ListPopulationsDef = types.ToolDefinition{
	IsReadOnly: true,
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "list_populations",
		Title:        "List PingOne Populations",
		Description:  "Lists PingOne populations in a specified PingOne environment. Supports optional SCIM filtering to narrow results on the `id` and `name` attributes.  The `id` attribute supports the `eq` (equals) operator, and the `name` attribute supports the `sw` (starts with) operator.",
		InputSchema:  schema.MustGenerateSchema[ListPopulationsInput](),
		OutputSchema: schema.MustGenerateSchema[ListPopulationsOutput](),
	},
}

type ListPopulationsInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Filter        *string   `json:"filter,omitempty" jsonschema:"OPTIONAL. A SCIM filter string to filter populations based on attributes. Supports optional SCIM filtering to narrow results on the 'id' and 'name' attributes. The 'id' attribute supports the 'eq' (equals) operator, and the 'name' attribute supports the 'sw' (starts with) operator."`
}

type ListPopulationsOutput struct {
	Populations []management.Population `json:"populations" jsonschema:"List of populations with their configuration details"`
}

// ListPopulationsHandler lists all PingOne populations using the provided client
func ListPopulationsHandler(populationsClientFactory PopulationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListPopulationsInput,
) (
	*mcp.CallToolResult,
	*ListPopulationsOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListPopulationsInput) (*mcp.CallToolResult, *ListPopulationsOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, ListPopulationsDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListPopulationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := populationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(ListPopulationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		if input.Filter != nil {
			logger.FromContext(ctx).Debug("Using filter",
				slog.String("filter", *input.Filter))
		}

		logger.FromContext(ctx).Debug("Listing populations", slog.String("environmentId", input.EnvironmentId.String()))

		// Call the API to list populations
		pagedIterator, err := client.GetPopulations(ctx, input.EnvironmentId, input.Filter)
		if err != nil {
			toolErr := errs.NewToolError(ListPopulationsDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}
		// Aggregate all pages into one response
		result := ListPopulationsOutput{
			Populations: []management.Population{},
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
			logger.FromContext(ctx).Debug("Retrieved populations page", slog.Int("count", len(next.EntityArray.Embedded.Populations)))
			result.Populations = append(result.Populations, next.EntityArray.Embedded.Populations...)
		}

		return nil, &result, nil
	}
}
