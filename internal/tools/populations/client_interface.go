// Copyright © 2025 Ping Identity Corporation

package populations

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

type PopulationsClient interface {
	GetPopulations(ctx context.Context, environmentId uuid.UUID, filter *string) (management.EntityArrayPagedIterator, error)
	CreatePopulation(ctx context.Context, environmentId uuid.UUID, createRequest management.Population) (*management.Population, *http.Response, error)
	GetPopulationById(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID) (*management.Population, *http.Response, error)
	UpdatePopulationById(ctx context.Context, environmentId uuid.UUID, populationId uuid.UUID, updateRequest management.Population) (*management.Population, *http.Response, error)
}

type PopulationsClientFactory interface {
	GetAuthenticatedClient(ctx context.Context) (PopulationsClient, error)
}
