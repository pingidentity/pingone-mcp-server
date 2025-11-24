// Copyright © 2025 Ping Identity Corporation

package client

import (
	"context"
	"fmt"

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

	pingoneConfig := pingone.NewConfiguration(clientConfig)
	pingoneConfig.AppendUserAgent(audit.PingOneAPIUserAgent(p.serverVersion))

	return pingoneConfig.Service.TokenSource(ctx)
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
