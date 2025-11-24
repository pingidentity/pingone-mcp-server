// Copyright © 2025 Ping Identity Corporation

package logger

import (
	"log/slog"
	"os"
	"strings"
)

const debugEnvVar = "PINGONE_MCP_DEBUG"

var Logger *slog.Logger

func init() {
	// Check if debug mode is enabled via environment variable
	envVarValue := os.Getenv(debugEnvVar)
	debugEnabled := strings.EqualFold(envVarValue, "true")

	handlerOptions := &slog.HandlerOptions{}
	if debugEnabled {
		handlerOptions.Level = slog.LevelDebug
	}

	// MCP servers using stdio transport can't log to stdout
	Logger = slog.New(slog.NewTextHandler(os.Stderr, handlerOptions))

	if debugEnabled {
		Logger.Debug("Debug logging enabled by PINGONE_MCP_DEBUG environment variable", slog.String("PINGONE_MCP_DEBUG", envVarValue))
	}
}
