// Copyright © 2025 Ping Identity Corporation

package tokenstore

import "fmt"

type StoreType int

const (
	_ StoreType = iota
	StoreTypeKeychain
	StoreTypeFile
)

func (t StoreType) String() string {
	switch t {
	case StoreTypeKeychain:
		return "keychain"
	case StoreTypeFile:
		return "file"
	default:
		return "unknown"
	}
}

func ParseStoreType(s string) (StoreType, error) {
	switch s {
	case "keychain":
		return StoreTypeKeychain, nil
	case "file":
		return StoreTypeFile, nil
	default:
		return 0, fmt.Errorf("unable to parse store type from string: %s", s)
	}
}
