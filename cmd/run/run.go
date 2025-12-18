// Copyright © 2025 Ping Identity Corporation

package run

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk"
	"github.com/pingidentity/pingone-mcp-server/internal/sdk/legacy"
	"github.com/pingidentity/pingone-mcp-server/internal/server"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/pingidentity/pingone-mcp-server/internal/tools"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/filter"
	"github.com/pingidentity/pingone-mcp-server/internal/tools/types"
	"github.com/spf13/cobra"
)

const commandName = "run"

func NewCommand(tokenStoreFactory tokenstore.TokenStoreFactory, clientFactory sdk.ClientFactory, legacyClientFactory legacy.ClientFactory, authClientFactory client.AuthClientFactory, transport mcp.Transport, version string) *cobra.Command {
	var includedTools []string
	var excludedTools []string
	var includedToolCollections []string
	var excludedToolCollections []string
	var disableReadOnly bool
	var grantTypeFlag string
	var storeTypeFlag string

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Start the PingOne MCP server",
		Long: `Start the PingOne MCP server to handle Model Context Protocol requests.
The server will communicate over stdin/stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")
			if tokenStoreFactory == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStoreFactory is nil in run command"))
			}

			grantType, err := auth.ParseGrantType(grantTypeFlag)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			logger.FromContext(cmd.Context()).Debug("Using grant type", slog.String("grantType", grantType.String()))

			storeType, err := tokenstore.ParseStoreType(storeTypeFlag)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			tokenStore, err := tokenStoreFactory.NewTokenStore(storeType)
			if err != nil {
				return errs.NewCommandError(commandName, err)
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
					logger.FromContext(cmd.Context()).Debug("Active session is expired, authentication will be refreshed when a tool is invoked", slog.String("sessionId", session.SessionId))
				} else {
					logger.FromContext(cmd.Context()).Debug("Active session found", slog.String("sessionId", session.SessionId))
				}
			} else {
				logger.FromContext(cmd.Context()).Debug("No active session found, authentication will be refreshed when a tool is invoked")
			}

			toolFilter := filter.NewFilter(!disableReadOnly, includedTools, excludedTools, includedToolCollections, excludedToolCollections)

			logger.FromContext(cmd.Context()).Debug("Run command tool filter built",
				slog.Bool("disableReadOnly", disableReadOnly),
				slog.Any("includedTools", includedTools),
				slog.Any("excludedTools", excludedTools),
				slog.Any("includedToolCollections", includedToolCollections),
				slog.Any("excludedToolCollections", excludedToolCollections))

			// Warn if user may have specified write tools but forgot --disable-read-only
			if !disableReadOnly && len(includedTools) > 0 {
				warnAboutPotentialWriteToolsFiltered(cmd.Context(), includedTools)
			}

			err = server.Start(cmd.Context(), version, transport, clientFactory, legacyClientFactory, authClientFactory, tokenStore, toolFilter, grantType)
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
	cmd.Flags().StringVar(&grantTypeFlag, "grant-type", auth.GrantTypeAuthorizationCode.String(), "OAuth grant type to use for authentication (authorization_code or device_code). device_code is recommended in headless or CI/CD environments")
	cmd.Flags().StringVar(&storeTypeFlag, "store-type", tokenstore.StoreTypeKeychain.String(), "Token store type to use (keychain or file)")

	return cmd
}

// warnAboutPotentialWriteToolsFiltered checks if any of the included tools are write tools
// and warns the user that they will be filtered out due to read-only mode being enabled
func warnAboutPotentialWriteToolsFiltered(ctx context.Context, includedTools []string) {
	allTools := tools.ListTools()

	// Create a map of tool names to their definitions for quick lookup
	toolMap := make(map[string]*types.ToolDefinition)
	for i := range allTools {
		toolMap[allTools[i].McpTool.Name] = &allTools[i]
	}

	var writeToolsSpecified []string
	for _, toolName := range includedTools {
		if toolDef, exists := toolMap[toolName]; exists {
			if !toolDef.IsReadOnly() {
				writeToolsSpecified = append(writeToolsSpecified, toolName)
			}
		}
	}

	if len(writeToolsSpecified) > 0 {
		logger.FromContext(ctx).Warn("Write tools specified in --include-tools will be excluded due to read-only mode",
			slog.Any("writeTools", writeToolsSpecified),
			slog.String("suggestion", "Add --disable-read-only flag to enable write tools"))
	}
}
