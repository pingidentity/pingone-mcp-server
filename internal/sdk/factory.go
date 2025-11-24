// Copyright © 2025 Ping Identity Corporation

package sdk

import (
	"github.com/pingidentity/pingone-go-client/pingone"
)

type ClientFactory interface {
	NewClient(accessToken string) (*pingone.APIClient, error)
}
