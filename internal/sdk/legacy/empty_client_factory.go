// Copyright © 2025 Ping Identity Corporation

package legacy

import (
	"context"

	"github.com/patrickcping/pingone-go-sdk-v2/pingone"
)

var _ ClientFactory = &EmptyClientFactory{}

// The EmptyClientFactory creates an invalid api client with no configuration set.
// This can be useful for unit tests that do not actually invoke operations on the SDK.
type EmptyClientFactory struct{}

func NewEmptyClientFactory() *EmptyClientFactory {
	return &EmptyClientFactory{}
}

func (n *EmptyClientFactory) NewClient(ctx context.Context, _ string) (*pingone.Client, error) {
	return &pingone.Client{}, nil
}
