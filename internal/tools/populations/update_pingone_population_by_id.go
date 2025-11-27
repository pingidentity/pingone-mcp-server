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

var UpdatePopulationByIdDef = types.ToolDefinition{
	IsReadOnly: false,
	McpTool: &mcp.Tool{
		Name:  "update_population_by_id",
		Title: "Update PingOne Population by ID",
		Description: `Update a population's configuration by its unique ID within a specified PingOne environment.

VERY IMPORTANT: Before updating, first get the latest population configuration using the 'get_population_by_id' tool to avoid overwriting pre-existing optional configuration values.`,
		InputSchema:  schema.MustGenerateSchema[UpdatePopulationByIdInput](),
		OutputSchema: schema.MustGenerateSchema[UpdatePopulationByIdOutput](),
	},
}

type UpdatePopulationByIdInput struct {
	EnvironmentId          uuid.UUID                            `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	PopulationId           uuid.UUID                            `json:"populationId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne population to update"`
	Name                   string                               `json:"name" jsonschema:"REQUIRED. The population name, which must be unique within the environment."`
	AlternativeIdentifiers []string                             `json:"alternativeIdentifiers,omitempty" jsonschema:"OPTIONAL. Alternative identifiers that can be used to search for populations besides name."`
	Description            *string                              `json:"description,omitempty" jsonschema:"OPTIONAL. A description of the population."`
	PreferredLanguage      *string                              `json:"preferredLanguage,omitempty" jsonschema:"OPTIONAL. The language locale for the population. If absent, the environment default is used."`
	PasswordPolicy         *management.PopulationPasswordPolicy `json:"passwordPolicy,omitempty" jsonschema:"OPTIONAL. The object reference to the password policy resource for the population."`
	Theme                  *management.PopulationTheme          `json:"theme,omitempty" jsonschema:"OPTIONAL. The object reference to the theme resource for the population."`
}

type UpdatePopulationByIdOutput struct {
	Population management.Population `json:"population" jsonschema:"The updated population configuration"`
}

// UpdatePopulationByIdHandler updates a PingOne population by ID using the provided client
func UpdatePopulationByIdHandler(populationsClientFactory PopulationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdatePopulationByIdInput,
) (
	*mcp.CallToolResult,
	*UpdatePopulationByIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdatePopulationByIdInput) (*mcp.CallToolResult, *UpdatePopulationByIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdatePopulationByIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdatePopulationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := populationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdatePopulationByIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Updating population",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("populationId", input.PopulationId.String()),
		)

		updateRequest := management.Population{
			Name:                   input.Name,
			AlternativeIdentifiers: input.AlternativeIdentifiers,
			Description:            input.Description,
			PreferredLanguage:      input.PreferredLanguage,
			PasswordPolicy:         input.PasswordPolicy,
			Theme:                  input.Theme,
		}

		// Call the API to update the population
		populationResponse, httpResponse, err := client.UpdatePopulationById(ctx, input.EnvironmentId, input.PopulationId, updateRequest)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if populationResponse == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no population data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Population updated successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("populationId", input.PopulationId.String()),
		)

		result := &UpdatePopulationByIdOutput{
			Population: *populationResponse,
		}

		return nil, result, nil
	}
}
