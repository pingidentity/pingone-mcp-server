// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/audit"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/collections"
)

var _ EnvironmentsClient = &PingOneClientEnvironmentsWrapper{}
var _ EnvironmentsClientFactory = &PingOneClientEnvironmentsWrapperFactory{}

type PingOneClientEnvironmentsWrapper struct {
	client *pingone.APIClient
}

type PingOneClientEnvironmentsWrapperFactory struct {
	clientFactory sdk.ClientFactory
	tokenStore    tokenstore.TokenStore
}

func NewPingOneClientEnvironmentsWrapper(client *pingone.APIClient) *PingOneClientEnvironmentsWrapper {
	return &PingOneClientEnvironmentsWrapper{client: client}
}

func NewPingOneClientEnvironmentsWrapperFactory(clientFactory sdk.ClientFactory, tokenStore tokenstore.TokenStore) *PingOneClientEnvironmentsWrapperFactory {
	return &PingOneClientEnvironmentsWrapperFactory{
		clientFactory: clientFactory,
		tokenStore:    tokenStore,
	}
}

func (f *PingOneClientEnvironmentsWrapperFactory) GetAuthenticatedClient(ctx context.Context) (EnvironmentsClient, error) {
	client, err := collections.InitializeAuthenticatedClient(f.clientFactory, f.tokenStore)
	if err != nil {
		return nil, err
	}
	return NewPingOneClientEnvironmentsWrapper(client), nil
}

func (p *PingOneClientEnvironmentsWrapper) GetEnvironments(ctx context.Context, filter *string) (pingone.PagedIterator[pingone.EnvironmentsCollectionResponse], error) {
	if p.client == nil {
		return nil, errors.New("PingOne client is not initialized")
	}
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve environments")

	req := p.client.EnvironmentsApi.GetEnvironments(ctx)

	if filter != nil && *filter != "" {
		req = req.Filter(*filter)
	}

	req = req.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	req = req.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))

	return req.Execute(), nil
}

func (p *PingOneClientEnvironmentsWrapper) CreateEnvironment(ctx context.Context, request *pingone.EnvironmentCreateRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	if request == nil {
		return nil, nil, errors.New("environment create request is nil")
	}
	createRequest := p.client.EnvironmentsApi.CreateEnvironment(ctx).EnvironmentCreateRequest(*request)
	createRequest = createRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	createRequest = createRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to create environment",
		slog.String("name", request.Name),
		slog.String("region", string(request.Region)),
		slog.String("type", string(request.Type)))
	return createRequest.Execute()
}

func (p *PingOneClientEnvironmentsWrapper) GetEnvironment(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.EnvironmentsApi.GetEnvironmentById(ctx, environmentId)
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve environment by ID",
		slog.String("environmentId", environmentId.String()))
	return getRequest.Execute()
}

func (p *PingOneClientEnvironmentsWrapper) UpdateEnvironment(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentReplaceRequest) (*pingone.EnvironmentResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	if request == nil {
		return nil, nil, errors.New("environment replace request is nil")
	}
	replaceRequest := p.client.EnvironmentsApi.ReplaceEnvironmentById(ctx, environmentId).EnvironmentReplaceRequest(*request)
	replaceRequest = replaceRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	replaceRequest = replaceRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to update environment by ID",
		slog.String("environmentId", environmentId.String()),
		slog.String("name", request.Name),
		slog.String("region", string(request.Region)),
		slog.String("type", string(request.Type)))
	return replaceRequest.Execute()
}

func (p *PingOneClientEnvironmentsWrapper) GetEnvironmentServices(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	getRequest := p.client.EnvironmentsApi.GetBillOfMaterialsByEnvironmentId(ctx, environmentId)
	getRequest = getRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	getRequest = getRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to retrieve environment services by ID",
		slog.String("environmentId", environmentId.String()))
	return getRequest.Execute()
}

func (p *PingOneClientEnvironmentsWrapper) UpdateEnvironmentServices(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentBillOfMaterialsReplaceRequest) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error) {
	if p.client == nil {
		return nil, nil, errors.New("PingOne client is not initialized")
	}
	updateRequest := p.client.EnvironmentsApi.ReplaceBillOfMaterialsByEnvironmentId(ctx, environmentId).EnvironmentBillOfMaterialsReplaceRequest(*request)
	updateRequest = updateRequest.XPingExternalSessionID(audit.SessionIdFromContext(ctx))
	updateRequest = updateRequest.XPingExternalTransactionID(audit.TransactionIdFromContext(ctx))
	logger.FromContext(ctx).Debug("Calling PingOne API to update environment services by ID",
		slog.String("environmentId", environmentId.String()))
	return updateRequest.Execute()
}
