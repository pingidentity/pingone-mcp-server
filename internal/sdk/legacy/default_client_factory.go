// Copyright © 2025 Ping Identity Corporation

package legacy

import (
	"context"
	"fmt"

	"github.com/patrickcping/pingone-go-sdk-v2/pingone"
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

func (f *DefaultClientFactory) NewClient(ctx context.Context, accessToken string) (*pingone.Client, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("provided accessToken is empty, legacy client cannot be initialized")
	}
	config := &pingone.Config{}
	userAgentSuffix := audit.PingOneAPIUserAgent(f.serverVersion)
	config.UserAgentSuffix = &userAgentSuffix
	config.AccessToken = &accessToken
	apiClient, err := config.APIClient(ctx)
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}
