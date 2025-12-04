// Copyright © 2025 Ping Identity Corporation

package legacy

import (
	"context"
	"os"
	"testing"

	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClientFactory_NewClient_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		accessToken string
		wantErrMsg  string
	}{
		{
			name:        "Empty access token",
			accessToken: "",
			wantErrMsg:  "access token is empty or contains only whitespace",
		},
		{
			name:        "Whitespace only token",
			accessToken: "   ",
			wantErrMsg:  "access token is empty or contains only whitespace",
		},
		{
			name:        "Tab and space token",
			accessToken: "\t  \n",
			wantErrMsg:  "access token is empty or contains only whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewDefaultClientFactory("1.0.0")
			ctx := context.Background()

			client, err := factory.NewClient(ctx, tt.accessToken)

			assert.Nil(t, client)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestDefaultClientFactory_NewClient_MissingEnvironmentVariable(t *testing.T) {
	// Save original value and clear it
	originalDomain := os.Getenv("PINGONE_ROOT_DOMAIN")
	err := os.Unsetenv("PINGONE_ROOT_DOMAIN")
	require.NoError(t, err)
	defer func() {
		if originalDomain != "" {
			err := os.Setenv("PINGONE_ROOT_DOMAIN", originalDomain)
			require.NoError(t, err)
		}
	}()

	factory := NewDefaultClientFactory("1.0.0")
	ctx := context.Background()

	client, err := factory.NewClient(ctx, "valid-token")

	assert.Nil(t, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to determine region from root domain")
	assert.Contains(t, err.Error(), "empty or contains only whitespace")
}

func TestDefaultClientFactory_regionCodeFromRootDomain(t *testing.T) {
	tests := []struct {
		name           string
		rootDomain     string
		expectedRegion management.EnumRegionCode
		expectError    bool
	}{
		// Valid domain tests
		{
			name:           "Valid NA domain",
			rootDomain:     "pingone.com",
			expectedRegion: management.ENUMREGIONCODE_NA,
			expectError:    false,
		},
		{
			name:           "Valid EU domain",
			rootDomain:     "pingone.eu",
			expectedRegion: management.ENUMREGIONCODE_EU,
			expectError:    false,
		},
		{
			name:           "Valid AP domain",
			rootDomain:     "pingone.asia",
			expectedRegion: management.ENUMREGIONCODE_AP,
			expectError:    false,
		},
		{
			name:           "Valid AU domain",
			rootDomain:     "pingone.com.au",
			expectedRegion: management.ENUMREGIONCODE_AU,
			expectError:    false,
		},
		{
			name:           "Valid CA domain",
			rootDomain:     "pingone.ca",
			expectedRegion: management.ENUMREGIONCODE_CA,
			expectError:    false,
		},
		{
			name:           "Valid SG domain",
			rootDomain:     "pingone.sg",
			expectedRegion: management.ENUMREGIONCODE_SG,
			expectError:    false,
		},
		// Case sensitivity tests
		{
			name:           "Uppercase domain",
			rootDomain:     "PINGONE.COM",
			expectedRegion: management.ENUMREGIONCODE_NA,
			expectError:    false,
		},
		{
			name:           "Mixed case domain",
			rootDomain:     "PingOne.EU",
			expectedRegion: management.ENUMREGIONCODE_EU,
			expectError:    false,
		},
		// Whitespace handling tests
		{
			name:           "Domain with leading whitespace",
			rootDomain:     "  pingone.com",
			expectedRegion: management.ENUMREGIONCODE_NA,
			expectError:    false,
		},
		{
			name:           "Domain with trailing whitespace",
			rootDomain:     "pingone.eu  ",
			expectedRegion: management.ENUMREGIONCODE_EU,
			expectError:    false,
		},
		{
			name:           "Domain with surrounding whitespace",
			rootDomain:     "  pingone.asia  ",
			expectedRegion: management.ENUMREGIONCODE_AP,
			expectError:    false,
		},
		// Invalid input tests
		{
			name:        "Empty domain",
			rootDomain:  "",
			expectError: true,
		},
		{
			name:        "Whitespace only domain",
			rootDomain:  "   ",
			expectError: true,
		},
		{
			name:        "Invalid domain",
			rootDomain:  "invalid.com",
			expectError: true,
		},
		{
			name:        "Partial domain",
			rootDomain:  "pingone",
			expectError: true,
		},
		{
			name:        "Domain with protocol",
			rootDomain:  "https://pingone.com",
			expectError: true,
		},
		{
			name:        "Invalid PingIdentity domain",
			rootDomain:  "www.pingidentity.com",
			expectError: true,
		},
		// Security validation tests - potential injection attempts
		{
			name:        "Command injection with semicolon",
			rootDomain:  "pingone.com; rm -rf /",
			expectError: true,
		},
		{
			name:        "Command injection with ampersand",
			rootDomain:  "pingone.com && malicious",
			expectError: true,
		},
		{
			name:        "Command injection with pipe",
			rootDomain:  "pingone.com | cat /etc/passwd",
			expectError: true,
		},
		{
			name:        "Newline injection",
			rootDomain:  "pingone.com\nmalicious",
			expectError: true,
		},
		{
			name:        "Carriage return injection",
			rootDomain:  "pingone.com\rmalicious",
			expectError: true,
		},
		{
			name:        "Path traversal attempt 1",
			rootDomain:  "../../etc/passwd",
			expectError: true,
		},
		{
			name:        "Path traversal attempt 2",
			rootDomain:  "../pingone.com",
			expectError: true,
		},
		{
			name:        "Path traversal attempt 3",
			rootDomain:  "pingone.com/../../etc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewDefaultClientFactory("1.0.0")

			regionCode, err := factory.regionCodeFromRootDomain(tt.rootDomain)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, regionCode)
				// Verify error message contains helpful information
				if tt.rootDomain == "" || len(tt.rootDomain) == 0 || len(tt.rootDomain) > 0 && tt.rootDomain[0] == ' ' {
					assert.Contains(t, err.Error(), "empty or contains only whitespace")
				} else {
					assert.Contains(t, err.Error(), "unrecognized root PingOne domain")
					assert.Contains(t, err.Error(), "Supported domains are:")
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, regionCode)
				assert.Equal(t, tt.expectedRegion, *regionCode)
			}
		})
	}
}
