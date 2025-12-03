// Copyright © 2025 Ping Identity Corporation

package environments_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
envtestutils "github.com/pingidentity/pingone-mcp-server/internal/tools/environments/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test data shared across all environment tests
var (
	testEnv1 = environmentTestData{
		name:    "Test Environment 1",
		region:  pingone.ENVIRONMENTREGIONCODE_NA,
		envType: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
		id:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
	}
	testEnv2 = environmentTestData{
		name:    "Test Environment 2",
		region:  pingone.ENVIRONMENTREGIONCODE_EU,
		envType: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
		id:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
	}
	testEnv3 = environmentTestData{
		name:    "Page 2 Environment 1",
		region:  pingone.ENVIRONMENTREGIONCODE_AP,
		envType: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
		id:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	}
	testEnv4 = environmentTestData{
		name:    "Page 3 Environment 1",
		region:  pingone.ENVIRONMENTREGIONCODE_CA,
		envType: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
		id:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
	}

	testLicenseID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440999")
)

// environmentTestData represents simplified test data for creating mock environments
type environmentTestData struct {
	name    string
	region  pingone.EnvironmentRegionCode
	envType pingone.EnvironmentTypeValue
	id      uuid.UUID
}

// Helper functions for test setup and assertions

// createEnvironmentResponse converts test data into a full EnvironmentResponse
func createEnvironmentResponse(t testing.TB, data environmentTestData) pingone.EnvironmentResponse {
	t.Helper()

	return pingone.EnvironmentResponse{
		Name:   data.name,
		Region: data.region,
		Type:   data.envType,
		Id:     data.id,
	}
}

// createMockPage creates a mock page response with the given environments
func createMockPage(t testing.TB, environments []environmentTestData) testutils.MockPage[pingone.EnvironmentsCollectionResponse] {
	t.Helper()

	envResponses := make([]pingone.EnvironmentResponse, len(environments))
	for i, env := range environments {
		envResponses[i] = createEnvironmentResponse(t, env)
	}

	return testutils.MockPage[pingone.EnvironmentsCollectionResponse]{
		Data: &pingone.EnvironmentsCollectionResponse{
			Embedded: &pingone.EnvironmentsCollectionResponseEmbedded{
				Environments: envResponses,
			},
		},
		HTTPResponse: &http.Response{StatusCode: http.StatusOK},
	}
}

// assertEnvironmentMatches verifies that an actual environment matches the expected test data
func assertEnvironmentMatches(t *testing.T, expected environmentTestData, actual pingone.EnvironmentResponse) {
	t.Helper()

	assert.Equal(t, expected.name, actual.Name, "Environment name should match")
	assert.Equal(t, expected.id, actual.Id, "Environment ID should match")
	assert.Equal(t, expected.region, actual.Region, "Environment region should match")
	assert.Equal(t, expected.envType, actual.Type, "Environment type should match")
}

// assertCreateEnvironmentOutput verifies that a CreateEnvironmentOutput matches the input used to create it
func assertCreateEnvironmentOutput(t *testing.T, input environments.CreateEnvironmentInput, output *environments.CreateEnvironmentOutput) {
	t.Helper()

	require.NotNil(t, output, "Output should not be nil")
	assert.Equal(t, input.Name, output.Environment.Name, "Environment name should match input")
	assert.Equal(t, input.Region, output.Environment.Region, "Environment region should match input")

	// Validate optional fields
	if input.Description != nil {
		require.NotNil(t, output.Environment.Description, "Description should not be nil when provided in input")
		assert.Equal(t, *input.Description, *output.Environment.Description, "Description should match input")
	}

	if input.Icon != nil {
		require.NotNil(t, output.Environment.Icon, "Icon should not be nil when provided in input")
		assert.Equal(t, *input.Icon, *output.Environment.Icon, "Icon should match input")
	}
}

