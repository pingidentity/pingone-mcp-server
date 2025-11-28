// Copyright © 2025 Ping Identity Corporation

package directory_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/directory"
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
	return pingone.DirectoryTotalIdentitiesCountCollectionResponse{
		Count: &count,
	}
}

// mockGetTotalIdentitiesByEnvironmentIdSetup configures a mock for GetTotalIdentitiesByEnvironmentId calls
func mockGetTotalIdentitiesByEnvironmentIdSetup(m *mockPingOneClientDirectoryWrapper, envID uuid.UUID, response *pingone.DirectoryTotalIdentitiesCountCollectionResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	m.On("GetTotalIdentitiesByEnvironmentId", mock.Anything, envID, mock.Anything).Return(response, httpResponse, err)
}

// calculateFilter calculates the filter string based on input dates
func calculateFilter(input directory.GetTotalIdentitiesByEnvironmentIdInput) string {
	var startDate, endDate time.Time

	if input.EndDate == nil {
		endDate = time.Now().UTC()
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.UTC)
	} else {
		endDate = input.EndDate.UTC()
	}

	if input.StartDate == nil {
		startDate = endDate.AddDate(0, 0, -32)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		startDate = input.StartDate.UTC()
	}

	startDateStr := startDate.Format("2006-01-02T15:04:05-07:00")
	endDateStr := endDate.Format("2006-01-02T15:04:05-07:00")

	return fmt.Sprintf("startDate eq \"%s\" and endDate eq \"%s\"", startDateStr, endDateStr)
}
