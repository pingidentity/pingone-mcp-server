// Copyright © 2025 Ping Identity Corporation

package run

import (
	"errors"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/server"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/spf13/cobra"
)

const commandName = "run"

func NewCommand(tokenStore tokenstore.TokenStore, clientFactory sdk.ClientFactory, legacyClientFactory legacy.ClientFactory, transport mcp.Transport) *cobra.Command {
	var includedTools []string
	var excludedTools []string
	var includedToolCollections []string
	var excludedToolCollections []string
	var disableReadOnly bool

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Start the PingOne MCP server",
		Long: `Start the PingOne MCP server to handle Model Context Protocol requests.
The server will communicate over stdin/stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")
			if tokenStore == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStore is nil in run command"))
			}

			hasSession, err := tokenStore.HasSession()
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if hasSession {
				session, err := tokenStore.GetSession()
				if err != nil {
					return errs.NewCommandError(commandName, err)
				}
				if session == nil {
					// Shouldn't happen
					return errs.NewCommandError(commandName, errors.New("active session is nil"))
				}
				if session.Expiry.Before(time.Now()) {
					logger.FromContext(cmd.Context()).Warn("Active session is expired, please run the login command", slog.String("sessionId", session.SessionId))
				} else {
					logger.FromContext(cmd.Context()).Debug("Active session found", slog.String("sessionId", session.SessionId))
				}
			} else {
				logger.FromContext(cmd.Context()).Warn("No active session found, please run the login command")
			}

			toolFilter := filter.NewFilter(!disableReadOnly, includedTools, excludedTools, includedToolCollections, excludedToolCollections)

			logger.FromContext(cmd.Context()).Debug("Run command tool filter built",
				slog.Bool("disableReadOnly", disableReadOnly),
				slog.Any("includedTools", includedTools),
				slog.Any("excludedTools", excludedTools),
				slog.Any("includedToolCollections", includedToolCollections),
				slog.Any("excludedToolCollections", excludedToolCollections))

			err = server.Start(cmd.Context(), transport, clientFactory, legacyClientFactory, tokenStore, toolFilter)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&includedTools, "include-tools", []string{}, "A list of tools to enable")
	cmd.Flags().StringSliceVar(&excludedTools, "exclude-tools", []string{}, "A list of tools to disable")
	cmd.Flags().StringSliceVar(&includedToolCollections, "include-tool-collections", []string{}, "A list of tool collections to enable")
	cmd.Flags().StringSliceVar(&excludedToolCollections, "exclude-tool-collections", []string{}, "A list of tool collections to disable")
	cmd.Flags().BoolVar(&disableReadOnly, "disable-read-only", false, "Disable read-only mode to include write tools")

	return cmd
}