// mockListEnvironmentsSetup creates a setup function for mocking GetEnvironments calls.
// If err is not nil, it configures the mock to return that error.
// Otherwise, it configures the mock to return a paginated response with the provided pageData.
func mockListEnvironmentsSetup(t *testing.T, err error, pageData ...[]environmentTestData) func(*envtestutils.MockEnvironmentsClient, *string) {
	t.Helper()

	return func(m *envtestutils.MockEnvironmentsClient, filter *string) {
		if err != nil {
			m.On("GetEnvironments", mock.Anything, filter).Return(nil, err)
			return
		}

		pages := make([]testutils.MockPage[pingone.EnvironmentsCollectionResponse], len(pageData))
		for i, data := range pageData {
			pages[i] = createMockPage(t, data)
		}

		m.On("GetEnvironments", mock.Anything, filter).
			Return(testutils.MockPaginationIterator(pages), nil)
	}
}

// mockGetEnvironmentByIdSetup configures a mock for GetEnvironmentById calls
func mockGetEnvironmentByIdSetup(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID, response *pingone.EnvironmentResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	m.On("GetEnvironmentById", mock.Anything, envID).Return(response, httpResponse, err)
}

// mockCreateEnvironmentSetup configures a mock for CreateEnvironment calls with a matcher function.
// If matcher is provided and MatchedBy returns false, the mock will not match the call,
// causing the test to fail with an "unexpected method call" error from testify/mock.
// This is the intended behavior - it ensures tests validate the exact request parameters.
func mockCreateEnvironmentSetup(m *envtestutils.MockEnvironmentsClient, matcher func(*pingone.EnvironmentCreateRequest) bool, response *pingone.EnvironmentResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	if matcher != nil {
		m.On("CreateEnvironment", mock.Anything, mock.MatchedBy(matcher)).Return(response, httpResponse, err)
	} else {
		m.On("CreateEnvironment", mock.Anything, mock.Anything).Return(response, httpResponse, err)
	}
}

// mockUpdateEnvironmentByIdSetup configures a mock for UpdateEnvironmentById calls with a matcher function
func mockUpdateEnvironmentByIdSetup(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID, matcher func(*pingone.EnvironmentReplaceRequest) bool, response *pingone.EnvironmentResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	if matcher != nil {
		m.On("UpdateEnvironmentById", mock.Anything, envID, mock.MatchedBy(matcher)).Return(response, httpResponse, err)
	} else {
		m.On("UpdateEnvironmentById", mock.Anything, envID, mock.Anything).Return(response, httpResponse, err)
	}
}

// createEnvironmentServicesResponse creates a mock environment services response for testing
func createEnvironmentServicesResponse(t testing.TB) pingone.EnvironmentBillOfMaterialsResponse {
	t.Helper()

	return pingone.EnvironmentBillOfMaterialsResponse{
		Products: []pingone.EnvironmentBillOfMaterialsProduct{
			{
				Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_BASE,
			},
			{
				Type: pingone.ENVIRONMENTBILLOFMATERIALSPRODUCTTYPE_PING_ONE_MFA,
			},
		},
	}
}

// mockGetEnvironmentServicesByIdSetup configures a mock for GetEnvironmentServicesById calls
func mockGetEnvironmentServicesByIdSetup(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID, response *pingone.EnvironmentBillOfMaterialsResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	m.On("GetEnvironmentServicesById", mock.Anything, envID).Return(response, httpResponse, err)
}

// mockUpdateEnvironmentServicesByIdSetup configures a mock for UpdateEnvironmentServicesById calls with a matcher function
func mockUpdateEnvironmentServicesByIdSetup(m *envtestutils.MockEnvironmentsClient, envID uuid.UUID, matcher func(*pingone.EnvironmentBillOfMaterialsReplaceRequest) bool, response *pingone.EnvironmentBillOfMaterialsResponse, statusCode int, err error) {
	httpResponse := &http.Response{StatusCode: statusCode}
	if matcher != nil {
		m.On("UpdateEnvironmentServicesById", mock.Anything, envID, mock.MatchedBy(matcher)).Return(response, httpResponse, err)
	} else {
		m.On("UpdateEnvironmentServicesById", mock.Anything, envID, mock.Anything).Return(response, httpResponse, err)
	}
}
