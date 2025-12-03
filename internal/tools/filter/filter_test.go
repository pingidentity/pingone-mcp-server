// Copyright © 2025 Ping Identity Corporation

package filter_test

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		include  []string
		exclude  []string
		expected bool
	}{
		{
			name:     "Included",
			itemName: "A",
			include:  []string{"A", "B", "C"},
			exclude:  []string{},
			expected: true,
		},
		{
			name:     "Excluded",
			itemName: "A",
			include:  []string{"B", "C"},
			exclude:  []string{"A"},
			expected: false,
		},
		{
			name:     "Not included",
			itemName: "D",
			include:  []string{"A", "B", "C"},
			exclude:  []string{},
			expected: false,
		},
		{
			name:     "Exclude takes priority",
			itemName: "A",
			include:  []string{"A"},
			exclude:  []string{"A"},
			expected: false,
		},
		{
			name:     "Empty inclusion list allows all",
			itemName: "Z",
			include:  []string{},
			exclude:  []string{"A"},
			expected: true,
		},
		{
			name:     "Default to inclusion",
			itemName: "Z",
			include:  nil,
			exclude:  nil,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := filter.ShouldInclude(test.itemName, test.include, test.exclude)
			if actual != test.expected {
				t.Errorf("Expected %t, got %t", test.expected, actual)
			}
		})
	}
}

func TestFilterStruct(t *testing.T) {
	tests := []struct {
		name                    string
		readOnly                bool
		includedTools           []string
		excludedTools           []string
		includedToolCollections []string
		excludedToolCollections []string
		testToolIsReadOnly      bool
		testToolName            string
		testCollectionName      string
		expectedTool            bool
		expectedCollection      bool
	}{
		{
			name:               "default filter allows all",
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       true,
			expectedCollection: true,
		},
		{
			name:               "include specific tool",
			includedTools:      []string{"test-tool"},
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       true,
			expectedCollection: true,
		},
		{
			name:               "exclude specific tool",
			excludedTools:      []string{"test-tool"},
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       false,
			expectedCollection: true,
		},
		{
			name:                    "include specific collection",
			includedToolCollections: []string{"test-collection"},
			testToolName:            "test-tool",
			testCollectionName:      "test-collection",
			expectedTool:            true,
			expectedCollection:      true,
		},
		{
			name:                    "exclude specific collection",
			excludedToolCollections: []string{"test-collection"},
			testToolName:            "test-tool",
			testCollectionName:      "test-collection",
			expectedTool:            true,
			expectedCollection:      false,
		},
		{
			name:               "read-only mode includes read-only tool",
			readOnly:           true,
			testToolIsReadOnly: true,
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       true,
			expectedCollection: true,
		},
		{
			name:               "read-only mode excludes non-read-only tool",
			readOnly:           true,
			testToolIsReadOnly: false,
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       false,
			expectedCollection: true,
		},
		{
			name:               "read-only mode with included tools still filters by read-only",
			readOnly:           true,
			includedTools:      []string{"test-tool"},
			testToolIsReadOnly: false,
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       false,
			expectedCollection: true,
		},
		{
			name:               "read-only mode with excluded read-only tool",
			readOnly:           true,
			excludedTools:      []string{"test-tool"},
			testToolIsReadOnly: true,
			testToolName:       "test-tool",
			testCollectionName: "test-collection",
			expectedTool:       false,
			expectedCollection: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := filter.NewFilter(test.readOnly, test.includedTools, test.excludedTools, test.includedToolCollections, test.excludedToolCollections)

			testToolDef := &types.ToolDefinition{
				McpTool: &mcp.Tool{
					Name: test.testToolName,
					Annotations: &mcp.ToolAnnotations{
						ReadOnlyHint: test.testToolIsReadOnly,
					},
				},
			}

			actualTool := f.ShouldIncludeTool(testToolDef)
			if actualTool != test.expectedTool {
				t.Errorf("ShouldIncludeTool: Expected %t, got %t", test.expectedTool, actualTool)
			}

			actualCollection := f.ShouldIncludeCollection(test.testCollectionName)
			if actualCollection != test.expectedCollection {
				t.Errorf("ShouldIncludeCollection: Expected %t, got %t", test.expectedCollection, actualCollection)
			}
		})
	}
}
