// Copyright © 2025 Ping Identity Corporation

package client

import (
	"context"
	"fmt"
	"log"

	"github.com/pingidentity/pingone-go-client/config"
	pingoneOauth2 "github.com/pingidentity/pingone-go-client/oauth2"
	"github.com/pingidentity/pingone-go-client/pingone"
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

	log.Printf("We're on new code")

	// Configure custom UX handlers for headless operation
	p.configureHeadlessHandlers(ctx, clientConfig, grantType)

	pingoneConfig := pingone.NewConfiguration(clientConfig)
	pingoneConfig.AppendUserAgent(audit.PingOneAPIUserAgent(p.serverVersion))

	return pingoneConfig.Service.TokenSource(ctx)
}

// configureHeadlessHandlers sets up custom UX handlers for headless MCP server operation.
// This prevents automatic browser opening and provides custom branded callback pages.
func (p *PingOneClientAuthWrapper) configureHeadlessHandlers(ctx context.Context, cfg *config.Configuration, grantType auth.GrantType) {
	log := logger.FromContext(ctx)

	switch grantType {
	case auth.GrantTypeDeviceCode:
		// Initialize DeviceCode struct if it doesn't exist
		if cfg.Auth.DeviceCode == nil {
			cfg.Auth.DeviceCode = &config.DeviceCode{}
		}

		// Set custom device code prompt handler - don't auto-open browser, just log the URL
		cfg.Auth.DeviceCode.OnDisplayPrompt = func(verificationURI, userCode string) error {
			log.Info("Device authorization required",
				"verification_uri", verificationURI,
				"user_code", userCode)
			fmt.Printf("\n=== PingOne MCP Server Authentication ===\n")
			fmt.Printf("Please open this URL in your browser:\n")
			fmt.Printf("  %s\n\n", verificationURI)
			fmt.Printf("Enter this code when prompted:\n")
			fmt.Printf("  %s\n\n", userCode)
			fmt.Printf("Waiting for authorization...\n")
			return nil
		}

	case auth.GrantTypeAuthorizationCode:
		// Initialize AuthorizationCode struct if it doesn't exist
		if cfg.Auth.AuthorizationCode == nil {
			cfg.Auth.AuthorizationCode = &config.AuthorizationCode{}
		}

		// Set custom authorization code handlers - don't auto-open browser, provide custom HTML
		cfg.Auth.AuthorizationCode.OnOpenBrowser = func(url string) error {
			log.Info("Authorization required", "authorization_url", url)
			fmt.Printf("\n=== PingOne MCP Server Authentication ===\n")
			fmt.Printf("Please open this URL in your browser to authenticate:\n")
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
