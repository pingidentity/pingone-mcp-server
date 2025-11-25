// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

var _ client.AuthClient = &MockAuthClient{}
var _ client.AuthClientFactory = &MockAuthClientFactory{}

type MockAuthClient struct {
	mock.Mock
}

func NewMockAuthClient(tokenSource oauth2.TokenSource) *MockAuthClient {
	result := &MockAuthClient{}
	result.On("TokenSource", mock.Anything, mock.Anything).Return(tokenSource, nil)
	return result
}

func (m *MockAuthClient) TokenSource(ctx context.Context, grantType auth.GrantType) (oauth2.TokenSource, error) {
	args := m.Called(ctx, grantType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(oauth2.TokenSource), args.Error(1)
}

type MockAuthClientFactory struct {
	mock.Mock
}

func NewMockAuthClientFactory(tokenSource oauth2.TokenSource) *MockAuthClientFactory {
	result := &MockAuthClientFactory{}
	result.On("NewAuthClient").Return(NewMockAuthClient(tokenSource), nil)
	return result
}

func (f *MockAuthClientFactory) NewAuthClient() (client.AuthClient, error) {
	args := f.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.AuthClient), args.Error(1)
}

func NewEmptyMockAuthClientFactory() *MockAuthClientFactory {
	return &MockAuthClientFactory{}
}
