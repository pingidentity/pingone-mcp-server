// Copyright © 2025 Ping Identity Corporation

package directory

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
)

type DirectoryClient interface {
	GetTotalIdentitiesByEnvironmentId(ctx context.Context, environmentId uuid.UUID, filter string) (*pingone.DirectoryTotalIdentitiesCountCollectionResponse, *http.Response, error)
}

type DirectoryClientFactory interface {
	GetAuthenticatedClient(ctx context.Context) (DirectoryClient, error)
}
