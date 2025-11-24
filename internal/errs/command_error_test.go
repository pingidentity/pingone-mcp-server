// Copyright © 2025 Ping Identity Corporation

package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/errs"
)

func TestCommandError_Error(t *testing.T) {
	tests := []struct {
		name          string
		commandName   string
		originalError error
		expected      string
	}{
		{
			name:     "no error and no command name",
			expected: "unknown command error",
		},
		{
			name:        "no error with command name",
			commandName: "run",
			expected:    "pingone-mcp-server run command failed: unknown error",
		},
		{
			name:          "original error only (empty command name)",
			commandName:   "",
			originalError: errors.New("connection failed"),
			expected:      "connection failed",
		},
		{
			name:          "original error with command name",
			commandName:   "init",
			originalError: errors.New("authentication failed"),
			expected:      "pingone-mcp-server init command failed: authentication failed",
		},
		{
			name:        "nil error with command name",
			commandName: "test-command",
			expected:    "pingone-mcp-server test-command command failed: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := &errs.CommandError{
				CommandName:   tt.commandName,
				OriginalError: tt.originalError,
			}

			result := cmdErr.Error()
			if result != tt.expected {
				t.Errorf("Expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}

func TestNewCommandError(t *testing.T) {
	tests := []struct {
		name         string
		commandName  string
		originalErr  error
		expectedType string
		checkFunc    func(t *testing.T, err error)
	}{
		{
			name:         "nil error",
			commandName:  "run",
			originalErr:  nil,
			expectedType: "*errs.CommandError",
			checkFunc: func(t *testing.T, err error) {
				cmdErr := err.(*errs.CommandError)
				if cmdErr.CommandName != "run" {
					t.Errorf("Expected command name 'run', got: %s", cmdErr.CommandName)
				}
				if cmdErr.OriginalError != nil {
					t.Errorf("Expected nil original error, got: %v", cmdErr.OriginalError)
				}
			},
		},
		{
			name:         "with original error",
			commandName:  "init",
			originalErr:  errors.New("client factory initialization failed"),
			expectedType: "*errs.CommandError",
			checkFunc: func(t *testing.T, err error) {
				cmdErr := err.(*errs.CommandError)
				if cmdErr.CommandName != "init" {
					t.Errorf("Expected command name 'init', got: %s", cmdErr.CommandName)
				}
				if cmdErr.OriginalError == nil || cmdErr.OriginalError.Error() != "client factory initialization failed" {
					t.Errorf("Expected original error 'client factory initialization failed', got: %v", cmdErr.OriginalError)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.NewCommandError(tt.commandName, tt.originalErr)

			if err == nil {
				t.Fatal("Expected non-nil error")
			}

			// Check type
			if got := fmt.Sprintf("%T", err); got != tt.expectedType {
				t.Errorf("Expected type %s, got: %s", tt.expectedType, got)
			}

			// Run custom checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}
