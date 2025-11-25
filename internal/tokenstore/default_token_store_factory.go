// Copyright © 2025 Ping Identity Corporation

package tokenstore

import "fmt"

var _ TokenStoreFactory = &DefaultTokenStoreFactory{}

type DefaultTokenStoreFactory struct{}

func NewDefaultTokenStoreFactory() *DefaultTokenStoreFactory {
	return &DefaultTokenStoreFactory{}
}

func (d *DefaultTokenStoreFactory) NewTokenStore(storeType StoreType) (TokenStore, error) {
	switch storeType {
	case StoreTypeKeychain:
		return NewKeychainTokenStore()
	case StoreTypeFile:
		return NewFileTokenStore()
	default:
		return nil, fmt.Errorf("unsupported token store type when creating token store: %s", storeType.String())
	}
}
