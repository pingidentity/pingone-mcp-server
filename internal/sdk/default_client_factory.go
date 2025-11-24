// Copyright © 2025 Ping Identity Corporation

package sdk

import (
	"fmt"

	"github.com/pingidentity/pingone-go-client/config"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
)

var _ ClientFactory = &DefaultClientFactory{}

type DefaultClientFactory struct {
	serverVersion string
}

func NewDefaultClientFactory(serverVersion string) *DefaultClientFactory {
	return &DefaultClientFactory{
		serverVersion: serverVersion,
	}
}

func (f *DefaultClientFactory) NewClient(accessToken string) (*pingone.APIClient, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("provided accessToken is empty, client cannot be initialized")
	}
	clientConfig := config.NewConfiguration().WithAccessToken(accessToken)
	pingOneConfig := pingone.NewConfiguration(clientConfig)
	pingOneConfig.AppendUserAgent(audit.PingOneAPIUserAgent(f.serverVersion))
	apiClient, err := pingone.NewAPIClient(pingOneConfig)
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}
