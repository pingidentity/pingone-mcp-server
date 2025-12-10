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

var UpdatePopulationDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:  "update_population",
		Title: "Update PingOne Population by ID",
		Description: `Update population configuration using full replacement (HTTP PUT).

WORKFLOW - Required to avoid data loss:
1. Call 'get_population' to fetch current configuration
2. Modify only the fields you want to change
3. Pass the complete merged object to this tool

Omitted optional fields will be cleared.`,
		InputSchema:  schema.MustGenerateSchema[UpdatePopulationInput](),
		OutputSchema: schema.MustGenerateSchema[UpdatePopulationOutput](),
	},
}

type UpdatePopulationInput struct {
	EnvironmentId          uuid.UUID                            `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	PopulationId           uuid.UUID                            `json:"populationId" jsonschema:"REQUIRED. Population UUID."`
	Name                   string                               `json:"name" jsonschema:"REQUIRED. Population name, must be unique within environment."`
	AlternativeIdentifiers []string                             `json:"alternativeIdentifiers,omitempty" jsonschema:"OPTIONAL. Alternative search identifiers."`
	Description            *string                              `json:"description,omitempty" jsonschema:"OPTIONAL. Description."`
	PreferredLanguage      *string                              `json:"preferredLanguage,omitempty" jsonschema:"OPTIONAL. Locale code. Defaults to environment setting if omitted."`
	PasswordPolicy         *management.PopulationPasswordPolicy `json:"passwordPolicy,omitempty" jsonschema:"OPTIONAL. Reference to password policy."`
	Theme                  *management.PopulationTheme          `json:"theme,omitempty" jsonschema:"OPTIONAL. Reference to theme."`
}

type UpdatePopulationOutput struct {
	Population management.Population `json:"population" jsonschema:"The updated population configuration"`
}

// UpdatePopulationHandler updates a PingOne population by ID using the provided client
func UpdatePopulationHandler(populationsClientFactory PopulationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdatePopulationInput,
) (
	*mcp.CallToolResult,
	*UpdatePopulationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdatePopulationInput) (*mcp.CallToolResult, *UpdatePopulationOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, UpdatePopulationDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdatePopulationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := populationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(UpdatePopulationDef.McpTool.Name, err)
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
		populationResponse, httpResponse, err := client.UpdatePopulation(ctx, input.EnvironmentId, input.PopulationId, updateRequest)
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

		// Filter out _links field from response
		populationResponse.Links = nil

		result := &UpdatePopulationOutput{
			Population: *populationResponse,
		}

		return nil, result, nil
	}
}
