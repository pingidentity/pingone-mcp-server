// Copyright © 2025 Ping Identity Corporation

package client

type AuthClientFactory interface {
	NewAuthClient() (AuthClient, error)
}
