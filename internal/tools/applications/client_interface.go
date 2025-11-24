// Copyright © 2025 Ping Identity Corporation

package applications

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

type ApplicationsClient interface {
	GetApplications(ctx context.Context, environmentId uuid.UUID) (management.EntityArrayPagedIterator, error)
	CreateApplication(ctx context.Context, environmentId uuid.UUID, app management.CreateApplicationRequest) (*management.CreateApplication201Response, *http.Response, error)
	GetApplication(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID) (*management.ReadOneApplication200Response, *http.Response, error)
	UpdateApplicationById(ctx context.Context, environmentId uuid.UUID, applicationId uuid.UUID, app management.UpdateApplicationRequest) (*management.ReadOneApplication200Response, *http.Response, error)
}

type ApplicationsClientFactory interface {
	GetAuthenticatedClient(ctx context.Context) (ApplicationsClient, error)
}
