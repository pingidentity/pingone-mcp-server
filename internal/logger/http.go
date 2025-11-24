// Copyright © 2025 Ping Identity Corporation

package logger

import (
	"context"
	"log/slog"
	"net/http"
)

func LogHttpResponse(ctx context.Context, httpResp *http.Response) {
	if httpResp == nil {
		return
	}
	attrs := []slog.Attr{
		slog.Int("responseStatusCode", httpResp.StatusCode),
		slog.String("responseStatus", httpResp.Status),
	}
	if httpResp.Request != nil {
		attrs = append(attrs,
			slog.String("requestMethod", httpResp.Request.Method),
			slog.String("requestHost", httpResp.Request.Host),
		)

		if httpResp.Request.URL != nil {
			attrs = append(attrs, slog.String("requestURL", httpResp.Request.URL.Redacted()))
		}

		if httpResp.Request.Header != nil {
			if externalTransactionId := httpResp.Request.Header.Get("X-Ping-External-Transaction-Id"); externalTransactionId != "" {
				attrs = append(attrs, slog.String("requestXPingExternalTransactionId", externalTransactionId))
			}
			if externalSessionId := httpResp.Request.Header.Get("X-Ping-External-Session-Id"); externalSessionId != "" {
				attrs = append(attrs, slog.String("requestXPingExternalSessionId", externalSessionId))
			}
		}
	}

	if httpResp.Header != nil {
		if contentType := httpResp.Header.Get("Content-Type"); contentType != "" {
			attrs = append(attrs, slog.String("responseContentType", contentType))
		}
		if contentLength := httpResp.Header.Get("Content-Length"); contentLength != "" {
			attrs = append(attrs, slog.String("responseContentLength", contentLength))
		}
	}
	FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "Received HTTP response", attrs...)
}
