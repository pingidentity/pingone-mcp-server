// Copyright © 2025 Ping Identity Corporation

package main

import (
	"context"
	"os"

	"github.com/pingidentity/pingone-mcp-server/cmd"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
)

// This version will be set by the goreleaser configuration at build time.
var version = "dev"

func main() {
	rootCmd := cmd.NewRootCommand(version)
	if err := rootCmd.Execute(); err != nil {
		errs.Log(context.Background(), err)
		os.Exit(1)
	}
}
