// Copyright © 2026 Ping Identity Corporation

package tokenstore

type TokenStoreFactory interface {
	NewTokenStore(storeType StoreType) (TokenStore, error)
}
