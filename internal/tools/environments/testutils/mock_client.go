// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/mock"
)

var _ environments.EnvironmentsClient = &MockEnvironmentsClient{}
var _ environments.EnvironmentsClientFactory = &MockEnvironmentsClientFactory{}

// MockEnvironmentsClient is a mock implementation of the EnvironmentsClient interface.
// It provides test doubles for all environments API operations and uses testify/mock
// for flexible test assertions and behavior configuration.
type MockEnvironmentsClient struct {
	mock.Mock
}

// MockEnvironmentsClientFactory is a mock implementation of the EnvironmentsClientFactory interface.
// It returns a pre-configured mock client and optional error for testing authentication flows.
type MockEnvironmentsClientFactory struct {
	Client environments.EnvironmentsClient
	Err    error
}

// NewMockEnvironmentsClientFactory creates a new MockEnvironmentsClientFactory with the provided client and error.
// The client parameter is the mock client instance to return from GetAuthenticatedClient.
// The err parameter is the error to return from GetAuthenticatedClient, or nil for successful authentication.
func NewMockEnvironmentsClientFactory(client environments.EnvironmentsClient, err error) *MockEnvironmentsClientFactory {
	return &MockEnvironmentsClientFactory{
		Client: client,
		Err:    err,
	}
}

// GetAuthenticatedClient returns the pre-configured mock client and error.
// This method implements the EnvironmentsClientFactory interface for testing purposes.
// The ctx parameter provides context for the authentication operation (not used in mock).
func (m *MockEnvironmentsClientFactory) GetAuthenticatedClient(ctx context.Context) (environments.EnvironmentsClient, error) {
	if m.Client == nil && m.Err == nil {
		return nil, errors.New("client not initialized")
	}
	return m.Client, m.Err
}

// GetEnvironments retrieves a paginated list of environments matching the optional filter.
// Returns a PagedIterator for accessing environment collection responses and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The filter parameter specifies optional filter criteria to limit the environments returned.
func (m *MockEnvironmentsClient) GetEnvironments(ctx context.Context, filter *string) (pingone.PagedIterator[pingone.EnvironmentsCollectionResponse], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pingone.PagedIterator[pingone.EnvironmentsCollectionResponse]), args.Error(1)
}

// CreateEnvironment creates a new PingOne environment with the provided configuration.
// Returns the created environment response, HTTP response details, and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The request parameter contains the environment configuration including name, type, and license.
func (m *MockEnvironmentsClient) CreateEnvironment(ctx context.Context, request *pingone.EnvironmentCreateRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, request)
	var envResponse *pingone.EnvironmentResponse
	if args.Get(0) != nil {
		envResponse = args.Get(0).(*pingone.EnvironmentResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return envResponse, httpResponse, args.Error(2)
}

// GetEnvironment retrieves a specific environment by its unique identifier.
// Returns the environment response, HTTP response details, and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The environmentId parameter specifies the UUID of the environment to retrieve.
func (m *MockEnvironmentsClient) GetEnvironment(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId)
	var envResponse *pingone.EnvironmentResponse
	if args.Get(0) != nil {
		envResponse = args.Get(0).(*pingone.EnvironmentResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return envResponse, httpResponse, args.Error(2)
}

// UpdateEnvironment updates an existing environment with new configuration.
// Returns the updated environment response, HTTP response details, and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The environmentId parameter specifies the UUID of the environment to update.
// The request parameter contains the updated environment configuration.
func (m *MockEnvironmentsClient) UpdateEnvironment(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentReplaceRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId, request)
	var envResponse *pingone.EnvironmentResponse
	if args.Get(0) != nil {
		envResponse = args.Get(0).(*pingone.EnvironmentResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return envResponse, httpResponse, args.Error(2)
}

// GetEnvironmentServices retrieves the bill of materials (enabled services) for an environment.
// Returns the bill of materials response, HTTP response details, and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The environmentId parameter specifies the UUID of the environment whose services to retrieve.
func (m *MockEnvironmentsClient) GetEnvironmentServices(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId)
	var bomResponse *pingone.EnvironmentBillOfMaterialsResponse
	if args.Get(0) != nil {
		bomResponse = args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return bomResponse, httpResponse, args.Error(2)
}

// UpdateEnvironmentServices updates the enabled services (bill of materials) for an environment.
// Returns the updated bill of materials response, HTTP response details, and any error encountered.
// The ctx parameter provides context for the API operation including cancellation and timeouts.
// The environmentId parameter specifies the UUID of the environment whose services to update.
// The request parameter contains the updated bill of materials configuration.
func (m *MockEnvironmentsClient) UpdateEnvironmentServices(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentBillOfMaterialsReplaceRequest) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := m.Called(ctx, environmentId, request)
	var bomResponse *pingone.EnvironmentBillOfMaterialsResponse
	if args.Get(0) != nil {
		bomResponse = args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return bomResponse, httpResponse, args.Error(2)
}
