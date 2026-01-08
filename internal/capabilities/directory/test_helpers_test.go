// Copyright © 2025 Ping Identity Corporation

package directory_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-go-client/types"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/directory"
	"github.com/stretchr/testify/mock"
)

// Test data shared across all tests
var (
	testEnvId     = uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	testStartDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testEndDate   = time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
)

// Helper functions for test setup

// createTotalIdentitiesResponse creates a mock total identities response for testing
func createTotalIdentitiesResponse(t testing.TB) pingone.DirectoryTotalIdentitiesCountCollectionResponse {
	t.Helper()

	count := int32(100)
	size := int32(1)
	totalCount := int32(100)
	date := types.UnixTime{Time: time.Now().UTC()}

	// Create the inner response with identity count
	totalIdentity := pingone.DirectoryTotalIdentitiesCountResponse{
		Date:            &date,
		TotalIdentities: &totalCount,
	}

	// Create the embedded structure
	embedded := pingone.DirectoryTotalIdentitiesCountCollectionResponseEmbedded{
		TotalIdentities: []pingone.DirectoryTotalIdentitiesCountResponse{totalIdentity},
	}

	return pingone.DirectoryTotalIdentitiesCountCollectionResponse{
		Count:    &count,
		Size:     &size,
		Embedded: &embedded,
	}
}

// mockGetTotalIdentitiesByEnvironmentSetup configures a mock for GetTotalIdentitiesByEnvironmentId calls
func mockGetTotalIdentitiesByEnvironmentSetup(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, response *pingone.DirectoryTotalIdentitiesCountCollectionResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	m.On("GetTotalIdentitiesByEnvironmentId", mock.Anything, envID, mock.Anything).Return(response, httpResponse, err)
}

// calculateFilter calculates the filter string based on input dates
// Matches the logic in GetTotalIdentitiesByEnvironmentIdHandler
func calculateFilter(input directory.GetTotalIdentitiesByEnvironmentInput) string {
	// If neither date is provided, default to today at midnight UTC with no endDate
	if input.StartDate == nil && input.EndDate == nil {
		now := time.Now().UTC()
		startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		startDateStr := startDate.Format("2006-01-02T15:04:05-07:00")
		return fmt.Sprintf("startDate eq \"%s\"", startDateStr)
	}

	// If both dates are provided
	if input.StartDate != nil && input.EndDate != nil {
		startDateStr := input.StartDate.UTC().Format("2006-01-02T15:04:05-07:00")
		endDateStr := input.EndDate.UTC().Format("2006-01-02T15:04:05-07:00")
		return fmt.Sprintf("startDate eq \"%s\" and endDate eq \"%s\"", startDateStr, endDateStr)
	}

	// If only startDate is provided
	if input.StartDate != nil {
		startDateStr := input.StartDate.UTC().Format("2006-01-02T15:04:05-07:00")
		return fmt.Sprintf("startDate eq \"%s\"", startDateStr)
	}

	// If only endDate is provided
	if input.EndDate != nil {
		endDateStr := input.EndDate.UTC().Format("2006-01-02T15:04:05-07:00")
		return fmt.Sprintf("endDate eq \"%s\"", endDateStr)
	}

	return ""
}
