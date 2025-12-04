// Copyright © 2025 Ping Identity Corporation

package cmd

import (
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/cmd/logout"
	"github.com/pingidentity/pingone-mcp-server/cmd/run"
	"github.com/pingidentity/pingone-mcp-server/cmd/session"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/spf13/cobra"
)

const (
	mcpEnvironmentIdEnvVar    = "PINGONE_MCP_ENVIRONMENT_ID"
	clientEnvironmentIdEnvVar = "PINGONE_ENVIRONMENT_ID"
)

func NewRootCommand(serverVersion string) *cobra.Command {
	result := &cobra.Command{
		Use:     "pingone-mcp-server",
		Short:   "The PingOne MCP Server provides Model Context Protocol integration for PingOne",
		Long:    "The PingOne MCP Server provides Model Context Protocol integration for PingOne",
		Version: serverVersion,
	}
	// Create default client factories and token store
	clientFactory := sdk.NewDefaultClientFactory(serverVersion)
	legacyClientFactory := legacy.NewDefaultClientFactory(serverVersion)
	tokenStoreFactory := tokenstore.NewDefaultTokenStoreFactory()
	// Have to workaround mutually exclusive requirement in legacy SDK that
	// prevents setting both access token and environment ID by using an mcp-specific
	// environment variable here, and unsetting the environment variable
	// normally used by the client SDKs.
	os.Unsetenv(clientEnvironmentIdEnvVar)
	mcpEnvironmentId := os.Getenv(mcpEnvironmentIdEnvVar)

	authClientFactory := client.NewPingOneClientAuthWrapperFactory(serverVersion, mcpEnvironmentId)
	// Always run on stdio transport
	result.AddCommand(run.NewCommand(tokenStoreFactory, clientFactory, legacyClientFactory, authClientFactory, &mcp.StdioTransport{}))

	result.AddCommand(logout.NewCommand(tokenStoreFactory))

	result.AddCommand(session.NewCommand(tokenStoreFactory))
	return result
}
