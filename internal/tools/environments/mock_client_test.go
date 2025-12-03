// Copyright © 2025 Ping Identity Corporation

package environments_test

import (
	"context"

	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments/testutils"
)

var _ environments.EnvironmentsClient = &testutils.MockEnvironmentsClient{}
var _ environments.EnvironmentsClientFactory = &mockPingOneClientEnvironmentsWrapperFactory{}

type mockPingOneClientEnvironmentsWrapperFactory struct {
	mockClient environments.EnvironmentsClient
	err        error
}

// NewMockPingOneClientEnvironmentsWrapperFactory creates a factory that directly returns the provided mock client and error.
// The mockClient parameter is the mock client instance to return from GetAuthenticatedClient.
// The err parameter is the error to return from GetAuthenticatedClient, or nil for successful authentication.
func NewMockPingOneClientEnvironmentsWrapperFactory(mockClient environments.EnvironmentsClient, err error) *mockPingOneClientEnvironmentsWrapperFactory {
	return &mockPingOneClientEnvironmentsWrapperFactory{
		mockClient: mockClient,
		err:        err,
	}
}

// GetAuthenticatedClient returns the pre-configured mock client and error.
// This method implements the EnvironmentsClientFactory interface for testing purposes.
// The ctx parameter provides context for the authentication operation (not used in mock).
func (f *mockPingOneClientEnvironmentsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (environments.EnvironmentsClient, error) {
	return f.mockClient, f.err
}
