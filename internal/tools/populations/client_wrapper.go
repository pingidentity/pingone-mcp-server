// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/patrickcping/pingone-go-sdk-v2/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/collections"
)

var _ PopulationsClient = &PingOneClientPopulationsWrapper{}
var _ PopulationsClientFactory = &PingOneClientPopulationsWrapperFactory{}

type PingOneClientPopulationsWrapper struct {
	client *pingone.Client
}

type PingOneClientPopulationsWrapperFactory struct {
	clientFactory legacy.ClientFactory
	tokenStore    tokenstore.TokenStore
}

func NewPingOneClientPopulationsWrapper(client *pingone.Client) *PingOneClientPopulationsWrapper {
	return &PingOneClientPopulationsWrapper{client: client}
}

func NewPingOneClientPopulationsWrapperFactory(clientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore) *PingOneClientPopulationsWrapperFactory {
	return &PingOneClientPopulationsWrapperFactory{
		clientFactory: clientFactory,
		tokenStore:    tokenStore,
	}
}

func (f *PingOneClientPopulationsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (PopulationsClient, error) {
	client, err := collections.InitializeAuthenticatedLegacyClient(ctx, f.clientFactory, f.tokenStore)
	if err != nil {
		return nil, err
	}
	return NewPingOneClientPopulationsWrapper(client), nil
}

func (p *PingOneClientPopulationsWrapper) GetPopulations(ctx context.Context, environmentId uuid.UUID, filter *string) (management.EntityArrayPagedIterator, error) {
	if p.client == nil {
		return nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.ManagementAPIClient.PopulationsApi.ReadAllPopulations(ctx, environmentId.String())

	if filter != nil && *filter != "" {
		getRequest = getRequest.Filter(*filter)
	}

	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve populations",
		slog.String("environmentId", environmentId.String()),
	)
	return getRequest.Execute(), nil
}

func (p *PingOneClientPopulationsWrapper) CreatePopulation(ctx context.Context, environmentId uuid.UUID, createRequest management.Population) (*management.Population, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	postRequest := p.client.ManagementAPIClient.PopulationsApi.CreatePopulation(ctx, environmentId.String()).Population(createRequest)
	postRequest = postRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	postRequest = postRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to create population",
		slog.String("environmentId", environmentId.String()),
		slog.String("name", createRequest.Name),
	)
	return postRequest.Execute()
}

func (p *PingOneClientPopulationsWrapper) GetPopulationById(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID) (*management.Population, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.ManagementAPIClient.PopulationsApi.ReadOnePopulation(ctx, environmentId.String(), populationId.String())
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve population by ID",
		slog.String("environmentId", environmentId.String()),
		slog.String("populationId", populationId.String()),
	)
	return getRequest.Execute()
}

func (p *PingOneClientPopulationsWrapper) UpdatePopulationById(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID, updateRequest management.Population) (*management.Population, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	putRequest := p.client.ManagementAPIClient.PopulationsApi.UpdatePopulation(ctx, environmentId.String(), populationId.String()).Population(updateRequest)
	putRequest = putRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	putRequest = putRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to update population by ID",
		slog.String("environmentId", environmentId.String()),
		slog.String("populationId", populationId.String()),
	)
	return putRequest.Execute()
}
