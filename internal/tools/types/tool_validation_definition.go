// Copyright © 2025 Ping Identity Corporation

package types

type ToolValidationPolicy struct {
	// AllowProductionEnvironmentWrite when set to true, allows the tool to make write operations on production-type environments.
	// When false (default), write operations on PRODUCTION environments are blocked to prevent unintended changes.
	// This modifier only applies to WRITE operations; READ operations are governed by AllowProductionEnvironmentRead.
	AllowProductionEnvironmentWrite bool
	// AllowProductionEnvironmentRead when set to true, allows the tool to make read operations on production-type environments.
	// When false (default), read operations on PRODUCTION environments are blocked.
	// This modifier only applies to READ operations; WRITE operations are governed by AllowProductionEnvironmentWrite.
	AllowProductionEnvironmentRead bool
	// ProductionEnvironmentNotApplicable when set to true, indicates that the tool does not operate on environments
	// and therefore production environment validation should be skipped entirely.
	// This is typically used for tools that don't have an environmentId parameter (e.g., list_environments) or operate at the organization level.
	// When true, both AllowProductionEnvironmentWrite and AllowProductionEnvironmentRead are ignored.
	ProductionEnvironmentNotApplicable bool
}
