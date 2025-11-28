// Copyright © 2025 Ping Identity Corporation

package validation

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-go-client/pingone"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/environments"
)

// OperationType represents the type of operation being performed on an environment.
type OperationType string

const (
	// OperationTypeRead represents read-only operations (GET requests).
	OperationTypeRead OperationType = "READ"
	// OperationTypeWrite represents write operations (POST, PUT, PATCH, DELETE requests).
	OperationTypeWrite OperationType = "WRITE"
)

// EnvironmentValidator validates that an environment exists and is accessible.
// For write operations, it also enforces that the environment is not a PRODUCTION environment.
type EnvironmentValidator interface {
	ValidateEnvironment(ctx context.Context, environmentId uuid.UUID, operationType OperationType) error
}

// CachingEnvironmentValidator validates environments with caching to reduce API calls.
// Only PRODUCTION environments are cached after successful validation, as PRODUCTION
// environments cannot be downgraded to SANDBOX (ensuring cache consistency).
// SANDBOX environments are not cached since they can be upgraded to PRODUCTION.
// For write operations, it enforces that the environment type is not PRODUCTION.
type CachingEnvironmentValidator struct {
	clientFactory environments.EnvironmentsClientFactory
	cache         sync.Map // uuid.UUID -> *pingone.EnvironmentResponse
}

// NewCachingEnvironmentValidator creates a new caching environment validator.
// The validator uses the provided client factory to fetch environment information
// and caches successful validations to improve performance.
func NewCachingEnvironmentValidator(clientFactory environments.EnvironmentsClientFactory) *CachingEnvironmentValidator {
	return &CachingEnvironmentValidator{
		clientFactory: clientFactory,
		cache:         sync.Map{},
	}
}

// ValidateEnvironment checks if the given environment exists and is accessible.
// It first checks the cache, and if not found, makes an API call to verify the environment.
// Only PRODUCTION environments are cached after successful validation, as they cannot be
// downgraded to SANDBOX (ensuring cache consistency).
// For write operations, it enforces that the environment type is not PRODUCTION.
// Returns an error if:
//   - The environment does not exist or is not accessible
//   - The operation is a write operation and the environment type is PRODUCTION
func (v *CachingEnvironmentValidator) ValidateEnvironment(ctx context.Context, environmentId uuid.UUID, operationType OperationType) error {
	// Check cache first
	if cachedEnv, ok := v.cache.Load(environmentId); ok {
		env := cachedEnv.(*pingone.EnvironmentResponse)
		return v.validateEnvironmentType(env, operationType)
	}

	// Get authenticated client
	client, err := v.clientFactory.GetAuthenticatedClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated client for validation: %w", err)
	}

	// Validate with API
	env, resp, err := client.GetEnvironmentById(ctx, environmentId)
	if err != nil {
		return fmt.Errorf("environment %s does not exist or is not accessible: %w", environmentId, err)
	}

	// Cache successful validation only for PRODUCTION environments
	// PRODUCTION environments cannot be downgraded to SANDBOX, so caching is safe
	// SANDBOX environments can be upgraded to PRODUCTION, so we should not cache them
	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && env != nil && env.Type == pingone.ENVIRONMENTTYPEVALUE_PRODUCTION {
		v.cache.Store(environmentId, env)
	}

	// Validate environment type for write operations
	return v.validateEnvironmentType(env, operationType)
}

// validateEnvironmentType checks if the operation type is allowed for the given environment.
// to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed.
func (v *CachingEnvironmentValidator) validateEnvironmentType(env *pingone.EnvironmentResponse, operationType OperationType) error {
	if env == nil {
		return fmt.Errorf("environment response is nil")
	}

	// to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed
	if operationType == OperationTypeWrite && env.Type == pingone.ENVIRONMENTTYPEVALUE_PRODUCTION {
		return fmt.Errorf("to safeguard against unintended or breaking changes to PRODUCTION environments, write operations are not allowed (environment ID: %s, name: %s)", env.Id, env.Name)
	}

	return nil
}

// ClearCache removes all cached environment validations.
// This can be useful in testing or when you want to force revalidation.
func (v *CachingEnvironmentValidator) ClearCache() {
	v.cache.Range(func(key, value interface{}) bool {
		v.cache.Delete(key)
		return true
	})
}

// RemoveFromCache removes a specific environment from the cache.
// This can be useful when an environment is deleted or becomes inaccessible.
func (v *CachingEnvironmentValidator) RemoveFromCache(environmentId uuid.UUID) {
	v.cache.Delete(environmentId)
}
