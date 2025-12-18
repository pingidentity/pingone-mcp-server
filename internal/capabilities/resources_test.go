// Copyright © 2025 Ping Identity Corporation

package capabilities_test

import (
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/capabilities"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/applications"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/environments"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/populations"
	"github.com/pingidentity/pingone-mcp-server/internal/capabilities/types"
)

func TestAllDynamicResourcesRegistered(t *testing.T) {
	// Get tools from ListDynamicResources that are actually registered with the server
	allDynamicResources := capabilities.ListDynamicResources()

	// Get tools from individual collections
	var expectedDynamicResources []types.DynamicResourceDefinition
	expectedDynamicResources = append(expectedDynamicResources, (&environments.EnvironmentsCollection{}).ListDynamicResources()...)
	expectedDynamicResources = append(expectedDynamicResources, (&populations.PopulationsCollection{}).ListDynamicResources()...)
	expectedDynamicResources = append(expectedDynamicResources, (&applications.ApplicationsCollection{}).ListDynamicResources()...)

	// Verify lists match
	if len(allDynamicResources) != len(expectedDynamicResources) {
		t.Errorf("ListDynamicResources() returned %d dynamic resources, but individual collections returned %d dynamic resources", len(allDynamicResources), len(expectedDynamicResources))
	}

	expectedToolNames := make(map[string]bool)
	for _, tool := range expectedDynamicResources {
		expectedToolNames[tool.McpResource.Name] = true
	}

	for _, tool := range allDynamicResources {
		if !expectedToolNames[tool.McpResource.Name] {
			t.Errorf("ListDynamicResources() returned unexpected dynamic resource: %s", tool.McpResource.Name)
		}
	}

	actualToolNames := make(map[string]bool)
	for _, tool := range allDynamicResources {
		actualToolNames[tool.McpResource.Name] = true
	}

	for _, tool := range expectedDynamicResources {
		if !actualToolNames[tool.McpResource.Name] {
			t.Errorf("ListDynamicResources() missing expected dynamic resource, collection may not be registered: %s", tool.McpResource.Name)
		}
	}
}
