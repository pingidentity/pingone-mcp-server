// Copyright © 2025 Ping Identity Corporation

package environments

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
)

type EnvironmentsClient interface {
	GetEnvironments(ctx context.Context, filter *string) (pingone.PagedIterator[pingone.EnvironmentsCollectionResponse], error)
	CreateEnvironment(ctx context.Context, request *pingone.EnvironmentCreateRequest) (*pingone.EnvironmentResponse, *http.Response, error)
	GetEnvironment(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentResponse, *http.Response, error)
	UpdateEnvironment(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentReplaceRequest) (*pingone.EnvironmentResponse, *http.Response, error)
	GetEnvironmentServices(ctx context.Context, environmentId uuid.UUID) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error)
	UpdateEnvironmentServices(ctx context.Context, environmentId uuid.UUID, request *pingone.EnvironmentBillOfMaterialsReplaceRequest) (*pingone.EnvironmentBillOfMaterialsResponse, *http.Response, error)
}

type EnvironmentsClientFactory interface {
	GetAuthenticatedClient(ctx context.Context) (EnvironmentsClient, error)
}
