// Copyright © 2025 Ping Identity Corporation

package tokenstore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/zalando/go-keyring"
)

const keychainServiceName = "pingone_mcp_server"
const keychainUsername = "auth_session"

var (
	_ TokenStore = &KeychainTokenStore{}
)

// KeychainTokenStore provides a keychain-based implementation of TokenStore
type KeychainTokenStore struct{}

func NewKeychainTokenStore() *KeychainTokenStore {
	return &KeychainTokenStore{}
}

func (k *KeychainTokenStore) PutSession(session auth.AuthSession) error {
	tokenJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal auth session: %w", err)
	}

	err = keyring.Set(keychainServiceName, keychainUsername, string(tokenJSON))
	if err != nil {
		return fmt.Errorf("failed to save auth session to keychain: %w", err)
	}
	return nil
}

func (k *KeychainTokenStore) HasSession() (bool, error) {
	_, err := k.GetSession()
	if err != nil {
		// If the error indicates that the item was not found, return false
		if errors.Is(err, keyring.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (k *KeychainTokenStore) GetSession() (*auth.AuthSession, error) {
	savedSession, err := keyring.Get(keychainServiceName, keychainUsername)
	if err != nil {
		return nil, err
	}
	var session auth.AuthSession
	if err := json.Unmarshal([]byte(savedSession), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth session from keychain: %w", err)
	}
	return &session, nil
}

func (k *KeychainTokenStore) DeleteSession() error {
	err := keyring.Delete(keychainServiceName, keychainUsername)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("failed to clear token from keychain: %w", err)
	}
	return nil
}
