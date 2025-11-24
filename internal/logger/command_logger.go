// Copyright © 2025 Ping Identity Corporation

package logger

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func InitCommandLogger(cmd *cobra.Command, commandName string) {
	cmd.SetContext(ContextWithLogger(cmd.Context(), FromContext(cmd.Context()).With(slog.String("command", commandName))))
}
