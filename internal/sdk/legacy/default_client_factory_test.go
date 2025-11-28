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

func TestNewDefaultClientFactory(t *testing.T) {
	tests := []struct {
		name          string
		serverVersion string
	}{
		{
			name:          "Valid version string",
			serverVersion: "1.0.0",
		},
		{
			name:          "Empty version string",
			serverVersion: "",
		},
		{
			name:          "Complex version string",
			serverVersion: "1.2.3-beta.1+build.456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewDefaultClientFactory(tt.serverVersion)
			require.NotNil(t, factory)
			assert.Equal(t, tt.serverVersion, factory.serverVersion)
		})
	}
}

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

func TestDefaultClientFactory_regionCodeFromRootDomain_SecurityValidation(t *testing.T) {
	// Test that potential injection attempts are handled safely
	maliciousInputs := []string{
		"pingone.com; rm -rf /",
		"pingone.com && malicious",
		"pingone.com | cat /etc/passwd",
		"pingone.com\nmalicious",
		"pingone.com\rmalicious",
		"../../etc/passwd",
		"../pingone.com",
		"pingone.com/../../etc",
	}

	factory := NewDefaultClientFactory("1.0.0")

	for _, input := range maliciousInputs {
		t.Run("Malicious input: "+input, func(t *testing.T) {
			regionCode, err := factory.regionCodeFromRootDomain(input)

			// All malicious inputs should be rejected
			assert.Error(t, err)
			assert.Nil(t, regionCode)
			assert.Contains(t, err.Error(), "unrecognized root PingOne domain")
		})
	}
}

func TestDefaultClientFactory_NewClient_InvalidDomain(t *testing.T) {
	// Save and set invalid domain
	originalDomain := os.Getenv("PINGONE_ROOT_DOMAIN")
	err := os.Setenv("PINGONE_ROOT_DOMAIN", "www.pingidentity.com")
	require.NoError(t, err)
	defer func() {
		if originalDomain != "" {
			err := os.Setenv("PINGONE_ROOT_DOMAIN", originalDomain)
			require.NoError(t, err)
		} else {
			err := os.Unsetenv("PINGONE_ROOT_DOMAIN")
			require.NoError(t, err)
		}
	}()

	factory := NewDefaultClientFactory("1.0.0")
	ctx := context.Background()

	client, err := factory.NewClient(ctx, "valid-token")

	assert.Nil(t, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to determine region from root domain")
	assert.Contains(t, err.Error(), "unrecognized root PingOne domain")
	assert.Contains(t, err.Error(), "www.pingidentity.com")
}
