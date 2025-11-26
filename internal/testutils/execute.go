// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"context"
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/cmd"
	"github.com/pingidentity/pingone-mcp-server/cmd/login"
	"github.com/pingidentity/pingone-mcp-server/cmd/logout"
	"github.com/pingidentity/pingone-mcp-server/cmd/run"
	"github.com/pingidentity/pingone-mcp-server/cmd/session"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/spf13/cobra"
)

const TestServerVersion = "test"

func prepareTestCommand(cmd *cobra.Command, args ...string) {
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
}

func ExecuteCliRootCommand(t *testing.T, ctx context.Context, args ...string) (err error) {
	t.Helper()

	root := cmd.NewRootCommand(TestServerVersion)
	prepareTestCommand(root, args...)

	return root.ExecuteContext(ctx)
}

func ExecuteCliRunCommand(t *testing.T, ctx context.Context, tokenStoreFactory tokenstore.TokenStoreFactory, clientFactory sdk.ClientFactory, legacyClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, transport mcp.Transport, args ...string) (err error) {
	t.Helper()

	runCmd := run.NewCommand(tokenStoreFactory, clientFactory, legacyClientFactory, authClientFactory, transport)
	prepareTestCommand(runCmd, args...)

	return runCmd.ExecuteContext(ctx)
}

func ExecuteCliLoginCommand(t *testing.T, ctx context.Context, authClientFactory client.AuthClientFactory, tokenStoreFactory tokenstore.TokenStoreFactory, args ...string) (err error) {
	t.Helper()

	loginCmd := login.NewCommand(authClientFactory, tokenStoreFactory)
	prepareTestCommand(loginCmd, args...)

	return loginCmd.ExecuteContext(ctx)
}

func ExecuteCliLogoutCommand(t *testing.T, ctx context.Context, tokenStoreFactory tokenstore.TokenStoreFactory, args ...string) (err error) {
	t.Helper()

	logoutCmd := logout.NewCommand(tokenStoreFactory)
	prepareTestCommand(logoutCmd, args...)

	return logoutCmd.ExecuteContext(ctx)
}

func ExecuteCliSessionCommand(t *testing.T, ctx context.Context, tokenStoreFactory tokenstore.TokenStoreFactory, args ...string) (err error) {
	t.Helper()

	sessionCmd := session.NewCommand(tokenStoreFactory)
	prepareTestCommand(sessionCmd, args...)

	return sessionCmd.ExecuteContext(ctx)
}
