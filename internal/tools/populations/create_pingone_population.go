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

var CreatePopulationDef = types.ToolDefinition{
	McpTool: &mcp.Tool{
		Name:         "create_population",
		Title:        "Create PingOne Population",
		Description:  "Create a new PingOne population within a specified PingOne environment.",
		InputSchema:  schema.MustGenerateSchema[CreatePopulationInput](),
		OutputSchema: schema.MustGenerateSchema[CreatePopulationOutput](),
	},
}

type CreatePopulationInput struct {
	EnvironmentId          uuid.UUID                            `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) string of the PingOne environment"`
	Name                   string                               `json:"name" jsonschema:"REQUIRED. The population name, which must be unique within the environment."`
	AlternativeIdentifiers []string                             `json:"alternativeIdentifiers,omitempty" jsonschema:"OPTIONAL. Alternative identifiers that can be used to search for populations besides name."`
	Description            *string                              `json:"description,omitempty" jsonschema:"OPTIONAL. A description of the population."`
	PreferredLanguage      *string                              `json:"preferredLanguage,omitempty" jsonschema:"OPTIONAL. The language locale for the population. If absent, the environment default is used."`
	PasswordPolicy         *management.PopulationPasswordPolicy `json:"passwordPolicy,omitempty" jsonschema:"OPTIONAL. The object reference to the password policy resource for the population."`
	Theme                  *management.PopulationTheme          `json:"theme,omitempty" jsonschema:"OPTIONAL. The object reference to the theme resource for the population."`
}

type CreatePopulationOutput struct {
	Population management.Population `json:"population" jsonschema:"The created population details including ID, name, and description"`
}

// CreatePopulationHandler creates a new PingOne population using the provided client
func CreatePopulationHandler(populationsClientFactory PopulationsClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreatePopulationInput,
) (
	*mcp.CallToolResult,
	*CreatePopulationOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreatePopulationInput) (*mcp.CallToolResult, *CreatePopulationOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, CreatePopulationDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(CreatePopulationDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

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

		result := &CreatePopulationOutput{
			Population: *populationResponse,
		}

		return nil, result, nil
	}
}
