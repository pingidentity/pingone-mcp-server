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

// mockAuthContextInitializer returns a mock auth context initializer that just returns the context unchanged
func mockAuthContextInitializer() func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	}
}

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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment not found")
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_ValidateEnvironment_ClientFactoryError(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockFactory := &mockEnvironmentsClientFactory{client: nil}

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get authenticated client")
}

func TestCachingEnvironmentValidator_ValidateEnvironment_NilEnvironmentResponse(t *testing.T) {
	ctx := context.Background()
	envId := uuid.New()

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	resp := &http.Response{StatusCode: 200}

	// API returns success but nil environment (should not happen in practice but code handles it)
	mockClient.On("GetEnvironmentById", ctx, envId).Return(nil, resp, nil)

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no environment data in response")
	mockClient.AssertExpectations(t)
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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Production Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_PRODUCTION,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	mockClient := new(mockEnvironmentsClient)
	mockFactory := &mockEnvironmentsClientFactory{client: mockClient}

	env := &pingone.EnvironmentResponse{
		Id:   envId,
		Name: "Sandbox Environment",
		Type: pingone.ENVIRONMENTTYPEVALUE_SANDBOX,
	}
	resp := &http.Response{StatusCode: 200}

	mockClient.On("GetEnvironmentById", ctx, envId).Return(env, resp, nil).Once()

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

	// Write operations should be allowed on SANDBOX environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeWrite)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCachingEnvironmentValidator_SandboxEnvironment_ReadAllowed(t *testing.T) {
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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

	// Read operations should be allowed on SANDBOX environments
	err := validator.ValidateEnvironment(ctx, envId, OperationTypeRead)
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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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

	validator := NewCachingEnvironmentValidator(mockFactory, mockAuthContextInitializer())

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
