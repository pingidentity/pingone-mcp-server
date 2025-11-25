// Copyright © 2025 Ping Identity Corporation

package tokenstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
)

const defaultTokenFileName = ".pingone_mcp_session.json"

var (
	_ TokenStore = &FileTokenStore{}
)

type FileTokenStore struct {
	filePath string
}

// NewFileTokenStore creates a new FileTokenStore with the default file path in the user's home directory
func NewFileTokenStore() (*FileTokenStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory when creating file token store: %w", err)
	}

	return NewFileTokenStoreWithBasePath(homeDir)
}

func NewFileTokenStoreWithBasePath(basePath string) (*FileTokenStore, error) {
	filePath := filepath.Join(basePath, defaultTokenFileName)
	return &FileTokenStore{
		filePath: filePath,
	}, nil
}

func (f *FileTokenStore) PutSession(session auth.AuthSession) error {
	tokenJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal auth session: %w", err)
	}

	// Ensure the directory exists, and create it if necessary
	dir := filepath.Dir(f.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory for auth session file: %w", err)
	}

	err = os.WriteFile(f.filePath, tokenJSON, 0600)
	if err != nil {
		return fmt.Errorf("failed to save auth session to file: %w", err)
	}
	return nil
}

func (f *FileTokenStore) HasSession() (bool, error) {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No session on filesystem
			return false, nil
		}
		return false, fmt.Errorf("failed to read auth session from file: %w", err)
	}

	var session auth.AuthSession
	if err := json.Unmarshal(data, &session); err != nil {
		return false, fmt.Errorf("failed to unmarshal auth session from file: %w", err)
	}
	return true, nil
}

func (f *FileTokenStore) GetSession() (*auth.AuthSession, error) {
	data, err := os.ReadFile(f.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("auth session file not found")
		}
		return nil, fmt.Errorf("failed to read auth session from file: %w", err)
	}

	var session auth.AuthSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth session from file: %w", err)
	}
	return &session, nil
}

func (f *FileTokenStore) DeleteSession() error {
	err := os.Remove(f.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete auth session file: %w", err)
	}
	return nil
}

func (f *FileTokenStore) GetFilePath() string {
	return f.filePath
}
