// Copyright © 2025 Ping Identity Corporation

package applications

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

var _ ApplicationsClient = &PingOneClientApplicationsWrapper{}
var _ ApplicationsClientFactory = &PingOneClientApplicationsWrapperFactory{}

type PingOneClientApplicationsWrapper struct {
	client *pingone.Client
}

type PingOneClientApplicationsWrapperFactory struct {
	clientFactory legacy.ClientFactory
	tokenStore    tokenstore.TokenStore
}

func NewPingOneClientApplicationsWrapper(client *pingone.Client) *PingOneClientApplicationsWrapper {
	return &PingOneClientApplicationsWrapper{client: client}
}

func NewPingOneClientApplicationsWrapperFactory(clientFactory legacy.ClientFactory, tokenStore tokenstore.TokenStore) *PingOneClientApplicationsWrapperFactory {
	return &PingOneClientApplicationsWrapperFactory{
		clientFactory: clientFactory,
		tokenStore:    tokenStore,
	}
}

func (f *PingOneClientApplicationsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (ApplicationsClient, error) {
	client, err := collections.InitializeAuthenticatedLegacyClient(ctx, f.clientFactory, f.tokenStore)
	if err != nil {
		return nil, err
	}
	return NewPingOneClientApplicationsWrapper(client), nil
}

func (p *PingOneClientApplicationsWrapper) GetApplications(ctx context.Context, environmentId uuid.UUID) (management.EntityArrayPagedIterator, error) {
	if p.client == nil {
		return nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.ManagementAPIClient.ApplicationsApi.ReadAllApplications(ctx, environmentId.String())
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve applications",
		slog.String("environmentId", environmentId.String()),
	)
	return getRequest.Execute(), nil
}

func (p *PingOneClientApplicationsWrapper) CreateApplication(ctx context.Context, environmentId uuid.UUID, app management.CreateApplicationRequest) (*management.CreateApplication201Response, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	createRequest := p.client.ManagementAPIClient.ApplicationsApi.CreateApplication(ctx, environmentId.String()).CreateApplicationRequest(app)
	createRequest = createRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	createRequest = createRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to create application",
		slog.String("environmentId", environmentId.String()),
	)
	return createRequest.Execute()
}

func (p *PingOneClientApplicationsWrapper) GetApplication(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID) (*management.ReadOneApplication200Response, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.ManagementAPIClient.ApplicationsApi.ReadOneApplication(ctx, environmentId.String(), applicationId.String())
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve application",
		slog.String("environmentId", environmentId.String()),
		slog.String("applicationId", applicationId.String()),
	)
	return getRequest.Execute()
}

func (p *PingOneClientApplicationsWrapper) UpdateApplicationById(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID, app management.UpdateApplicationRequest) (*management.ReadOneApplication200Response, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	updateRequest := p.client.ManagementAPIClient.ApplicationsApi.UpdateApplication(ctx, environmentId.String(), applicationId.String()).UpdateApplicationRequest(app)
	updateRequest = updateRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	updateRequest = updateRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to update application",
		slog.String("environmentId", environmentId.String()),
		slog.String("applicationId", applicationId.String()),
	)
	return updateRequest.Execute()
}
