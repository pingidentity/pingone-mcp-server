// Copyright © 2025 Ping Identity Corporation

// Package legacy provides client factory implementations for the legacy PingOne Go SDK v2.
// This package is used for backward compatibility with the patrickcping/pingone-go-sdk-v2
// SDK implementation and will be deprecated in future versions in favor of the
// pingidentity/pingone-go-client SDK.
package legacy

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/patrickcping/pingone-go-sdk-v2/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
)

var _ ClientFactory = &DefaultClientFactory{}

// DefaultClientFactory creates PingOne API clients using the legacy SDK (v2).
// It configures clients with proper authentication, region settings, and user agent
// information. The factory requires a valid access token and the PINGONE_ROOT_DOMAIN
// environment variable to determine the correct regional endpoints.
type DefaultClientFactory struct {
	// serverVersion is the version of the MCP server, included in the User-Agent header
	// for API request tracking and debugging purposes.
	serverVersion string
}

// NewDefaultClientFactory creates a new DefaultClientFactory instance.
// The serverVersion parameter is used to construct the User-Agent header for API requests,
// enabling tracking and debugging of API calls from specific server versions.
//
// The serverVersion parameter should be a valid semantic version string (e.g., "1.0.0").
// While empty values are accepted for testing purposes, production usage should always
// provide a meaningful version string for proper API request auditing.
//
// Example:
//
//	factory := NewDefaultClientFactory("1.2.3")
//	client, err := factory.NewClient(ctx, accessToken)
func NewDefaultClientFactory(serverVersion string) *DefaultClientFactory {
	return &DefaultClientFactory{
		serverVersion: serverVersion,
	}
}

// NewClient creates a new PingOne API client instance using the legacy SDK.
// It returns a fully configured client ready to make API calls to PingOne services.
//
// The ctx parameter provides the context for the client initialization and should include
// appropriate timeout and cancellation controls. This context is used during the initial
// client setup but does not control the lifetime of the returned client.
//
// The accessToken parameter must be a valid, non-empty OAuth2 access token with appropriate
// scopes for the intended API operations. The token should be obtained through a proper
// OAuth2 flow before calling this method. Empty or whitespace-only tokens will result in
// an error.
//
// External Dependencies:
// This method requires the PINGONE_ROOT_DOMAIN environment variable to be set to determine
// the correct regional API endpoints. Supported values are:
//   - "pingone.com" for North America (NA)
//   - "pingone.eu" for Europe (EU)
//   - "pingone.asia" for Asia Pacific (AP)
//   - "pingone.com.au" for Australia (AU)
//   - "pingone.ca" for Canada (CA)
//   - "pingone.sg" for Singapore (SG)
//
// Returns:
//   - A configured *pingone.Client instance ready for API operations
//   - An error if validation fails, the environment variable is missing/invalid, or client
//     initialization fails
//
// Example:
//
//	ctx := context.Background()
//	os.Setenv("PINGONE_ROOT_DOMAIN", "pingone.com")
//	factory := NewDefaultClientFactory("1.0.0")
//	client, err := factory.NewClient(ctx, "your-access-token")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f *DefaultClientFactory) NewClient(ctx context.Context, accessToken string) (*pingone.Client, error) {
	// Validate access token is not empty or whitespace-only
	// This prevents accidental use of uninitialized or invalid tokens
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("access token is empty or contains only whitespace, cannot initialize client")
	}

	// Retrieve and validate the root domain from environment
	rootDomain := os.Getenv("PINGONE_ROOT_DOMAIN")
	regionCode, err := f.regionCodeFromRootDomain(rootDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to determine region from root domain %q: %w", rootDomain, err)
	}

	// Build client configuration with region and authentication
	config := &pingone.Config{
		RegionCode: regionCode,
	}

	// Add user agent for request tracking and debugging
	userAgentSuffix := audit.PingOneAPIUserAgent(f.serverVersion)
	config.UserAgentSuffix = &userAgentSuffix

	// Configure access token (no need to validate again as we checked above)
	config.AccessToken = &accessToken

	// Initialize the API client
	apiClient, err := config.APIClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}

	return apiClient, nil
}

// regionCodeFromRootDomain converts a PingOne root domain string to a region code enum.
// It returns the appropriate EnumRegionCode for the provided domain, which is used to
// configure API endpoints for the correct regional PingOne instance.
//
// The rootDomain parameter should be a valid PingOne root domain without protocol or path.
// Leading/trailing whitespace is automatically trimmed, and the comparison is case-insensitive
// to handle common input variations.
//
// Supported domains and their corresponding regions:
//   - "pingone.com"    -> NA (North America)
//   - "pingone.eu"     -> EU (Europe)
//   - "pingone.asia"   -> AP (Asia Pacific)
//   - "pingone.com.au" -> AU (Australia)
//   - "pingone.ca"     -> CA (Canada)
//   - "pingone.sg"     -> SG (Singapore)
//
// Security Note:
// This method sanitizes input by trimming whitespace and converting to lowercase to prevent
// injection attacks and ensure consistent comparisons. Unrecognized domains are rejected
// to prevent misconfiguration and potential routing to incorrect API endpoints.
//
// Returns:
//   - A pointer to the EnumRegionCode for the specified domain
//   - An error if the domain is empty, contains only whitespace, or is not recognized
func (f *DefaultClientFactory) regionCodeFromRootDomain(rootDomain string) (*management.EnumRegionCode, error) {
	// Sanitize input: trim whitespace and convert to lowercase for consistent comparison
	// This prevents common input errors and potential security issues
	sanitizedDomain := strings.ToLower(strings.TrimSpace(rootDomain))

	// Validate domain is not empty after sanitization
	if sanitizedDomain == "" {
		return nil, fmt.Errorf("root PingOne domain is empty or contains only whitespace")
	}

	// Map domain to region code
	// Note: We use the sanitized domain for comparison to ensure case-insensitive matching
	switch sanitizedDomain {
	case "pingone.com":
		regionCode := management.ENUMREGIONCODE_NA
		return &regionCode, nil
	case "pingone.eu":
		regionCode := management.ENUMREGIONCODE_EU
		return &regionCode, nil
	case "pingone.asia":
		regionCode := management.ENUMREGIONCODE_AP
		return &regionCode, nil
	case "pingone.com.au":
		regionCode := management.ENUMREGIONCODE_AU
		return &regionCode, nil
	case "pingone.ca":
		regionCode := management.ENUMREGIONCODE_CA
		return &regionCode, nil
	case "pingone.sg":
		regionCode := management.ENUMREGIONCODE_SG
		return &regionCode, nil
	default:
		// Return descriptive error with list of supported domains to help users fix configuration
		return nil, fmt.Errorf("unrecognized root PingOne domain: %q. Supported domains are: pingone.com, pingone.eu, pingone.asia, pingone.com.au, pingone.ca, pingone.sg", sanitizedDomain)
	}
}
