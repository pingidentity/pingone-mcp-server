// Copyright © 2025 Ping Identity Corporation

package applications_test

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/applications"
	"github.com/stretchr/testify/mock"
)

var _ applications.ApplicationsClient = &mockPingOneClientApplicationsWrapper{}
var _ applications.ApplicationsClientFactory = &mockPingOneClientApplicationsWrapperFactory{}

type mockPingOneClientApplicationsWrapper struct {
	mock.Mock
}

type mockPingOneClientApplicationsWrapperFactory struct {
	mockClient applications.ApplicationsClient
	err        error
}

// Directly returns the provided mock client and error
func NewMockPingOneClientApplicationsWrapperFactory(mockClient applications.ApplicationsClient, err error) *mockPingOneClientApplicationsWrapperFactory {
	return &mockPingOneClientApplicationsWrapperFactory{
		mockClient: mockClient,
		err:        err,
	}
}

func (f *mockPingOneClientApplicationsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (applications.ApplicationsClient, error) {
	return f.mockClient, f.err
}

func (p *mockPingOneClientApplicationsWrapper) GetApplications(ctx context.Context, environmentId uuid.UUID) (management.EntityArrayPagedIterator, error) {
	args := p.Called(ctx, environmentId)
	var response management.EntityArrayPagedIterator
	response, ok := args.Get(0).(management.EntityArrayPagedIterator)
	if !ok {
		return nil, args.Error(1)
	}
	return response, args.Error(1)
}

func (p *mockPingOneClientApplicationsWrapper) CreateApplication(ctx context.Context, environmentId uuid.UUID, app management.CreateApplicationRequest) (*management.CreateApplication201Response, *http.Response, error) {
	args := p.Called(ctx, environmentId, app)
	var response *management.CreateApplication201Response
	response, ok := args.Get(0).(*management.CreateApplication201Response)
	if !ok {
		return nil, nil, args.Error(2)
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok {
		return response, nil, args.Error(2)
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientApplicationsWrapper) GetApplication(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID) (*management.ReadOneApplication200Response, *http.Response, error) {
	args := p.Called(ctx, environmentId, applicationId)
	var response *management.ReadOneApplication200Response
	response, ok := args.Get(0).(*management.ReadOneApplication200Response)
	if !ok {
		return nil, nil, args.Error(2)
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok {
		return response, nil, args.Error(2)
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientApplicationsWrapper) UpdateApplicationById(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID, app management.UpdateApplicationRequest) (*management.ReadOneApplication200Response, *http.Response, error) {
	args := p.Called(ctx, environmentId, applicationId, app)
	var response *management.ReadOneApplication200Response
	response, ok := args.Get(0).(*management.ReadOneApplication200Response)
	if !ok {
		return nil, nil, args.Error(2)
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok {
		return response, nil, args.Error(2)
	}
	return response, httpResponse, args.Error(2)
}
