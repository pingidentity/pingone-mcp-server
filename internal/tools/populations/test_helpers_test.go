// Copyright © 2025 Ping Identity Corporation

package populations_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testEnvironmentId = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

var (
	testPop1 = management.Population{
		Name:        "Test Population 1",
		Id:          testutils.Pointer(testEnvironmentId.String()),
		Description: testutils.Pointer("This is a test population"),
	}
	testPop2OnlyRequiredFields = management.Population{
		Name: "Test Population 2 - Only Required Fields",
		Id:   testutils.Pointer(testEnvironmentId.String()),
	}
	testPop3 = management.Population{
		Name:        "Test Population 3",
		Id:          testutils.Pointer(testEnvironmentId.String()),
		Description: testutils.Pointer("This is yet another test population"),
	}
	testPop4 = management.Population{
		Name: "Test Population 4",
		Id:   testutils.Pointer(testEnvironmentId.String()),
	}
	testPop5AllFields = management.Population{
		Name:                   "Test Population 5 - All Fields",
		Id:                     testutils.Pointer(testEnvironmentId.String()),
		Description:            testutils.Pointer("Population with all fields set"),
		AlternativeIdentifiers: []string{"alt-id-1", "alt-id-2"},
		PreferredLanguage:      testutils.Pointer("en-US"),
		PasswordPolicy: &management.PopulationPasswordPolicy{
			Id: "password-policy-id",
		},
		Theme: &management.PopulationTheme{
			Id: testutils.Pointer("theme-id"),
		},
	}
)

// assertCreatePopulationOutput verifies that a CreatePopulationOutput matches the input used to create it
func assertCreatePopulationOutput(t *testing.T, input populations.CreatePopulationInput, output *populations.CreatePopulationOutput) {
	t.Helper()

	require.NotNil(t, output, "Output should not be nil")
	assert.Equal(t, input.Name, output.Population.Name, "Population name should match input")

	// Validate optional fields
	if input.Description != nil {
		require.NotNil(t, output.Population.Description, "Description should not be nil when provided in input")
		assert.Equal(t, *input.Description, *output.Population.Description, "Description should match input")
	}

	if input.PreferredLanguage != nil {
		require.NotNil(t, output.Population.PreferredLanguage, "PreferredLanguage should not be nil when provided in input")
		assert.Equal(t, *input.PreferredLanguage, *output.Population.PreferredLanguage, "PreferredLanguage should match input")
	}

	assert.Equal(t, input.AlternativeIdentifiers, output.Population.AlternativeIdentifiers, "AlternativeIdentifiers should match input")

	if input.PasswordPolicy != nil {
		require.NotNil(t, output.Population.PasswordPolicy, "PasswordPolicy should not be nil when provided in input")
		assert.Equal(t, input.PasswordPolicy.Id, output.Population.PasswordPolicy.Id, "PasswordPolicy ID should match input")
	}

	if input.Theme != nil {
		require.NotNil(t, output.Population.Theme, "Theme should not be nil when provided in input")
		if input.Theme.Id != nil {
			require.NotNil(t, output.Population.Theme.Id, "Theme ID should not be nil when provided in input")
			assert.Equal(t, *input.Theme.Id, *output.Population.Theme.Id, "Theme ID should match input")
		}
	}
}

// assertPopulationMatches verifies that a Population matches the expected, for fields that can be set via the API
func assertPopulationMatches(t *testing.T, expected management.Population, actual management.Population) {
	t.Helper()

	assert.Equal(t, expected.Name, actual.Name, "Population name should match")
	require.NotNil(t, actual.Id, "Population ID should not be nil")
	assert.Equal(t, *expected.Id, *actual.Id, "Population ID should match")

	if expected.Description != nil {
		require.NotNil(t, actual.Description, "Description should not be nil when expected")
		assert.Equal(t, *expected.Description, *actual.Description, "Description should match")
	} else {
		assert.Nil(t, actual.Description, "Description should be nil when not expected")
	}

	assert.Equal(t, expected.AlternativeIdentifiers, actual.AlternativeIdentifiers, "AlternativeIdentifiers should match")

	if expected.PreferredLanguage != nil {
		require.NotNil(t, actual.PreferredLanguage, "PreferredLanguage should not be nil when expected")
		assert.Equal(t, *expected.PreferredLanguage, *actual.PreferredLanguage, "PreferredLanguage should match")
	} else {
		assert.Nil(t, actual.PreferredLanguage, "PreferredLanguage should be nil when not expected")
	}
}

// assertPopulationSummaryMatches verifies that a Population summary matches the expected, for fields that can be set via the API
func assertPopulationSummaryMatches(t *testing.T, expected management.Population, actual populations.PopulationSummary) {
	t.Helper()

	assert.Equal(t, expected.Name, actual.Name, "Population name should match")
	require.NotNil(t, actual.Id, "Population ID should not be nil")
	assert.Equal(t, *expected.Id, *actual.Id, "Population ID should match")

	if expected.Default != nil {
		require.NotNil(t, actual.Default, "Default should not be nil when expected")
		assert.Equal(t, *expected.Default, *actual.Default, "Default should match")
	} else {
		assert.Nil(t, actual.Default, "Default should be nil when not expected")
	}

	if expected.CreatedAt != nil {
		require.NotNil(t, actual.CreatedAt, "CreatedAt should not be nil when expected")
		assert.Equal(t, *expected.CreatedAt, *actual.CreatedAt, "CreatedAt should match")
	} else {
		assert.Nil(t, actual.CreatedAt, "CreatedAt should be nil when not expected")
	}
}

func updatePopulationByIdInputFromPopulation(pop management.Population, envID uuid.UUID) populations.UpdatePopulationByIdInput {
	return populations.UpdatePopulationByIdInput{
		EnvironmentId:          envID,
		PopulationId:           uuid.MustParse(*pop.Id),
		Name:                   pop.Name,
		Description:            pop.Description,
		AlternativeIdentifiers: pop.AlternativeIdentifiers,
		PreferredLanguage:      pop.PreferredLanguage,
		PasswordPolicy:         pop.PasswordPolicy,
		Theme:                  pop.Theme,
	}
}

func createPopulationInputFromPopulation(pop management.Population, envID uuid.UUID) populations.CreatePopulationInput {
	return populations.CreatePopulationInput{
		EnvironmentId:          envID,
		Name:                   pop.Name,
		Description:            pop.Description,
		AlternativeIdentifiers: pop.AlternativeIdentifiers,
		PreferredLanguage:      pop.PreferredLanguage,
		PasswordPolicy:         pop.PasswordPolicy,
		Theme:                  pop.Theme,
	}
}

func getPopulationByIdInputFromPopulation(pop management.Population, envID uuid.UUID) populations.GetPopulationByIdInput {
	return populations.GetPopulationByIdInput{
		EnvironmentId: envID,
		PopulationId:  uuid.MustParse(*pop.Id),
	}
}
