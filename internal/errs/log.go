// Copyright © 2025 Ping Identity Corporation

package errs

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pingidentity/pingone-mcp-server/internal/logger"
)

func Log(ctx context.Context, err error) {
	LogWithLogger(logger.FromContext(ctx), ctx, err)
}

func LogWithLogger(logger *slog.Logger, ctx context.Context, err error) {
	if err == nil {
		return
	}

	var attrs []slog.Attr

	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		attrs = append(attrs,
			slog.String("errorType", "apiError"),
			slog.Int("statusCode", apiErr.StatusCode),
			slog.String("status", apiErr.Status),
			slog.String("method", apiErr.Method),
			slog.String("url", apiErr.URL),
		)
		if apiErr.OriginalError != nil {
			attrs = append(attrs, slog.String("originalError", apiErr.OriginalError.Error()))
		}
	}

	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		attrs = append(attrs,
			slog.String("errorType", "toolError"),
			slog.String("toolName", toolErr.ToolName),
		)
		if toolErr.OriginalError != nil {
			attrs = append(attrs, slog.String("originalError", toolErr.OriginalError.Error()))
		}
	}

	var cmdErr *CommandError
	if errors.As(err, &cmdErr) {
		attrs = append(attrs,
			slog.String("errorType", "commandError"),
			slog.String("commandName", cmdErr.CommandName),
		)
		if cmdErr.OriginalError != nil {
			attrs = append(attrs, slog.String("originalError", cmdErr.OriginalError.Error()))
		}
	}

	// If no specific error type was found, just log as generic error
	if len(attrs) == 0 {
		attrs = append(attrs, slog.String("errorType", "generic"))
	}

	logger.LogAttrs(ctx, slog.LevelError, err.Error(), attrs...)
}
