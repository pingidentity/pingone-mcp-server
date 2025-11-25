// Copyright © 2025 Ping Identity Corporation

package login

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/pingidentity/pingone-mcp-server/internal/auth"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/client"
	"github.com/pingidentity/pingone-mcp-server/internal/auth/logout"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/spf13/cobra"
)

const commandName = "login"
const authTimeout = 5 * time.Minute

func NewCommand(authClientFactory client.AuthClientFactory, tokenStoreFactory tokenstore.TokenStoreFactory) *cobra.Command {
	var grantTypeFlag string
	var storeTypeFlag string

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Login to PingOne",
		Long:  "Login to PingOne to authenticate and store credentials for the MCP server session. Must be run before starting the server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")
			log.Println("Logging in to PingOne...")

			if authClientFactory == nil {
				return errs.NewCommandError(commandName, errors.New("provided authClientFactory is nil in login command"))
			}
			if tokenStoreFactory == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStoreFactory is nil in login command"))
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
			logger.FromContext(cmd.Context()).Debug("Using store type", slog.String("storeType", storeType.String()))

			authClient, err := authClientFactory.NewAuthClient()
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if authClient == nil {
				return errs.NewCommandError(commandName, errors.New("authClientFactory returned nil AuthClient"))
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
				logger.FromContext(cmd.Context()).Info("An existing local auth session was found. Logging out.")
				err = logout.Logout(cmd.Context(), tokenStore)
				if err != nil {
					return errs.NewCommandError(commandName, err)
				}
			}

			// Add a long timeout to the context for the auth process
			authCtx, cancel := context.WithTimeout(cmd.Context(), authTimeout)
			defer cancel()

			tokenSource, err := authClient.TokenSource(authCtx, grantType)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if tokenSource == nil {
				return errs.NewCommandError(commandName, errors.New("authClient returned nil TokenSource"))
			}

			token, err := tokenSource.Token()
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if token == nil {
				// Should not happen
				return errs.NewCommandError(commandName, errors.New("token source returned nil token"))
			}
			logger.FromContext(cmd.Context()).Debug("Access token retrieved", slog.String("expiry", token.Expiry.Format(time.RFC3339)))

			sessionId := uuid.New().String()
			authSession := auth.NewAuthSession(*token, sessionId)
			if err := tokenStore.PutSession(authSession); err != nil {
				return errs.NewCommandError(commandName, err)
			}
			logger.FromContext(cmd.Context()).Debug("Auth session stored")

			log.Println("Login completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&grantTypeFlag, "grant-type", auth.GrantTypeAuthorizationCode.String(), "OAuth grant type to use for authentication (authorization_code or device_code)")
	cmd.Flags().StringVar(&storeTypeFlag, "store-type", tokenstore.StoreTypeKeychain.String(), "Token store type to use (keychain or file)")

	return cmd
}
