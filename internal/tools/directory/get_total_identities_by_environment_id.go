// Copyright © 2025 Ping Identity Corporation

package directory

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var GetTotalIdentitiesByEnvironmentIdDef = types.ToolDefinition{
	IsReadOnly: true,
	McpTool: &mcp.Tool{
		Name:         "get_total_identities_by_environment_id",
		Title:        "Get Total Identities Count for PingOne Environment",
		Description:  "Retrieve the total count of user identities in a PingOne environment within a specified date range. Returns aggregated identity count data for reporting and analytics purposes. If no dates are provided, defaults to a 32-day period ending today.",
		InputSchema:  schema.MustGenerateSchema[GetTotalIdentitiesByEnvironmentIdInput](),
		OutputSchema: schema.MustGenerateSchema[GetTotalIdentitiesByEnvironmentIdOutput](),
	},
}

// GetTotalIdentitiesByEnvironmentIdInput defines the input parameters for retrieving total identities count by environment ID
type GetTotalIdentitiesByEnvironmentIdInput struct {
	EnvironmentId uuid.UUID  `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) of the PingOne environment to query for identity counts."`
	StartDate     *time.Time `json:"startDate,omitempty" jsonschema:"OPTIONAL. The start date of the date range for counting identities in ISO 8601 format with timezone (e.g., '2024-01-01T00:00:00Z'). If not provided, defaults to 32 days before the end date."`
	EndDate       *time.Time `json:"endDate,omitempty" jsonschema:"OPTIONAL. The end date of the date range for counting identities in ISO 8601 format with timezone (e.g., '2024-12-31T23:59:59Z'). If not provided, defaults to today's date at 23:59:59 UTC."`
}

// GetTotalIdentitiesByEnvironmentIdOutput represents the result of retrieving total identities count for an environment
type GetTotalIdentitiesByEnvironmentIdOutput struct {
	TotalIdentitiesReport pingone.DirectoryTotalIdentitiesCountCollectionResponse `json:"totalIdentitiesReport" jsonschema:"The total identities count report containing the aggregated number of user identities in the environment for the specified date range, along with related metadata such as timestamps and environment information."`
}

// GetTotalIdentitiesByEnvironmentIdHandler retrieves the total identities count for a PingOne environment within a specified date range using the provided client
func GetTotalIdentitiesByEnvironmentIdHandler(directoryClientFactory DirectoryClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetTotalIdentitiesByEnvironmentIdInput,
) (
	*mcp.CallToolResult,
	*GetTotalIdentitiesByEnvironmentIdOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetTotalIdentitiesByEnvironmentIdInput) (*mcp.CallToolResult, *GetTotalIdentitiesByEnvironmentIdOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetTotalIdentitiesByEnvironmentIdDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetTotalIdentitiesByEnvironmentIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := directoryClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetTotalIdentitiesByEnvironmentIdDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving total identities by environment", slog.String("environmentId", input.EnvironmentId.String()))

		// Apply default dates if not provided
		var startDate, endDate time.Time

		if input.EndDate == nil {
			// Default to today at 23:59:59 UTC
			endDate = time.Now().UTC()
			// Set to end of day
			endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)
		} else {
			endDate = input.EndDate.UTC()
		}

		if input.StartDate == nil {
			// Default to 32 days before endDate at 00:00:00 UTC
			startDate = endDate.AddDate(0, 0, -32)
			// Set to start of day
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			startDate = input.StartDate.UTC()
		}

		// Format dates for the API filter
		startDateStr := startDate.Format("2006-01-02T15:04:05-07:00")
		endDateStr := endDate.Format("2006-01-02T15:04:05-07:00")

		filter := fmt.Sprintf("startDate eq \"%s\" and endDate eq \"%s\"", startDateStr, endDateStr)
		logger.FromContext(ctx).Debug("Using filter",
			slog.String("filter", filter),
			slog.Time("startDate", startDate),
			slog.Time("endDate", endDate))

		// Call the API to retrieve the totalIdentitiesReport
		totalIdentitiesReport, httpResponse, err := client.GetTotalIdentitiesByEnvironmentId(ctx, input.EnvironmentId, filter)
		logger.LogHttpResponse(ctx, httpResponse)

		if err != nil {
			apiErr := errs.NewApiError(httpResponse, err)
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		if totalIdentitiesReport == nil {
			apiErr := errs.NewApiError(httpResponse, fmt.Errorf("no total identities report data in response"))
			errs.Log(ctx, apiErr)
			return nil, nil, apiErr
		}

		logger.FromContext(ctx).Debug("Total identities report retrieved successfully",
			slog.String("environmentId", input.EnvironmentId.String()),
		)

		result := &GetTotalIdentitiesByEnvironmentIdOutput{
			TotalIdentitiesReport: *totalIdentitiesReport,
		}

		return nil, result, nil
	}
}
