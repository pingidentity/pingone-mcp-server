// Copyright © 2025 Ping Identity Corporation

package auth

import "fmt"

type GrantType int

const (
	_ GrantType = iota
	GrantTypeAuthorizationCode
	GrantTypeDeviceCode
)

func (g GrantType) String() string {
	switch g {
	case GrantTypeAuthorizationCode:
		return "authorization_code"
	case GrantTypeDeviceCode:
		return "device_code"
	default:
		return "unknown"
	}
}

func ParseGrantType(s string) (GrantType, error) {
	switch s {
	case "authorization_code":
		return GrantTypeAuthorizationCode, nil
	case "device_code":
		return GrantTypeDeviceCode, nil
	default:
		return 0, fmt.Errorf("unable to parse grant type from string: %s", s)
	}
}
