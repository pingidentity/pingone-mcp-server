// Copyright © 2025 Ping Identity Corporation

package sdk

import (
	"github.com/pingidentity/pingone-go-client/pingone"
)

var _ ClientFactory = &EmptyClientFactory{}

// The EmptyClientFactory creates an invalid api client with no configuration set.
// This can be useful for unit tests that do not actually invoke operations on the SDK.
type EmptyClientFactory struct{}

func NewEmptyClientFactory() *EmptyClientFactory {
	return &EmptyClientFactory{}
}

func (n *EmptyClientFactory) NewClient(_ string) (*pingone.APIClient, error) {
	return &pingone.APIClient{}, nil
}
