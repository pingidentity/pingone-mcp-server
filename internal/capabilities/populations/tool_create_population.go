// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

var CreatePopulationDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:         "create_population",
		Title:        "Create PingOne Population",
		Description:  "Create a population in an environment. Populations group users logically in an environment and allow per population customization of branding theme, password policy, preferred language. Only 'name' and 'environmentId' are required.",
		InputSchema:  schema.MustGenerateSchema[CreatePopulationInput](),
		OutputSchema: schema.MustGenerateSchema[CreatePopulationOutput](),
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: func() *bool { b := false; return &b }(),
		},
	},
}

type CreatePopulationInput struct {
	EnvironmentId          uuid.UUID                            `json:"environmentId" jsonschema:"REQUIRED. Environment UUID."`
	Name                   string                               `json:"name" jsonschema:"REQUIRED. Population name, must be unique within environment."`
	AlternativeIdentifiers []string                             `json:"alternativeIdentifiers,omitempty" jsonschema:"OPTIONAL. Alternative search identifiers."`
	Description            *string                              `json:"description,omitempty" jsonschema:"OPTIONAL. Description."`
	PreferredLanguage      *string                              `json:"preferredLanguage,omitempty" jsonschema:"OPTIONAL. Locale code (e.g., 'en', 'fr'). Defaults to environment setting if omitted."`
	PasswordPolicy         *management.PopulationPasswordPolicy `json:"passwordPolicy,omitempty" jsonschema:"OPTIONAL. Reference to password policy."`
	Theme                  *management.PopulationTheme          `json:"theme,omitempty" jsonschema:"OPTIONAL. Reference to theme."`
}

type CreatePopulationOutput struct {
	Population management.Population `json:"population" jsonschema:"The created population details including ID, name, and description"`
}

// CreatePopulationHandler creates a new PingOne population using the provided client
func CreatePopulationHandler(populationsClientFactory PopulationsClientFactory) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreatePopulationInput,
) (
	*mcp.CallToolResult,
	*CreatePopulationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreatePopulationInput) (*mcp.CallToolResult, *CreatePopulationOutput, error) {
		client, err := populationsClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(CreatePopulationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Creating population",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("name", input.Name),
		)

		createRequest := management.Population{
			Name:                   input.Name,
			AlternativeIdentifiers: input.AlternativeIdentifiers,
			Description:            input.Description,
			PreferredLanguage:      input.PreferredLanguage,
			PasswordPolicy:         input.PasswordPolicy,
			Theme:                  input.Theme,
		}

		// Call the API to create the population
		populationResponse, httpResponse, err := client.CreatePopulation(ctx, input.EnvironmentId, createRequest)
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

		if populationResponse.Id == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("created population has no ID"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Population created successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
			slog.String("populationId", *populationResponse.Id),
			slog.String("name", populationResponse.Name))

		// Filter out _links field from response
		populationResponse.Links = nil

		result := &CreatePopulationOutput{
			Population: *populationResponse,
		}

		return nil, result, nil
	}
}
