// Copyright © 2026 Ping Identity Corporation

package types

type DynamicResourceDefinition struct {
	StaticResourceDefinition
	// ValidationPolicy allows modification of in-built validation rules and constraints for the tool's execution
	ValidationPolicy *DynamicResourceValidationPolicy
}
