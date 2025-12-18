// Copyright © 2025 Ping Identity Corporation

package testutils

import "context"

// Just returns the context as-is, without modification or error
func MockContextInitializer() func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	}
}

// Returns a context initializer that always returns the specified error
func MockContextInitializerWithError(err error) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return ctx, err
	}
}
