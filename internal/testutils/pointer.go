// Copyright © 2025 Ping Identity Corporation

package testutils

func Pointer[T any](v T) *T {
	return &v
}
