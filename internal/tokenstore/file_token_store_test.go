// Copyright © 2025 Ping Identity Corporation

package tokenstore_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempTokenStore(t *testing.T) *tokenstore.FileTokenStore {
	t.Helper()

	tempDir := t.TempDir()
	store, err := tokenstore.NewFileTokenStoreWithBasePath(tempDir)
	require.NoError(t, err, "Failed to create temp FileTokenStore")
	return store
}

func createTestAuthSession() auth.AuthSession {
	return auth.AuthSession{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
		SessionId:    "test-session-id",
	}
}

func TestFileTokenStore_PutSession(t *testing.T) {
	store := createTempTokenStore(t)
	// Save first session
	session1 := createTestAuthSession()
	err := store.PutSession(session1)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(store.GetFilePath())
	assert.NoError(t, err, "Session file should exist")

	// Verify file contents
	data, err := os.ReadFile(store.GetFilePath())
	require.NoError(t, err, "Should be able to read session file")

	var savedSession auth.AuthSession
	err = json.Unmarshal(data, &savedSession)
	require.NoError(t, err, "Should be able to unmarshal session")

	assert.Equal(t, session1.AccessToken, savedSession.AccessToken)
	assert.Equal(t, session1.RefreshToken, savedSession.RefreshToken)
	assert.Equal(t, session1.SessionId, savedSession.SessionId)
	assert.True(t, session1.Expiry.Equal(savedSession.Expiry))

	// Save second session with different values
	session2 := auth.AuthSession{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		Expiry:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		SessionId:    "new-session-id",
	}
	err = store.PutSession(session2)
	require.NoError(t, err)

	// Verify the second session is saved
	retrieved, err := store.GetSession()
	require.NoError(t, err)
	assert.Equal(t, session2.AccessToken, retrieved.AccessToken)
	assert.Equal(t, session2.SessionId, retrieved.SessionId)
}

func TestFileTokenStore_HasSession(t *testing.T) {
	store := createTempTokenStore(t)

	exists, err := store.HasSession()
	require.NoError(t, err, "HasSession should not return an error")
	assert.False(t, exists, "HasSession should return false when session file does not exist")

	session := createTestAuthSession()
	err = store.PutSession(session)
	require.NoError(t, err)

	exists, err = store.HasSession()
	require.NoError(t, err, "HasSession should not return an error")
	assert.True(t, exists, "HasSession should return true when session exists")

	// Write invalid JSON to file
	err = os.WriteFile(store.GetFilePath(), []byte("invalid json"), 0600)
	require.NoError(t, err)

	_, err = store.HasSession()
	assert.Error(t, err, "HasSession should return an error for invalid JSON")
}

func TestFileTokenStore_GetSession(t *testing.T) {
	store := createTempTokenStore(t)

	retrieved, err := store.GetSession()
	assert.Error(t, err, "GetSession should return an error when file does not exist")
	assert.Nil(t, retrieved, "Retrieved session should be nil")
	assert.Contains(t, err.Error(), "auth session file not found")

	session := createTestAuthSession()
	err = store.PutSession(session)
	require.NoError(t, err)

	retrieved, err = store.GetSession()
	require.NoError(t, err, "GetSession should not return an error")
	require.NotNil(t, retrieved, "Retrieved session should not be nil")

	assert.Equal(t, session.AccessToken, retrieved.AccessToken)
	assert.Equal(t, session.RefreshToken, retrieved.RefreshToken)
	assert.Equal(t, session.SessionId, retrieved.SessionId)
	assert.True(t, session.Expiry.Equal(retrieved.Expiry))

	// Write invalid JSON to file
	err = os.WriteFile(store.GetFilePath(), []byte("invalid json"), 0600)
	require.NoError(t, err)

	retrieved, err = store.GetSession()
	assert.Error(t, err, "GetSession should return an error for invalid JSON")
	assert.Nil(t, retrieved, "Retrieved session should be nil")
}

func TestFileTokenStore_DeleteSession(t *testing.T) {
	store := createTempTokenStore(t)
	// Attempt to delete non-existent file
	err := store.DeleteSession()
	assert.NoError(t, err, "DeleteSession should not return an error for non-existent file")

	// Create a session
	session := createTestAuthSession()
	err = store.PutSession(session)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(store.GetFilePath())
	require.NoError(t, err, "Session file should exist before deletion")

	// Delete session
	err = store.DeleteSession()
	require.NoError(t, err, "DeleteSession should not return an error")

	// Verify file no longer exists
	_, err = os.Stat(store.GetFilePath())
	assert.True(t, os.IsNotExist(err), "Session file should not exist after deletion")
	exists, err := store.HasSession()
	require.NoError(t, err, "HasSession should not return an error")
	assert.False(t, exists, "HasSession should return false when session file does not exist")
}

func TestFileTokenStore_NewFileTokenStore(t *testing.T) {
	store, err := tokenstore.NewFileTokenStore()
	require.NoError(t, err, "NewFileTokenStore should not return an error")

	// Verify the path is in the home directory
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	relPath, err := filepath.Rel(homeDir, store.GetFilePath())
	require.NoError(t, err)

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		t.Errorf("File path %s is not in home directory %s. Relative path: %s", store.GetFilePath(), homeDir, relPath)
	}
}
