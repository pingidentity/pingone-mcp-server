// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

var ListPopulationsDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "list_populations",
		Title:        "List PingOne Populations",
		Description:  "Lists populations in an environment. Use to find population IDs or review configurations. Filter examples: name sw \"External\", id eq \"pop-uuid\". Supported: 'id' with 'eq', 'name' with 'sw'.",
		InputSchema:  schema.MustGenerateSchema[ListPopulationsInput](),
		OutputSchema: schema.MustGenerateSchema[ListPopulationsOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type ListPopulationsInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	Filter        *string   `json:"filter,omitempty" jsonschema:"OPTIONAL. SCIM filter. Supported: 'id' with 'eq', 'name' with 'sw'."`
}

type PopulationSummary struct {
	Id        *string `json:"id" jsonschema:"The unique identifier of the population"`
	Name      string  `json:"name" jsonschema:"The name of the population"`
	Default   *bool   `json:"default,omitempty" jsonschema:"Indicates if this is the default population"`
	CreatedAt *string `json:"createdAt,omitempty" jsonschema:"The creation timestamp of the population"`
}

type ListPopulationsOutput struct {
	Populations []PopulationSummary `json:"populations" jsonschema:"List of populations with their id and name"`
}

// ListPopulationsHandler lists all PingOne populations using the provided client
func ListPopulationsHandler(populationsClientFactory PopulationsClientFactory) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListPopulationsInput,
) (
	*mcp.CallToolResult,
	*ListPopulationsOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListPopulationsInput) (*mcp.CallToolResult, *ListPopulationsOutput, error) {
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
			Populations: []PopulationSummary{},
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

			// Convert each population to PopulationSummary
			for _, pop := range next.EntityArray.Embedded.Populations {
				result.Populations = append(result.Populations, PopulationSummary{
					Id:        pop.Id,
					Name:      pop.Name,
					Default:   pop.Default,
					CreatedAt: pop.CreatedAt,
				})
			}
		}

		return nil, &result, nil
	}
}
