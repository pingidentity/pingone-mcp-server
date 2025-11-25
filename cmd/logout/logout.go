// Copyright © 2025 Ping Identity Corporation

package logout

import (
	"errors"
	"log"

	"github.com/pingidentity/pingone-mcp-server/internal/auth/logout"
	"github.com/pingidentity/pingone-mcp-server/internal/errs"
	"github.com/pingidentity/pingone-mcp-server/internal/logger"
	"github.com/pingidentity/pingone-mcp-server/internal/tokenstore"
	"github.com/spf13/cobra"
)

const commandName = "logout"

func NewCommand(tokenStoreFactory tokenstore.TokenStoreFactory) *cobra.Command {
	var storeTypeFlag string

	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Logout from PingOne",
		Long:  "Logout from PingOne by revoking the access token and clearing the authentication session stored in the OS keychain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")
			log.Println("Logging out from PingOne...")

			if tokenStoreFactory == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStoreFactory is nil in logout command"))
			}

			storeType, err := tokenstore.ParseStoreType(storeTypeFlag)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			tokenStore, err := tokenStoreFactory.NewTokenStore(storeType)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			err = logout.Logout(cmd.Context(), tokenStore)
			if err != nil {
				return errs.NewCommandError(commandName, err)
			}

			log.Println("Logout completed successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&storeTypeFlag, "store-type", tokenstore.StoreTypeKeychain.String(), "Token store type to use (keychain or file)")

	return cmd
}
