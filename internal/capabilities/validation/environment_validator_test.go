// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/environments/testutils"
	"github.com/stretchr/testify/assert"
)

type mockEnvironmentsClientFactory struct {
	client *testutils.MockEnvironmentsClient
}

// GetAuthenticatedClient returns the pre-configured mock client or an error if client is nil.
// This method implements the EnvironmentsClientFactory interface for testing purposes.
// The ctx parameter provides context for the authentication operation (not used in mock).
func (m *mockEnvironmentsClientFactory) GetAuthenticatedClient(ctx context.Context) (environments.EnvironmentsClient, error) {
	if m.client == nil {
		return nil, errors.New("client not initialized")
	}
	return m.client, nil
}

func TestCachingEnvironmentValidator_ValidateEnvironment_Success(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Test Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	// SANDBOX environments are not cached, so expect multiple API calls
	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Times(3)

	validator := NewCachingEnvironmentValidator(mockFactory)

	// First call should hit the API (read operation)
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Second call should also hit the API (SANDBOX not cached)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Write operation should also hit the API
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ValidateEnvironment_NotFound(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	resp := &http.Response{StatusCode: 404}
	apiErr := errors.New("environment not found")

	mockClient.On("GetEnvironment", ctx, envId).Return(nil, resp, apiErr)

	validator := NewCachingEnvironmentValidator(mockFactory)

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment not found")
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ValidateEnvironment_ClientFactoryError(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockFactory := &mockEnvironmentsClientFactory{client: nil}

	validator := NewCachingEnvironmentValidator(mockFactory)

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
}

func TestCachingEnvironmentValidator_ValidateEnvironment_NilEnvironmentResponse(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	resp := &http.Response{StatusCode: 200}

	// API returns success but nil environment (should not happen in practice but code handles it)
	mockClient.On("GetEnvironment", ctx, envId).Return(nil, resp, nil)

	validator := NewCachingEnvironmentValidator(mockFactory)

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no environment data in response")
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ClearCache(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	// Use PRODUCTION environment since only PRODUCTION is cached
	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// Expect two API calls since we'll clear cache
	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Twice()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Validate to populate cache (READ will be blocked on PRODUCTION)
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")

	// Clear cache
	validator.ClearCache()

	// After clearing cache, should hit API again (and still be blocked)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_RemoveFromCache(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	// Use PRODUCTION environment since only PRODUCTION is cached
	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// Expect two API calls since we'll remove from cache
	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Twice()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Validate to populate cache (READ will be blocked on PRODUCTION)
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")

	// Remove specific environment from cache
	validator.RemoveFromCache(envId)

	// After removing from cache, should hit API again (and still be blocked)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_ReadBlocked(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Read operations should NOT be allowed on PRODUCTION environments by default
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")
	assert.Contains(t, err.Error(), envId.String())
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_WriteBlocked(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Write operations should NOT be allowed on PRODUCTION environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended or breaking changes, this write operation is not allowed against PRODUCTION environments")
	assert.Contains(t, err.Error(), envId.String())
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_WriteAllowed(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Write operations should be allowed on SANDBOX environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_ReadAllowed(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Read operations should be allowed on SANDBOX environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_NotCached(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	// SANDBOX environments should NOT be cached, expect API call each time
	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Twice()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// First read operation
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Second read operation should make another API call (not cached)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Verify both API calls were made
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_IsCached(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(testutils.MockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// PRODUCTION environments should be cached, expect only one API call
	mockClient.On("GetEnvironment", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// First read operation - populates cache (and gets blocked)
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")

	// Second read operation should use cache (no additional API call, also blocked)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended access to sensitive data or configuration, this read operation is not allowed against PRODUCTION environments")

	// Third read operation should also use cache
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)

	// Write operation should use cached data and block
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended or breaking changes, this write operation is not allowed against PRODUCTION environments")

	// Verify only one API call was made (all subsequent calls used cache)
	mockClient.AssertExpectations(t)
}
