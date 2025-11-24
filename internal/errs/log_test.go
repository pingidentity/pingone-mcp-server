// Copyright © 2025 Ping Identity Corporation

package errs

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestLogError_GenericError(t *testing.T) {
	err := errors.New("this is a generic error")

	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	LogWithLogger(testLogger, context.Background(), err)

	output := buf.String()
	if !strings.Contains(output, "this is a generic error") {
		t.Errorf("Expected stderr to contain 'this is a generic error', got: %s", output)
	}

	if !strings.Contains(output, "errorType=generic") {
		t.Errorf("Expected stderr to contain 'errorType=generic', got: %s", output)
	}

	if !strings.Contains(output, "level=ERROR") {
		t.Errorf("Expected stderr to contain 'level=ERROR', got: %s", output)
	}
}

func TestLogError_ApiError(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "https", Host: "api.pingone.com", Path: "/v1/environments"},
	}
	resp := &http.Response{
		StatusCode: 401,
		Status:     "401 Unauthorized",
		Request:    req,
	}

	originalErr := errors.New("invalid credentials")
	apiErr := NewApiError(resp, originalErr)

	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	LogWithLogger(testLogger, context.Background(), apiErr)

	expectedChecks := []string{
		"level=ERROR",
		"errorType=apiError",
		"statusCode=401",
		"status=\"401 Unauthorized\"",
		"method=GET",
		"url=https://api.pingone.com/v1/environments",
		"originalError=\"invalid credentials\"",
	}

	output := buf.String()
	for _, expected := range expectedChecks {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected stderr to contain '%s', got: %s", expected, output)
		}
	}
}

func TestLogError_ToolError(t *testing.T) {
	originalErr := errors.New("connection timeout")
	toolErr := NewToolError("list_environments", originalErr)

	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	LogWithLogger(testLogger, context.Background(), toolErr)

	expectedChecks := []string{
		"level=ERROR",
		"errorType=toolError",
		"toolName=list_environments",
		"originalError=\"connection timeout\"",
	}

	output := buf.String()
	for _, expected := range expectedChecks {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected stderr to contain '%s', got: %s", expected, output)
		}
	}
}

func TestLogError_CommandError(t *testing.T) {
	originalErr := errors.New("client factory initialization failed")
	cmdErr := NewCommandError("run", originalErr)

	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	LogWithLogger(testLogger, context.Background(), cmdErr)

	expectedChecks := []string{
		"level=ERROR",
		"errorType=commandError",
		"commandName=run",
		"originalError=\"client factory initialization failed\"",
	}

	output := buf.String()
	for _, expected := range expectedChecks {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected stderr to contain '%s', got: %s", expected, output)
		}
	}
}

func TestLogError_NilError(t *testing.T) {
	var buf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	LogWithLogger(testLogger, context.Background(), nil)

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no stderr output for nil error, got: %s", output)
	}
}
