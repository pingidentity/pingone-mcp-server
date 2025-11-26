// Copyright © 2025 Ping Identity Corporation

package client

import (
	"context"
	"fmt"

	"github.com/pingidentity/pingone-go-client/config"
	pingoneOauth2 "github.com/pingidentity/pingone-go-client/oauth2"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-go-client/utils/browser"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"golang.org/x/oauth2"
)

var _ AuthClient = &PingOneClientAuthWrapper{}
var _ AuthClientFactory = &PingOneClientAuthWrapperFactory{}

type PingOneClientAuthWrapper struct {
	serverVersion string
	environmentId string
}

func NewPingOneClientAuthWrapper(serverVersion, environmentId string) *PingOneClientAuthWrapper {
	return &PingOneClientAuthWrapper{
		serverVersion: serverVersion,
		environmentId: environmentId,
	}
}

func (p *PingOneClientAuthWrapper) TokenSource(ctx context.Context, grantType auth.GrantType) (oauth2.TokenSource, error) {
	logger.FromContext(ctx).Debug("Creating token source from PingOne go client")

	var clientGrantType pingoneOauth2.GrantType
	switch grantType {
	case auth.GrantTypeAuthorizationCode:
		clientGrantType = pingoneOauth2.GrantTypeAuthorizationCode
	case auth.GrantTypeDeviceCode:
		clientGrantType = pingoneOauth2.GrantTypeDeviceCode
	default:
		return nil, fmt.Errorf("unsupported grant type for PingOne client auth wrapper: %s", grantType.String())
	}

	// Rely on environment variables to complete the configuration
	clientConfig := config.NewConfiguration().
		WithEnvironmentID(p.environmentId).
		WithGrantType(clientGrantType).
		WithStorageType(config.StorageTypeNone) // keychain storage will be managed by the mcp server

	// Configure custom UX handlers for headless operation
	p.configureHeadlessHandlers(ctx, clientConfig, grantType)

	pingoneConfig := pingone.NewConfiguration(clientConfig)
	pingoneConfig.AppendUserAgent(audit.PingOneAPIUserAgent(p.serverVersion))

	return pingoneConfig.Service.TokenSource(ctx)
}

// configureHeadlessHandlers sets up custom UX handlers for headless MCP server operation.
// This provides environment-aware browser handling:
// - If browser is available: opens browser for both auth code and device code flows
// - If no browser: auth code fails (requires browser), device code prints instructions
func (p *PingOneClientAuthWrapper) configureHeadlessHandlers(ctx context.Context, cfg *config.Configuration, grantType auth.GrantType) {
	log := logger.FromContext(ctx)

	// Check if we're in an environment with browser support
	canOpenBrowser := browser.CanOpen()

	switch grantType {
	case auth.GrantTypeDeviceCode:
		// Initialize DeviceCode struct if it doesn't exist
		if cfg.Auth.DeviceCode == nil {
			cfg.Auth.DeviceCode = &config.DeviceCode{}
		}

		// Set custom device code prompt handler
		cfg.Auth.DeviceCode.OnDisplayPrompt = func(verificationURI, userCode string) error {
			// For device code, we have a VerificationURIComplete (full URL with code embedded)
			// Construct it by appending the user code as a query parameter
			fullURL := fmt.Sprintf("%s?user_code=%s", verificationURI, userCode)

			// Always log the instructions
			log.Info("Device authorization required",
				"verification_uri", verificationURI,
				"user_code", userCode)
			fmt.Printf("\n=== PingOne MCP Server OAuth 2.0 Authorization ===\n")

			// Try to open browser if available
			browserOpened := false
			if canOpenBrowser {
				if err := browser.Open(fullURL); err != nil {
					// Browser open failed
					log.Warn("Failed to open browser automatically", "error", err)
				} else {
					browserOpened = true
				}
			}

			if browserOpened {
				// Browser opened successfully
				fmt.Printf("Browser opened automatically.\n\n")
				fmt.Printf("If the browser window does not open automatically, please open this URL to complete authentication:\n")
				fmt.Printf("  %s\n\n", fullURL)
				fmt.Printf("Alternatively, open this URL to enter the code manually:\n")
				fmt.Printf("  %s\n\n", verificationURI)
				fmt.Printf("Enter this code when prompted:\n")
				fmt.Printf("  %s\n\n", userCode)
			} else {
				// Browser failed to open or not available - show manual instructions
				fmt.Printf("Please open this URL in your browser to complete authentication:\n")
				fmt.Printf("  %s\n\n", fullURL)
				fmt.Printf("Alternatively, open this URL to enter the code manually:\n")
				fmt.Printf("  %s\n\n", verificationURI)
				fmt.Printf("Enter this code when prompted:\n")
				fmt.Printf("  %s\n\n", userCode)
			}

			fmt.Printf("Waiting for authorization...\n")
			return nil
		}

	case auth.GrantTypeAuthorizationCode:
		// Initialize AuthorizationCode struct if it doesn't exist
		if cfg.Auth.AuthorizationCode == nil {
			cfg.Auth.AuthorizationCode = &config.AuthorizationCode{}
		}

		// Set custom authorization code handler
		cfg.Auth.AuthorizationCode.OnOpenBrowser = func(url string) error {
			log.Info("Authorization required", "authorization_url", url)

			// Authorization code flow REQUIRES a browser
			if !canOpenBrowser {
				return fmt.Errorf("authorization code flow requires a browser, but no browser is available in this environment")
			}

			// We have a browser - try to open it
			fmt.Printf("\n=== PingOne MCP Server OAuth 2.0 Authorization ===\n")

			if err := browser.Open(url); err != nil {
				// Browser open failed - this is a critical error for auth code flow
				log.Error("Failed to open browser", "error", err)
				return fmt.Errorf("failed to open browser for authorization: %w", err)
			}

			// Browser opened successfully
			fmt.Printf("Browser opened automatically.\n")
			fmt.Printf("If the browser window does not open automatically, please open this URL to complete authentication:\n")
			fmt.Printf("  %s\n\n", url)
			fmt.Printf("Waiting for authorization callback...\n")
			return nil
		}

		// Simple custom HTML messages
		cfg.Auth.AuthorizationCode.CustomHTMLSuccess = "<html><body><h1>PingOne MCP Server</h1><p>Authentication successful! You can close this window.</p></body></html>"
		cfg.Auth.AuthorizationCode.CustomHTMLError = "<html><body><h1>PingOne MCP Server</h1><p>Authentication failed. Please try again.</p></body></html>"
	}
}

type PingOneClientAuthWrapperFactory struct {
	serverVersion string
	environmentId string
}

func NewPingOneClientAuthWrapperFactory(serverVersion, environmentId string) *PingOneClientAuthWrapperFactory {
	return &PingOneClientAuthWrapperFactory{
		serverVersion: serverVersion,
		environmentId: environmentId,
	}
}

func (f *PingOneClientAuthWrapperFactory) NewAuthClient() (AuthClient, error) {
	return NewPingOneClientAuthWrapper(f.serverVersion, f.environmentId), nil
}
