// Copyright © 2025 Ping Identity Corporation

package filter

// Filter holds configuration for filtering which tools and tool collections to make available in the server
type Filter struct {
	// If true, only read-only tools will be included by the filter
	ReadOnly                bool
	IncludedTools           []string
	ExcludedTools           []string
	IncludedToolCollections []string
	ExcludedToolCollections []string
}

func NewFilter(readOnly bool, includedTools, excludedTools, includedToolCollections, excludedToolCollections []string) *Filter {
	return &Filter{
		ReadOnly:                readOnly,
		IncludedTools:           includedTools,
		ExcludedTools:           excludedTools,
		IncludedToolCollections: includedToolCollections,
		ExcludedToolCollections: excludedToolCollections,
	}
}

func PassthroughFilter() *Filter {
	return &Filter{
		ReadOnly: false,
	}
}

func (f *Filter) ShouldIncludeTool(toolName string, toolIsReadOnly bool) bool {
	return ShouldInclude(toolName, f.IncludedTools, f.ExcludedTools) && (!f.ReadOnly || toolIsReadOnly)
}

func (f *Filter) ShouldIncludeCollection(collectionName string) bool {
	return ShouldInclude(collectionName, f.IncludedToolCollections, f.ExcludedToolCollections)
}

// Filter a name given an include list (whitelist) and an exclude list (blacklist).
// Exclusion takes priority for duplicates.
// If the inclusion list is empty, all names that aren't explicitly excluded will be included.
func ShouldInclude(name string, included []string, excluded []string) bool {
	// Exclusion takes priority
	for _, t := range excluded {
		if t == name {
			return false
		}
	}

	// Default to inclusion if inclusion list is empty
	if len(included) == 0 {
		return true
	}
	for _, t := range included {
		if t == name {
			return true
		}
	}
	return false
}
