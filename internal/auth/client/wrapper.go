// Copyright © 2025 Ping Identity Corporation

package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func (p *PingOneClientAuthWrapper) TokenSource(ctx context.Context, grantType auth.GrantType, mcpServerSession *mcp.ServerSession) (oauth2.TokenSource, error) {
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
	err := p.configureHeadlessHandlers(ctx, clientConfig, grantType, mcpServerSession)
	if err != nil {
		return nil, err
	}

	pingoneConfig := pingone.NewConfiguration(clientConfig)
	pingoneConfig.AppendUserAgent(audit.PingOneAPIUserAgent(p.serverVersion))

	return pingoneConfig.Service.TokenSource(ctx)
}

func (p *PingOneClientAuthWrapper) BrowserLoginAvailable(grantType auth.GrantType) bool {
	switch grantType {
	case auth.GrantTypeAuthorizationCode, auth.GrantTypeDeviceCode:
		// These grant types can use browser login if a browser is available
		return browser.CanOpen()
	default:
		return false
	}
}

// configureHeadlessHandlers sets up custom UX handlers for headless MCP server operation.
// This provides environment-aware browser handling:
// - If browser is available: opens browser for both auth code and device code flows
// - If no browser: auth code fails (requires browser), device code prints instructions
func (p *PingOneClientAuthWrapper) configureHeadlessHandlers(ctx context.Context, cfg *config.Configuration, grantType auth.GrantType, mcpServerSession *mcp.ServerSession) error {
	log := logger.FromContext(ctx)

	// Check if we're in an environment with browser support
	canOpenBrowser := browser.CanOpen()

	switch grantType {
	case auth.GrantTypeDeviceCode:
		// If mcpServerSession is nil, we cannot proceed
		if mcpServerSession == nil {
			return fmt.Errorf("no MCP server session found. The MCP server session is required to elicit the URL for device code flow")
		}

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
			log.Info("=== PingOne MCP Server OAuth 2.0 Authorization ===")

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
				log.Info("Browser opened automatically")
				log.Info("If the browser window does not open automatically, please open this URL to complete authentication", "url", fullURL)
				log.Info("Alternatively, open this URL to enter the code manually", "url", verificationURI)
				log.Info("Enter this code when prompted", "code", userCode)
			} else {
				// Browser failed to open or not available - elicit with url mode elicitation
				elicitID := uuid.New().String()
				elicitResult, err := mcpServerSession.Elicit(
					ctx,
					&mcp.ElicitParams{
						Message:       "Open the following URL in your browser to complete authentication",
						URL:           fullURL,
						ElicitationID: elicitID,
					},
				)

				if err != nil {
					return fmt.Errorf("failed to elicit device code URL: %w", err)
				}

				switch elicitResult.Action {
				case "decline":
					return fmt.Errorf("device code URL elicitation was not completed. The request was declined.")
				case "cancel":
					return fmt.Errorf("device code URL elicitation was not completed. The request was canceled.")
				}
			}

			log.Info("Waiting for authorization...")
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
			log.Info("=== PingOne MCP Server OAuth 2.0 Authorization ===")

			if err := browser.Open(url); err != nil {
				// Browser open failed - this is a critical error for auth code flow
				log.Error("Failed to open browser", "error", err)
				return fmt.Errorf("failed to open browser for authorization: %w", err)
			}

			// Browser opened successfully
			log.Info("Browser opened automatically")
			log.Info("If the browser window does not open automatically, please open this URL to complete authentication", "url", url)
			log.Info("Waiting for authorization callback")
			return nil
		}

		// Custom page data for authorization result pages
		projectName := "PingOne MCP Server"
		cfg.Auth.AuthorizationCode.CustomPageDataSuccess = &config.AuthResultPageData{
			ProjectName: projectName,
			Heading:     "Authorization Successful",
			Message:     fmt.Sprintf("The %s can now access PingOne management APIs with your role permissions.", projectName),
		}
		cfg.Auth.AuthorizationCode.CustomPageDataError = &config.AuthResultPageData{
			ProjectName: projectName,
			Heading:     "Authorization Failed",
			Message:     "An error occurred.",
		}
	}

	return nil
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
