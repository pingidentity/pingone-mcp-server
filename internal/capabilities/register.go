package capabilities

import (
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/collections"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/populations"
)

// getDefaultCollections creates SDK collections
func getDefaultCollections() []collections.Collection {
	return []collections.Collection{
		&environments.EnvironmentsCollection{},
	}
}

// getLegacySdkCollections creates legacy SDK collections
func getLegacySdkCollections() []collections.LegacySdkCollection {
	return []collections.LegacySdkCollection{
		&populations.PopulationsCollection{},
		&applications.ApplicationsCollection{},
	}
}
