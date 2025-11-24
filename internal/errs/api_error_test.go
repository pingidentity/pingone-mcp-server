// Copyright © 2025 Ping Identity Corporation

package errs_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/pingidentity/pingone-mcp-server/internal/errs"
)

func TestApiError_Error(t *testing.T) {
	tests := []struct {
		name          string
		originalError error
		statusCode    int
		status        string
		method        string
		url           string
		expected      string
	}{
		{
			name:     "no error and no HTTP response",
			expected: "unknown API error",
		},
		{
			name:          "original error only",
			originalError: errors.New("connection failed"),
			expected:      "connection failed",
		},
		{
			name:       "HTTP response only",
			statusCode: 404,
			status:     "Not Found",
			expected:   "HTTP 404 Not Found",
		},
		{
			name:          "original error with HTTP response",
			originalError: errors.New("authentication failed"),
			statusCode:    401,
			status:        "Unauthorized",
			expected:      "authentication failed (HTTP 401 Unauthorized)",
		},
		{
			name:       "HTTP response with method and URL",
			statusCode: 500,
			status:     "Internal Server Error",
			method:     "GET",
			url:        "https://api.pingone.com/v1/environments",
			expected:   "GET https://api.pingone.com/v1/environments: HTTP 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := &errs.ApiError{
				OriginalError: tt.originalError,
				StatusCode:    tt.statusCode,
				Status:        tt.status,
				Method:        tt.method,
				URL:           tt.url,
			}

			result := apiErr.Error()
			if result != tt.expected {
				t.Errorf("Expected: %q, got: %q", tt.expected, result)
			}
		})
	}
}

func TestNewApiError(t *testing.T) {
	tests := []struct {
		name         string
		httpResp     *http.Response
		originalErr  error
		expectedType string
		checkFunc    func(t *testing.T, err error)
	}{
		{
			name:         "nil response and nil error",
			httpResp:     nil,
			originalErr:  nil,
			expectedType: "*errs.ApiError",
			checkFunc: func(t *testing.T, err error) {
				apiErr := err.(*errs.ApiError)
				if apiErr.OriginalError != nil {
					t.Errorf("Expected nil original error, got: %v", apiErr.OriginalError)
				}
				if apiErr.StatusCode != 0 {
					t.Errorf("Expected status code 0, got: %d", apiErr.StatusCode)
				}
			},
		},
		{
			name:         "error only",
			httpResp:     nil,
			originalErr:  errors.New("connection failed"),
			expectedType: "*errs.ApiError",
			checkFunc: func(t *testing.T, err error) {
				apiErr := err.(*errs.ApiError)
				if apiErr.OriginalError == nil || apiErr.OriginalError.Error() != "connection failed" {
					t.Errorf("Expected original error 'connection failed', got: %v", apiErr.OriginalError)
				}
			},
		},
		{
			name: "HTTP response with request details",
			httpResp: &http.Response{
				StatusCode: 404,
				Status:     "Not Found",
				Request: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "https",
						Host:   "api.pingone.com",
						Path:   "/v1/environments/123",
					},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"error":"not_found"}`)),
			},
			originalErr:  nil,
			expectedType: "*errs.ApiError",
			checkFunc: func(t *testing.T, err error) {
				apiErr := err.(*errs.ApiError)
				if apiErr.StatusCode != 404 {
					t.Errorf("Expected status code 404, got: %d", apiErr.StatusCode)
				}
				if apiErr.Status != "Not Found" {
					t.Errorf("Expected status 'Not Found', got: %s", apiErr.Status)
				}
				if apiErr.Method != "GET" {
					t.Errorf("Expected method 'GET', got: %s", apiErr.Method)
				}
				if !strings.Contains(apiErr.URL, "api.pingone.com") {
					t.Errorf("Expected URL to contain 'api.pingone.com', got: %s", apiErr.URL)
				}
			},
		},
		{
			name: "HTTP response without request",
			httpResp: &http.Response{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Request:    nil,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			},
			originalErr:  errors.New("server error"),
			expectedType: "*errs.ApiError",
			checkFunc: func(t *testing.T, err error) {
				apiErr := err.(*errs.ApiError)
				if apiErr.StatusCode != 500 {
					t.Errorf("Expected status code 500, got: %d", apiErr.StatusCode)
				}
				if apiErr.Method != "" {
					t.Errorf("Expected empty method, got: %s", apiErr.Method)
				}
				if apiErr.URL != "" {
					t.Errorf("Expected empty URL, got: %s", apiErr.URL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.NewApiError(tt.httpResp, tt.originalErr)

			if err == nil {
				t.Fatal("Expected non-nil error")
			}

			// Check type
			if got := fmt.Sprintf("%T", err); got != tt.expectedType {
				t.Errorf("Expected type %s, got: %s", tt.expectedType, got)
			}

			// Run custom checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}
