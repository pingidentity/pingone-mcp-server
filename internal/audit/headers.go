// Copyright © 2025 Ping Identity Corporation

package audit

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	transactionIdKey contextKey = "TransactionId"
	sessionIdKey     contextKey = "SessionId"
)

func GenerateTransactionId() string {
	return uuid.New().String()
}

func ContextWithTransactionId(ctx context.Context, transactionId string) context.Context {
	return context.WithValue(ctx, transactionIdKey, transactionId)
}

func ContextWithSessionId(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, sessionIdKey, sessionId)
}

func TransactionIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(transactionIdKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func SessionIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(sessionIdKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
