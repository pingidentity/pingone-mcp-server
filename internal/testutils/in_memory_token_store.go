// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"errors"
	"sync"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

var (
	ErrSessionNotFound                       = errors.New("session not found")
	_                  tokenstore.TokenStore = &InMemoryTokenStore{}
)

type InMemoryTokenStore struct {
	mu      sync.RWMutex
	session *auth.AuthSession
	// Errors to simulate failures for testing
	PutSessionError    error
	GetSessionError    error
	HasSessionError    error
	DeleteSessionError error
}

func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{}
}

func NewInMemoryTokenStoreWithDefaultSession() *InMemoryTokenStore {
	tokenStore := &InMemoryTokenStore{}
	tokenStore.session = &auth.AuthSession{
		SessionId:    "default-session-id",
		AccessToken:  "default-access-token",
		RefreshToken: "default-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	return tokenStore
}

func (s *InMemoryTokenStore) PutSession(session auth.AuthSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.PutSessionError != nil {
		return s.PutSessionError
	}
	s.session = &session
	return nil
}

func (s *InMemoryTokenStore) HasSession() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.HasSessionError != nil {
		return false, s.HasSessionError
	}
	return s.session != nil, nil
}

func (s *InMemoryTokenStore) GetSession() (*auth.AuthSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.GetSessionError != nil {
		return nil, s.GetSessionError
	}
	if s.session == nil {
		return nil, ErrSessionNotFound
	}
	return s.session, nil
}

func (s *InMemoryTokenStore) DeleteSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.DeleteSessionError != nil {
		return s.DeleteSessionError
	}
	s.session = nil
	return nil
}
