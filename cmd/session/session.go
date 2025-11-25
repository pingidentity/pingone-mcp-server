// Copyright © 2025 Ping Identity Corporation

package session

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/spf13/cobra"
)

const commandName = "session"

func NewCommand(tokenStoreFactory tokenstore.TokenStoreFactory) *cobra.Command {
	var storeTypeFlag string

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Display current session information",
		Long:  "Display the current authentication session information stored locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")

			if tokenStoreFactory == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStoreFactory is nil in session command"))
			}

			storeType, err := tokenstore.ParseStoreType(storeTypeFlag)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			tokenStore, err := tokenStoreFactory.NewTokenStore(storeType)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			sessionExists, err := tokenStore.HasSession()
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if !sessionExists {
				log.Println("No existing login session found.")
				return nil
			}

			authSession, err := tokenStore.GetSession()
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}
			if authSession == nil {
				// Should not happen as we checked HasSession above
				return errs.NewCommandError(commandName, errors.New("token store indicated session exists but returned nil session"))
			}

			logger.FromContext(cmd.Context()).Debug("Local auth session retrieved", slog.String("sessionId", authSession.SessionId))

			fmt.Println("Current Session Information:")
			fmt.Printf("  Session ID: %s\n", authSession.SessionId)
			fmt.Printf("  Expiry: %s\n", authSession.Expiry.Format(time.RFC3339))

			if authSession.Expiry.Before(time.Now()) {
				fmt.Println("  Token Status: Expired")
			} else {
				fmt.Printf("  Token Status: Active (expires in %s)\n", time.Until(authSession.Expiry).Truncate(time.Second))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&storeTypeFlag, "store-type", tokenstore.StoreTypeKeychain.String(), "Token store type to use (keychain or file)")

	return cmd
}
