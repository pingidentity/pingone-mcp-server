// Copyright © 2025 Ping Identity Corporation

package errs

import "fmt"

// ToolError represents an error that occurred within a specific tool
type ToolError struct {
	ToolName      string
	OriginalError error
}

func (e *ToolError) Error() string {
	if e.OriginalError == nil {
		if e.ToolName == "" {
			return "unknown tool error"
		} else {
			return fmt.Sprintf("pingone-mcp-server %s tool failed: unknown error", e.ToolName)
		}
	}

	msg := e.OriginalError.Error()

	if e.ToolName != "" {
		msg = fmt.Sprintf("pingone-mcp-server %s tool failed: %s", e.ToolName, msg)
	}

	return msg
}

func (e *ToolError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.OriginalError
}

func NewToolError(toolName string, err error) error {
	return &ToolError{
		ToolName:      toolName,
		OriginalError: err,
	}
}
