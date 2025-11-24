// Copyright © 2025 Ping Identity Corporation

package testutils

import "errors"

// APIErrorTestCase represents a test case for API error scenarios
type APIErrorTestCase struct {
	Name            string
	StatusCode      int
	ApiError        error
	WantErrContains string
}

func CommonAPIErrorTestCases() []APIErrorTestCase {
	return []APIErrorTestCase{
		{
			Name:            "500 Internal Server Error",
			StatusCode:      500,
			ApiError:        errors.New("internal server error"),
			WantErrContains: "internal server error",
		},
		{
			Name:            "401 Unauthorized",
			StatusCode:      401,
			ApiError:        errors.New("unauthorized"),
			WantErrContains: "unauthorized",
		},
		{
			Name:            "403 Forbidden",
			StatusCode:      403,
			ApiError:        errors.New("forbidden"),
			WantErrContains: "forbidden",
		},
	}
}
