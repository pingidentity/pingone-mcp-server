// Copyright © 2025 Ping Identity Corporation

package logger

import (
	"context"
	"log/slog"
)

type contextKey string

const LoggerContextKey contextKey = "logger"

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return Logger
	}
	if v := ctx.Value(LoggerContextKey); v != nil {
		if logger, ok := v.(*slog.Logger); ok {
			return logger
		}
	}
	return Logger
}
