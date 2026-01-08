// Copyright © 2025 Ping Identity Corporation

package directory_test

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/directory"
	"github.com/stretchr/testify/mock"
)

var _ directory.DirectoryClient = &mockPingOneClientDirectoryWrapper{}
var _ directory.DirectoryClientFactory = &mockPingOneClientDirectoryWrapperFactory{}

type mockPingOneClientDirectoryWrapper struct {
	mock.Mock
}

type mockPingOneClientDirectoryWrapperFactory struct {
	mockClient directory.DirectoryClient
	err        error
}

// NewMockPingOneClientDirectoryWrapperFactory directly returns the provided mock client and error
func NewMockPingOneClientDirectoryWrapperFactory(mockClient directory.DirectoryClient, err error) *mockPingOneClientDirectoryWrapperFactory {
	return &mockPingOneClientDirectoryWrapperFactory{
		mockClient: mockClient,
		err:        err,
	}
}

func (f *mockPingOneClientDirectoryWrapperFactory) GetAuthenticatedClient(ctx context.Context) (directory.DirectoryClient, error) {
	return f.mockClient, f.err
}

func (p *mockPingOneClientDirectoryWrapper) GetTotalIdentitiesByEnvironmentId(ctx context.Context, environmentId uuid.UUID, filter string) (*pingone.DirectoryTotalIdentitiesCountCollectionResponse, *http.Response, error) {
	args := p.Called(ctx, environmentId, filter)
	var response *pingone.DirectoryTotalIdentitiesCountCollectionResponse
	if args.Get(0) != nil {
		response = args.Get(0).(*pingone.DirectoryTotalIdentitiesCountCollectionResponse)
	}
	var httpResponse *http.Response
	if args.Get(1) != nil {
		httpResponse = args.Get(1).(*http.Response)
	}
	return response, httpResponse, args.Error(2)
}
