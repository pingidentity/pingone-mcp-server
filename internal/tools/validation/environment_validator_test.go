// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEnvironmentsClient struct {
	mock.Mock
}

func (m *mockEnvironmentsClient) GetEnvironments(ctx context.Context, filter *string) (pingone.PagedIterator[pingone.EnvironmentsCollectionResponse], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pingone.PagedIterator[pingone.EnvironmentsCollectionResponse]), args.Error(1)
}

func (m *mockEnvironmentsClient) CreateEnvironment(ctx context.Context, request *pingone.EnvironmentCreateRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*http.Response), args.Error(2)
	}
	return args.Get(0).(*pingone.EnvironmentResponse), args.Get(1).(*http.Response), args.Error(2)
}

func (m *mockEnvironmentsClient) GetEnvironmentById(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*http.Response), args.Error(2)
	}
	return args.Get(0).(*pingone.EnvironmentResponse), args.Get(1).(*http.Response), args.Error(2)
}

func (m *mockEnvironmentsClient) UpdateEnvironmentById(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentReplaceRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*http.Response), args.Error(2)
	}
	return args.Get(0).(*pingone.EnvironmentResponse), args.Get(1).(*http.Response), args.Error(2)
}

func (m *mockEnvironmentsClient) GetEnvironmentServicesById(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*http.Response), args.Error(2)
	}
	return args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse), args.Get(1).(*http.Response), args.Error(2)
}

func (m *mockEnvironmentsClient) UpdateEnvironmentServicesById(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentBillOfMaterialsReplaceRequest) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId, request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*http.Response), args.Error(2)
	}
	return args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse), args.Get(1).(*http.Response), args.Error(2)
}

type mockEnvironmentsClientFactory struct {
	client *mockEnvironmentsClient
}

func (m *mockEnvironmentsClientFactory) GetAuthenticatedClient(ctx context.Context) (environments.EnvironmentsClient, error) {
	if m.client == nil {
		return nil, errors.New("client not initialized")
	}
	return m.client, nil
}

func TestCachingEnvironmentValidator_ValidateEnvironment_Success(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Test Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	// SANDBOX environments are not cached, so expect multiple API calls
	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Times(3)

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

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	resp := &http.Response{StatusCode: 404}
	apiErr := errors.New("environment not found")

	mockClient.On("GetEnvironmentById", ctx, envId).Return(nil, resp, apiErr)

	validator := NewCachingEnvironmentValidator(mockFactory)

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist or is not accessible")
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

func TestCachingEnvironmentValidator_ClearCache(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	// Use PRODUCTION environment since only PRODUCTION is cached
	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// Expect two API calls since we'll clear cache
	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Twice()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Validate to populate cache
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Clear cache
	validator.ClearCache()

	// After clearing cache, should hit API again
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_RemoveFromCache(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	// Use PRODUCTION environment since only PRODUCTION is cached
	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// Expect two API calls since we'll remove from cache
	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Twice()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Validate to populate cache
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Remove specific environment from cache
	validator.RemoveFromCache(envId)

	// After removing from cache, should hit API again
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_ReadAllowed(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Read operations should be allowed on PRODUCTION environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_WriteBlocked(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Write operations should NOT be allowed on PRODUCTION environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed")
	assert.Contains(t, err.Error(), envId.String())
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ProductionEnvironment_WriteCached(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// First read operation caches the environment
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Write operation should use cached data and still block
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed")

	// Only one API call should have been made
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_WriteAllowed(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// Write operations should be allowed on SANDBOX environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_NotCached(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	// SANDBOX environments should NOT be cached, expect API call each time
	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Twice()

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

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	// PRODUCTION environments should be cached, expect only one API call
	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory)

	// First read operation
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Second read operation should use cache (no additional API call)
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Third read operation should also use cache
	err = validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.NoError(t, err)

	// Verify only one API call was made
	mockClient.AssertExpectations(t)
}
