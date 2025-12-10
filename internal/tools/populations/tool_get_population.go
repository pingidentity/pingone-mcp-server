// Copyright © 2025 Ping Identity Corporation

package populations

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

var GetPopulationDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true, // this is true while the tool does not return any actual user data
	},
	McpTool: &mcp.Tool{
		Name:         "get_population",
		Title:        "Get PingOne Population by ID",
		Description:  "Retrieve population configuration by ID. Use 'list_populations' first if you need to find the population ID. Call before 'update_population' to get current settings.",
		InputSchema:  schema.MustGenerateSchema[GetPopulationInput](),
		OutputSchema: schema.MustGenerateSchema[GetPopulationOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

type GetPopulationInput struct {
	EnvironmentId uuid.UUID `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	PopulationId  uuid.UUID `json:"populationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne population"`
}

type GetPopulationOutput struct {
	Population management.Population `json:"population" jsonschema:"The population details retrieved by ID"`
}

// GetPopulationHandler retrieves a PingOne population by ID using the provided client
func GetPopulationHandler(populationsClientFactory PopulationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetPopulationInput,
) (
	*mcp.CallToolResult,
	*GetPopulationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPopulationInput) (*mcp.CallToolResult, *GetPopulationOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetPopulationDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetPopulationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := populationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetPopulationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving population",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("populationId", input.PopulationId.String()))

		// Call the API to retrieve the population
		population, httpResponse, err := client.GetPopulation(ctx, input.EnvironmentId, input.PopulationId)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if population == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no population data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Population retrieved successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("populationId", input.PopulationId.String()),
			slog.String("populationName", population.Name),
		)

		// Filter out _links field from response
		population.Links = nil

		result := &GetPopulationOutput{
			Population: *population,
		}

		return nil, result, nil
	}
}
