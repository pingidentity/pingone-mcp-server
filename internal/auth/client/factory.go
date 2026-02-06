// Copyright © 2026 Ping Identity Corporation

package client

type AuthClientFactory interface {
	NewAuthClient() (AuthClient, error)
}
