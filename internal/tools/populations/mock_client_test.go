// Copyright © 2025 Ping Identity Corporation

package populations_test

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/populations"
	"github.com/stretchr/testify/mock"
)

var _ populations.PopulationsClient = &mockPingOneClientPopulationsWrapper{}
var _ populations.PopulationsClientFactory = &mockPingOneClientPopulationsWrapperFactory{}

type mockPingOneClientPopulationsWrapper struct {
	mock.Mock
}

type mockPingOneClientPopulationsWrapperFactory struct {
	mockClient populations.PopulationsClient
	err        error
}

// Directly returns the provided mock client and error
func NewMockPingOneClientPopulationsWrapperFactory(mockClient populations.PopulationsClient, err error) *mockPingOneClientPopulationsWrapperFactory {
	return &mockPingOneClientPopulationsWrapperFactory{
		mockClient: mockClient,
		err:        err,
	}
}

func (f *mockPingOneClientPopulationsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (populations.PopulationsClient, error) {
	return f.mockClient, f.err
}

func (p *mockPingOneClientPopulationsWrapper) GetPopulations(ctx context.Context, environmentId uuid.UUID, filter *string) (management.EntityArrayPagedIterator, error) {
	args := p.Called(ctx, environmentId, filter)
	var response management.EntityArrayPagedIterator
	response, ok := args.Get(0).(management.EntityArrayPagedIterator)
	if !ok {
		return nil, args.Error(1)
	}
	return response, args.Error(1)
}

func (p *mockPingOneClientPopulationsWrapper) CreatePopulation(ctx context.Context, environmentId uuid.UUID, createRequest management.Population) (*management.Population, *http.Response, error) {
	args := p.Called(ctx, environmentId, createRequest)
	var response *management.Population
	response, ok := args.Get(0).(*management.Population)
	if !ok && args.Get(0) != nil {
		panic("CreatePopulation mock setup error: expected *management.Population or nil")
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok && args.Get(1) != nil {
		panic("CreatePopulation mock setup error: expected *http.Response or nil")
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientPopulationsWrapper) GetPopulation(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID) (*management.Population, *http.Response, error) {
	args := p.Called(ctx, environmentId, populationId)
	var response *management.Population
	response, ok := args.Get(0).(*management.Population)
	if !ok && args.Get(0) != nil {
		panic("GetPopulation mock setup error: expected *management.Population or nil")
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok && args.Get(1) != nil {
		panic("GetPopulation mock setup error: expected *http.Response or nil")
	}
	return response, httpResponse, args.Error(2)
}

func (p *mockPingOneClientPopulationsWrapper) UpdatePopulation(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID, updateRequest management.Population) (*management.Population, *http.Response, error) {
	args := p.Called(ctx, environmentId, populationId, updateRequest)
	var response *management.Population
	response, ok := args.Get(0).(*management.Population)
	if !ok && args.Get(0) != nil {
		panic("UpdatePopulation mock setup error: expected *management.Population or nil")
	}
	var httpResponse *http.Response
	httpResponse, ok = args.Get(1).(*http.Response)
	if !ok && args.Get(1) != nil {
		panic("UpdatePopulation mock setup error: expected *http.Response or nil")
	}
	return response, httpResponse, args.Error(2)
}
