// Copyright © 2026 Ping Identity Corporation

package testutils

func Pointer[T any](v T) *T {
	return &v
}
