// Copyright © 2025 Ping Identity Corporation

package errs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pingidentity/pingone-go-client/pingone"
)

// ApiError represents a structured API error with HTTP response details.
// It wraps the original error and enriches it with HTTP context including
// status codes, request methods, and URLs for better debugging and logging.
//
// ApiError implements the error interface and supports error unwrapping,
// allowing it to integrate with Go's standard error handling patterns.
type ApiError struct {
	// OriginalError is the underlying error that occurred
	OriginalError error
	// StatusCode is the HTTP status code from the response (e.g., 404, 500)
	StatusCode int
	// Status is the HTTP status text (e.g., "Not Found", "Internal Server Error")
	Status string
	// Method is the HTTP method used in the request (e.g., "GET", "POST")
	Method string
	// URL is the full URL of the request that caused the error
	URL string
	// ResponseBody contains the raw response body from the API
	ResponseBody string
}

func (e *ApiError) Error() string {
	if e.OriginalError == nil && e.StatusCode == 0 && e.ResponseBody == "" {
		return "unknown API error"
	}

	var msg string
	if e.OriginalError != nil {
		originalErrorMsg := ""

		originalErrorMsg = parsePingOneErrorMsg(e.OriginalError)

		if originalErrorMsg != "" {
			msg = originalErrorMsg
		} else {
			// Fallback to the original error message
			msg = e.OriginalError.Error()
			// Append response body if available and not already parsed as a pingone error
			if e.ResponseBody != "" {
				if msg != "" {
					msg = fmt.Sprintf("%s. Response body: %s", msg, e.ResponseBody)
				} else {
					msg = fmt.Sprintf("Response body: %s", e.ResponseBody)
				}
			}
		}
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

		// Read and store response body if available
		if httpResp.Body != nil {
			bodyBytes, readErr := io.ReadAll(httpResp.Body)
			httpResp.Body.Close()
			if readErr == nil && len(bodyBytes) > 0 {
				apiErr.ResponseBody = string(bodyBytes)
			}
		}
	}

	return apiErr
}

// parsePingOneErrorMsg extracts and formats detailed error messages from PingOne API errors.
// It handles multiple PingOne error types and provides comprehensive error information including:
// - Validation constraint details (allowed patterns, values, ranges)
// - Inner error conditions and requirements
// - Multi-level error detail hierarchies
//
// Supported error types:
//   - pingone.NotFoundError: Returns the basic error message
//   - pingone.BadRequestError: Returns message with detailed validation constraints
//   - pingone.UnsupportedMediaTypeError: Returns message with simple detail list
//
// Returns an empty string if the error is not a recognized PingOne error type.
func parsePingOneErrorMsg(err error) string {
	if msg := parseNotFoundError(err); msg != "" {
		return msg
	}
	if msg := parseBadRequestError(err); msg != "" {
		return msg
	}
	if msg := parseUnsupportedMediaTypeError(err); msg != "" {
		return msg
	}
	return ""
}

// parseNotFoundError extracts the error message from NotFoundError types.
func parseNotFoundError(err error) string {
	var notFoundError pingone.NotFoundError
	if errors.As(err, &notFoundError) {
		return notFoundError.GetMessage()
	}
	return ""
}

// parseBadRequestError extracts and formats error messages from BadRequestError types,
// including detailed validation constraints and inner error conditions.
func parseBadRequestError(err error) string {
	var badRequestError pingone.BadRequestError
	if !errors.As(err, &badRequestError) {
		return ""
	}

	msg := badRequestError.GetMessage()
	if badRequestError.HasDetails() && len(badRequestError.GetDetails()) > 0 {
		msg += formatErrorDetails(badRequestError.GetDetails())
	}
	return msg
}

// parseUnsupportedMediaTypeError extracts and formats error messages from
// UnsupportedMediaTypeError types with simple detail formatting.
func parseUnsupportedMediaTypeError(err error) string {
	var unsupportedMediaTypeError pingone.UnsupportedMediaTypeError
	if !errors.As(err, &unsupportedMediaTypeError) {
		return ""
	}

	msg := unsupportedMediaTypeError.GetMessage()
	if unsupportedMediaTypeError.HasDetails() && len(unsupportedMediaTypeError.GetDetails()) > 0 {
		var builder strings.Builder
		builder.WriteString(msg)
		for i, detail := range unsupportedMediaTypeError.GetDetails() {
			builder.WriteString(fmt.Sprintf(" %d: %s", i+1, detail.GetMessage()))
		}
		return builder.String()
	}
	return msg
}

// formatErrorDetails formats a slice of error details into a comprehensive error message
// with validation constraints and inner error conditions.
func formatErrorDetails(details []pingone.BadRequestErrorDetail) string {
	if len(details) == 0 {
		return ""
	}

	detailMessages := make([]string, 0, len(details))

	for i, detail := range details {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Error detail %d: %s", i+1, detail.GetMessage()))

		if detail.HasInnerError() {
			innerConditions := formatInnerErrorConditions(detail.GetInnerError())
			if innerConditions != "" {
				builder.WriteString(" (")
				builder.WriteString(innerConditions)
				builder.WriteString(")")
			}
		}

		detailMessages = append(detailMessages, builder.String())
	}

	return " [" + strings.Join(detailMessages, "], [") + "]"
}

// formatInnerErrorConditions formats inner error conditions into a human-readable string
// with validation constraints such as allowed patterns, values, and ranges.
func formatInnerErrorConditions(innerErr pingone.BadRequestErrorDetailInnerError) string {
	conditions := make([]string, 0, 10)

	if innerErr.AllowedPattern != nil {
		conditions = append(conditions, fmt.Sprintf("allowed pattern: %s", *innerErr.AllowedPattern))
	}

	if innerErr.AllowedValues != nil {
		conditions = append(conditions, fmt.Sprintf("allowed values: %s", strings.Join(innerErr.AllowedValues, ", ")))
	}

	if innerErr.Claim != nil {
		conditions = append(conditions, fmt.Sprintf("claim: %s", *innerErr.Claim))
	}

	if innerErr.ExistingId != nil {
		conditions = append(conditions, fmt.Sprintf("existing ID: %s", *innerErr.ExistingId))
	}

	if innerErr.MaximumValue != nil {
		conditions = append(conditions, fmt.Sprintf("maximum value: %f", *innerErr.MaximumValue))
	}

	if innerErr.QuotaLimit != nil {
		conditions = append(conditions, fmt.Sprintf("quota limit: %f", *innerErr.QuotaLimit))
	}

	if innerErr.QuotaResetTime != nil {
		conditions = append(conditions, fmt.Sprintf("quota reset time: %s", *innerErr.QuotaResetTime))
	}

	if innerErr.RangeMaximumValue != nil {
		conditions = append(conditions, fmt.Sprintf("range maximum value: %f", *innerErr.RangeMaximumValue))
	}

	if innerErr.RangeMinimumValue != nil {
		conditions = append(conditions, fmt.Sprintf("range minimum value: %f", *innerErr.RangeMinimumValue))
	}

	if innerErr.AdditionalProperties != nil {
		conditions = append(conditions, fmt.Sprintf("additional conditions: %v", innerErr.AdditionalProperties))
	}

	return strings.Join(conditions, "; ")
}
