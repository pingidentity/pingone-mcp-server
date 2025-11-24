// Copyright © 2025 Ping Identity Corporation

package cmd_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		errorContains string
		description   string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: false,
			description: "Root command should execute without error",
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
			description: "Root command help should execute without error",
		},
		{
			name:          "invalid flag",
			args:          []string{"--invalid-flag"},
			expectError:   true,
			errorContains: "unknown flag",
			description:   "Root command should return error for invalid flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := testutils.ExecuteCliRootCommand(t, ctx, tt.args...)

			if tt.expectError {
				require.Error(t, err, tt.description)
				if tt.errorContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errorContains),
						"Error should contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				require.NoError(t, err, tt.description)
			}
		})
	}
}
