// Copyright © 2025 Ping Identity Corporation

package legacy

import (
	"context"

	"github.com/patrickcping/pingone-go-sdk-v2/pingone"
)

type ClientFactory interface {
	NewClient(ctx context.Context, accessToken string) (*pingone.Client, error)
}
