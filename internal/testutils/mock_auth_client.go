// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"
	"errors"

	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"golang.org/x/oauth2"
)

var _ client.AuthClient = &mockAuthClient{}
var _ client.AuthClientFactory = &mockAuthClientFactory{}

type mockAuthClient struct {
	returnErr                    error
	authorizationCodeTokenSource oauth2.TokenSource
	deviceCodeTokenSource        oauth2.TokenSource
}

func NewMockAuthClient(returnErr error, authorizationCodeTokenSource, deviceCodeTokenSource oauth2.TokenSource) *mockAuthClient {
	return &mockAuthClient{returnErr: returnErr, authorizationCodeTokenSource: authorizationCodeTokenSource, deviceCodeTokenSource: deviceCodeTokenSource}
}

func (m *mockAuthClient) TokenSource(_ context.Context, grantType auth.GrantType) (oauth2.TokenSource, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	switch grantType {
	case auth.GrantTypeAuthorizationCode:
		return m.authorizationCodeTokenSource, nil
	case auth.GrantTypeDeviceCode:
		return m.deviceCodeTokenSource, nil
	}
	return nil, errors.New("unsupported grant type in mock auth client")
}

type mockAuthClientFactory struct {
	returnErr                    error
	authorizationCodeTokenSource oauth2.TokenSource
	deviceCodeTokenSource        oauth2.TokenSource
}

func NewMockAuthClientFactory(returnErr error, authorizationCodeTokenSource, deviceCodeTokenSource oauth2.TokenSource) *mockAuthClientFactory {
	return &mockAuthClientFactory{returnErr: returnErr, authorizationCodeTokenSource: authorizationCodeTokenSource, deviceCodeTokenSource: deviceCodeTokenSource}
}

func (f *mockAuthClientFactory) NewAuthClient() (client.AuthClient, error) {
	return NewMockAuthClient(f.returnErr, f.authorizationCodeTokenSource, f.deviceCodeTokenSource), nil
}
