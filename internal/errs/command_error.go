// Copyright © 2025 Ping Identity Corporation

package errs

import "fmt"

// CommandError represents an error that occurred within a specific cobra command
type CommandError struct {
	CommandName   string
	OriginalError error
}

func (e *CommandError) Error() string {
	if e.OriginalError == nil {
		if e.CommandName == "" {
			return "unknown command error"
		} else {
			return fmt.Sprintf("pingone-mcp-server %s command failed: unknown error", e.CommandName)
		}
	}

	msg := e.OriginalError.Error()

	if e.CommandName != "" {
		msg = fmt.Sprintf("pingone-mcp-server %s command failed: %s", e.CommandName, msg)
	}

	return msg
}

func (e *CommandError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.OriginalError
}

func NewCommandError(commandName string, err error) error {
	return &CommandError{
		CommandName:   commandName,
		OriginalError: err,
	}
}
