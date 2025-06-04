package main

import (
	"os"
	"testing"
)

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "env var is empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)
			
			// Set env var if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvString(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "env var true",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "env var false",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "env var invalid",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "invalid",
			expected:     true, // should return default
		},
		{
			name:         "env var not set",
			key:          "NONEXISTENT_BOOL",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "env var 1",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "env var 0",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)
			
			// Set env var if provided
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSetEnvFromFlag(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    interface{}
		existing string
		expected string
	}{
		{
			name:     "set string when no existing env var",
			key:      "TEST_FLAG_STR",
			value:    "flag-value",
			existing: "",
			expected: "flag-value",
		},
		{
			name:     "don't override existing env var with string",
			key:      "TEST_FLAG_STR2",
			value:    "flag-value",
			existing: "env-value",
			expected: "env-value",
		},
		{
			name:     "set bool true when no existing env var",
			key:      "TEST_FLAG_BOOL",
			value:    true,
			existing: "",
			expected: "true",
		},
		{
			name:     "set bool false when no existing env var",
			key:      "TEST_FLAG_BOOL2",
			value:    false,
			existing: "",
			expected: "false",
		},
		{
			name:     "don't override existing env var with bool",
			key:      "TEST_FLAG_BOOL3",
			value:    true,
			existing: "false",
			expected: "false",
		},
		{
			name:     "skip empty string",
			key:      "TEST_FLAG_EMPTY",
			value:    "",
			existing: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			defer os.Unsetenv(tt.key)
			
			// Set existing env var if provided
			if tt.existing != "" {
				os.Setenv(tt.key, tt.existing)
			}

			setEnvFromFlag(tt.key, tt.value)
			
			result := os.Getenv(tt.key)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "/tmp/test-api-key"
	defer os.Remove(tmpFile)

	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		setupFile   func()
		expectFile  bool
		expectKey   bool
	}{
		{
			name: "read existing file",
			setupEnv: func() {
				os.Setenv("PINGONE_MCP_API_KEY_PATH", tmpFile)
			},
			cleanupEnv: func() {
				os.Unsetenv("PINGONE_MCP_API_KEY_PATH")
			},
			setupFile: func() {
				os.WriteFile(tmpFile, []byte("existing-key\n"), 0600)
			},
			expectFile: true,
			expectKey:  true,
		},
		{
			name: "use env var and create file",
			setupEnv: func() {
				os.Setenv("PINGONE_MCP_API_KEY_PATH", tmpFile)
				os.Setenv("PINGONE_MCP_API_KEY", "env-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("PINGONE_MCP_API_KEY_PATH")
				os.Unsetenv("PINGONE_MCP_API_KEY")
			},
			setupFile: func() {
				// No file setup - should create one
			},
			expectFile: true,
			expectKey:  true,
		},
		{
			name: "generate new key",
			setupEnv: func() {
				os.Setenv("PINGONE_MCP_API_KEY_PATH", tmpFile)
			},
			cleanupEnv: func() {
				os.Unsetenv("PINGONE_MCP_API_KEY_PATH")
			},
			setupFile: func() {
				// No file setup - should generate and create one
			},
			expectFile: true,
			expectKey:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer tt.cleanupEnv()
			tt.setupFile()
			defer os.Remove(tmpFile)

			// Test
			key := getAPIKey()

			// Verify key exists
			if tt.expectKey && key == "" {
				t.Error("expected non-empty API key")
			}

			// Verify file was created
			if tt.expectFile {
				if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
					t.Error("expected API key file to be created")
				}
			}

			// Verify key is hex encoded (should be 64 chars for 32 bytes)
			if tt.expectKey && tt.name == "generate new key" {
				if len(key) != 64 {
					t.Errorf("expected 64-char hex key, got %d chars", len(key))
				}
			}
		})
	}
}

// Test helper functions that don't require complex setup
func TestUtilityFunctions(t *testing.T) {
	t.Run("getEnvString with various inputs", func(t *testing.T) {
		// Test with special characters
		os.Setenv("SPECIAL_CHARS", "!@#$%^&*()")
		defer os.Unsetenv("SPECIAL_CHARS")
		
		result := getEnvString("SPECIAL_CHARS", "default")
		if result != "!@#$%^&*()" {
			t.Errorf("expected special chars to be preserved, got %q", result)
		}
	})

	t.Run("getEnvBool edge cases", func(t *testing.T) {
		edgeCases := map[string]bool{
			"TRUE":  true,
			"True":  true,
			"FALSE": false,
			"False": false,
			"YES":   false, // strconv.ParseBool doesn't recognize YES
			"NO":    false,
		}

		for envVal, expected := range edgeCases {
			t.Run("value_"+envVal, func(t *testing.T) {
				key := "TEST_EDGE_CASE"
				os.Setenv(key, envVal)
				defer os.Unsetenv(key)

				result := getEnvBool(key, false)
				if result != expected {
					t.Errorf("for %q expected %v, got %v", envVal, expected, result)
				}
			})
		}
	})

	t.Run("setEnvFromFlag with different types", func(t *testing.T) {
		// Test with int (should be ignored)
		setEnvFromFlag("TEST_INT", 42)
		if os.Getenv("TEST_INT") != "" {
			t.Error("expected int value to be ignored")
		}
		defer os.Unsetenv("TEST_INT")

		// Test with nil (should be ignored)
		setEnvFromFlag("TEST_NIL", nil)
		if os.Getenv("TEST_NIL") != "" {
			t.Error("expected nil value to be ignored")
		}
		defer os.Unsetenv("TEST_NIL")
	})
}