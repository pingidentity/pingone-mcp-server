// Copyright © 2025 Ping Identity Corporation

package environments_test

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/stretchr/testify/mock"
)

var _ environments.EnvironmentsClient = &mockPingOneClientEnvironmentsWrapper{}
var _ environments.EnvironmentsClientFactory = &mockPingOneClientEnvironmentsWrapperFactory{}

type mockPingOneClientEnvironmentsWrapper struct {
	mock.Mock
}

type mockPingOneClientEnvironmentsWrapperFactory struct {
	mockClient environments.EnvironmentsClient
	err        error
}

// Directly returns the provided mock client and error
func NewMockPingOneClientEnvironmentsWrapperFactory(mockClient environments.EnvironmentsClient, err error) *mockPingOneClientEnvironmentsWrapperFactory {
	return &mockPingOneClientEnvironmentsWrapperFactory{
		mockClient: mockClient,
		err:        err,
	}
}

func (f *mockPingOneClientEnvironmentsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (environments.EnvironmentsClient, error) {
	return f.mockClient, f.err
}

func (p *mockPingOneClientEnvironmentsWrapper) GetEnvironments(ctx context.Context, filter *string) (pingone.PagedIterator[pingone.EnvironmentsCollectionResponse], error) {
	args := p.Called(ctx, filter)
	var response pingone.PagedIterator[pingone.EnvironmentsCollectionResponse]
	response, ok := args.Get(0).(pingone.PagedIterator[pingone.EnvironmentsCollectionResponse])
	if !ok {
		return nil, args.Error(1)
	}
	return response, args.Error(1)
}

func (p *mockPingOneClientEnvironmentsWrapper) CreateEnvironment(ctx context.Context, request *pingone.EnvironmentCreateRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := p.Called(ctx, request)
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

func (p *mockPingOneClientEnvironmentsWrapper) GetEnvironmentById(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := p.Called(ctx, environmentId)
	var response *pingone.EnvironmentResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*pingone.EnvironmentResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientEnvironmentsWrapper) UpdateEnvironmentById(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentReplaceRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	args := p.Called(ctx, environmentId, request)
	var response *pingone.EnvironmentResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*pingone.EnvironmentResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientEnvironmentsWrapper) GetEnvironmentServicesById(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := p.Called(ctx, environmentId)
	var response *pingone.EnvironmentBillOfMaterialsResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientEnvironmentsWrapper) UpdateEnvironmentServicesById(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentBillOfMaterialsReplaceRequest) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	args := p.Called(ctx, environmentId, request)
	var response *pingone.EnvironmentBillOfMaterialsResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*pingone.EnvironmentBillOfMaterialsResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return response, httpResponse, args.Error(2)
}
