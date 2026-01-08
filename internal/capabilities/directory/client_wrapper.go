// Copyright © 2025 Ping Identity Corporation

package directory

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
)

var _ DirectoryClient = &PingOneClientDirectoryWrapper{}
var _ DirectoryClientFactory = &PingOneClientDirectoryWrapperFactory{}

type PingOneClientDirectoryWrapper struct {
	client *pingone.APIClient
}

type PingOneClientDirectoryWrapperFactory struct {
	clientFactory sdk.ClientFactory
	tokenStore    tokenstore.TokenStore
}

func NewPingOneClientDirectoryWrapper(client *pingone.APIClient) *PingOneClientDirectoryWrapper {
	return &PingOneClientDirectoryWrapper{client: client}
}

func NewPingOneClientDirectoryWrapperFactory(clientFactory sdk.ClientFactory, tokenStore tokenstore.TokenStore) *PingOneClientDirectoryWrapperFactory {
	return &PingOneClientDirectoryWrapperFactory{
		clientFactory: clientFactory,
		tokenStore:    tokenStore,
	}
}

func (f *PingOneClientDirectoryWrapperFactory) GetAuthenticatedClient(ctx context.Context) (DirectoryClient, error) {
	client, err := collections.InitializeAuthenticatedClient(f.clientFactory, f.tokenStore)
	if err != nil {
		return nil, err
	}
	return NewPingOneClientDirectoryWrapper(client), nil
}

func (p *PingOneClientDirectoryWrapper) GetTotalIdentitiesByEnvironmentId(ctx context.Context, environmentId uuid.UUID, filter string) (*pingone.DirectoryTotalIdentitiesCountCollectionResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.DirectoryTotalIdentitiesApi.GetTotalIdentities(ctx, environmentId).Filter(filter)
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve total identities by environment ID",
		slog.String("environmentId", environmentId.String()))
	return getRequest.Execute()
}
