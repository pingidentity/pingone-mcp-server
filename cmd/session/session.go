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

func NewCommand(tokenStore tokenstore.TokenStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   commandName,
		Short: "Display current session information",
		Long:  "Display the current authentication session information stored locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InitCommandLogger(cmd, commandName)
			logger.FromContext(cmd.Context()).Debug("Command invoked")

			if tokenStore == nil {
				return errs.NewCommandError(commandName, errors.New("provided tokenStore is nil in session command"))
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

	return cmd
}
