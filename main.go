// Copyright © 2025 Ping Identity Corporation

package main

import (
	"context"
	"os"

	"github.com/pingidentity/pingone-mcp-server/cmd"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
)

// These variables will be set by the goreleaser configuration at build time.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Build full version string with commit and date information
	fullVersion := version
	if commit != "unknown" {
		fullVersion += " (commit: " + commit + ")"
	}
	if date != "unknown" {
		fullVersion += " (built: " + date + ")"
	}
	
	rootCmd := cmd.NewRootCommand(fullVersion)
	if err := rootCmd.Execute(); err != nil {
		errs.Log(context.Background(), err)
		os.Exit(1)
	}
}
