// Copyright © 2025 Ping Identity Corporation

package directory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/initialize"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/schema"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

var GetTotalIdentitiesByEnvironmentDef = types.ToolDefinition{
	ValidationPolicy: &types.ToolValidationPolicy{
		AllowProductionEnvironmentRead: true,
	},
	McpTool: &mcp.Tool{
		Name:         "get_total_identities_by_environment",
		Title:        "Get Total Identities Count for PingOne Environment",
		Description:  "Retrieve the total count of user identities in a PingOne environment within a specified date range. Returns aggregated identity count data for reporting and analytics purposes. If no dates are provided, defaults to today at midnight UTC (showing a single day result). If specifying dates, at least one date must be provided.",
		InputSchema:  schema.MustGenerateSchema[GetTotalIdentitiesByEnvironmentInput](),
		OutputSchema: schema.MustGenerateSchema[GetTotalIdentitiesByEnvironmentOutput](),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	},
}

// GetTotalIdentitiesByEnvironmentInput defines the input parameters for retrieving total identities count by environment ID
type GetTotalIdentitiesByEnvironmentInput struct {
	EnvironmentId uuid.UUID  `json:"environmentId" jsonschema:"REQUIRED. The unique identifier (UUID) of the PingOne environment to query for identity counts."`
	StartDate     *time.Time `json:"startDate,omitempty" jsonschema:"OPTIONAL. The start date of the date range for counting identities in ISO 8601 format with timezone (e.g., '2024-01-01T00:00:00Z'). If neither date is provided, defaults to today at midnight UTC. If dates are specified, at least one must be set."`
	EndDate       *time.Time `json:"endDate,omitempty" jsonschema:"OPTIONAL. The end date of the date range for counting identities in ISO 8601 format with timezone (e.g., '2024-12-31T23:59:59Z'). If neither date is provided, remains unset. If dates are specified, at least one must be set."`
}

// GetTotalIdentitiesByEnvironmentOutput represents the result of retrieving total identities count for an environment
type GetTotalIdentitiesByEnvironmentOutput struct {
	TotalIdentitiesReport []GetTotalIdentitiesByEnvironmentOutputReport `json:"totalIdentitiesReport" jsonschema:"A list of total identities reports, by day, containing the aggregated number of user identities in the environment for the specified date range."`
}

type GetTotalIdentitiesByEnvironmentOutputReport struct {
	Date                 *string                `json:"date,omitempty" jsonschema:"The date and time the total identities count starts for the sampling period (ISO 8601 format)."`
	TotalIdentities      *int32                 `json:"totalIdentities,omitempty" jsonschema:"The total unique identities count for the sampling period."`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty"`
}

// GetTotalIdentitiesByEnvironmentHandler retrieves the total identities count for a PingOne environment within a specified date range using the provided client
func GetTotalIdentitiesByEnvironmentHandler(directoryClientFactory DirectoryClientFactory, initializeAuthContext initialize.ContextInitializer) func(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetTotalIdentitiesByEnvironmentInput,
) (
	*mcp.CallToolResult,
	*GetTotalIdentitiesByEnvironmentOutput,
	error,
) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetTotalIdentitiesByEnvironmentInput) (*mcp.CallToolResult, *GetTotalIdentitiesByEnvironmentOutput, error) {
		ctx = initialize.InitializeToolInvocation(ctx, GetTotalIdentitiesByEnvironmentDef.McpTool.Name, req)
		ctx, err := initializeAuthContext(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetTotalIdentitiesByEnvironmentDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		client, err := directoryClientFactory.GetAuthenticatedClient(ctx)
		if err != nil {
			toolErr := errs.NewToolError(GetTotalIdentitiesByEnvironmentDef.McpTool.Name, err)
			errs.Log(ctx, toolErr)
			return nil, nil, toolErr
		}

		logger.FromContext(ctx).Debug("Retrieving total identities by environment", slog.String("environmentId", input.EnvironmentId.String()))

		// Validate and apply default dates
		// If neither date is provided, default to today at midnight UTC with no endDate
		// If at least one date is provided, validation passes
		if input.StartDate == nil && input.EndDate == nil {
			// Default behavior: startDate = today at midnight UTC, endDate = nil (unset)
			now := time.Now().UTC()
			startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			input.StartDate = &startDate
			logger.FromContext(ctx).Debug("No dates provided, using default",
				slog.Time("startDate", startDate),
				slog.String("endDate", "unset"))
		}

		// Format dates for the API filter
		var filterBuilder strings.Builder
		var logAttrs []slog.Attr

		if input.StartDate != nil {
			fmt.Fprintf(&filterBuilder, "startDate eq \"%s\"", input.StartDate.UTC().Format("2006-01-02T15:04:05-07:00"))
			logAttrs = append(logAttrs, slog.Time("startDate", *input.StartDate))
		}

		if input.EndDate != nil {
			if filterBuilder.Len() > 0 {
				filterBuilder.WriteString(" and ")
			}
			fmt.Fprintf(&filterBuilder, "endDate eq \"%s\"", input.EndDate.UTC().Format("2006-01-02T15:04:05-07:00"))
			logAttrs = append(logAttrs, slog.Time("endDate", *input.EndDate))
		}

		filter := filterBuilder.String()
		logAttrs = append([]slog.Attr{slog.String("filter", filter)}, logAttrs...)
		logger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "Using filter", logAttrs...)

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

		totalIdentitiesReportOut := make([]GetTotalIdentitiesByEnvironmentOutputReport, 0, len(totalIdentitiesReport.Embedded.TotalIdentities))
		for _, report := range totalIdentitiesReport.Embedded.TotalIdentities {
			dateStr := report.Date.String()
			totalIdentitiesReportOut = append(totalIdentitiesReportOut, GetTotalIdentitiesByEnvironmentOutputReport{
				Date:                 &dateStr,
				TotalIdentities:      report.TotalIdentities,
				AdditionalProperties: report.AdditionalProperties,
			})
		}

		result := &GetTotalIdentitiesByEnvironmentOutput{
			TotalIdentitiesReport: totalIdentitiesReportOut,
		}

		return nil, result, nil
	}
}
