// Copyright © 2025 Ping Identity Corporation

package errs

import (
	"fmt"
	"net/http"
)

// ApiError represents a structured API error with HTTP response details
type ApiError struct {
	OriginalError error
	StatusCode    int
	Status        string
	Method        string
	URL           string
}

func (e *ApiError) Error() string {
	if e.OriginalError == nil && e.StatusCode == 0 {
		return "unknown API error"
	}

	var msg string
	if e.OriginalError != nil {
		msg = e.OriginalError.Error()
	}

	// Format with HTTP response information if available
	if e.StatusCode != 0 {
		httpInfo := fmt.Sprintf("HTTP %d %s", e.StatusCode, e.Status)
		if e.Method != "" && e.URL != "" {
			httpInfo = fmt.Sprintf("%s %s: %s", e.Method, e.URL, httpInfo)
		}

		if msg != "" {
			msg = fmt.Sprintf("%s (%s)", msg, httpInfo)
		} else {
			msg = httpInfo
		}
	}

	return msg
}

func (e *ApiError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.OriginalError
}

func NewApiError(httpResp *http.Response, err error) error {
	apiErr := &ApiError{
		OriginalError: err,
	}

	// Extract HTTP response information if available
	if httpResp != nil {
		apiErr.StatusCode = httpResp.StatusCode
		apiErr.Status = httpResp.Status

		if httpResp.Request != nil {
			apiErr.Method = httpResp.Request.Method
			if httpResp.Request.URL != nil {
				apiErr.URL = httpResp.Request.URL.String()
			}
		}
	}

	return apiErr
}
