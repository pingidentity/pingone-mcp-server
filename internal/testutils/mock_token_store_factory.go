// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/stretchr/testify/mock"
)

var _ tokenstore.TokenStoreFactory = &MockTokenStoreFactory{}

// MockTokenStoreFactory is a mock implementation of tokenstore.TokenStoreFactory using testify mock
type MockTokenStoreFactory struct {
	mock.Mock
}

func (m *MockTokenStoreFactory) NewTokenStore(storeType tokenstore.StoreType) (tokenstore.TokenStore, error) {
	args := m.Called(storeType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(tokenstore.TokenStore), args.Error(1)
}

func NewMockTokenStoreFactory() *MockTokenStoreFactory {
	return &MockTokenStoreFactory{}
}

// NewMockTokenStoreFactoryWithStore creates a MockTokenStoreFactory that returns the given store for any StoreType
func NewMockTokenStoreFactoryWithStore(store tokenstore.TokenStore) *MockTokenStoreFactory {
	factory := &MockTokenStoreFactory{}
	factory.On("NewTokenStore", mock.Anything).Return(store, nil)
	return factory
}

// NewMockTokenStoreFactoryWithError creates a MockTokenStoreFactory that returns an error for any StoreType
func NewMockTokenStoreFactoryWithError(err error) *MockTokenStoreFactory {
	factory := &MockTokenStoreFactory{}
	factory.On("NewTokenStore", mock.Anything).Return(nil, err)
	return factory
}
