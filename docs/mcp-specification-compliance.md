# MCP Specification Compliance

This document explains how the PingOne MCP Server enforces compliance with the [Model Context Protocol (MCP) Specification](https://spec.modelcontextprotocol.io/).

## Overview

The MCP specification requires that servers use **structured logging only** to stderr. Direct output to stdout/stderr via `fmt.Printf`, `fmt.Println`, `print`, or `println` will break MCP protocol communication.

## Automated Enforcement

The project enforces MCP specification compliance through multiple layers:

### 1. golangci-lint with forbidigo

The `.golangci.yml` configuration uses the `forbidigo` linter to prevent `fmt.Printf`, `fmt.Println`, `print`, and `println` in MCP server code.

**Protected packages** (where fmt.Printf is forbidden):
- `internal/server`
- `internal/tools`
- `internal/sdk`
- `internal/auth/client`
- `internal/auth/login`
- `internal/auth/logout`

**Allowed packages** (where fmt.Printf is permitted):
- `cmd/` - CLI commands for user output
- `internal/errs` - Error formatting
- `internal/logger` - Logger implementation
- `*_test.go` - Test files

### 2. Makefile Validation

The `Makefile` includes an `mcp-spec-check` target that performs grep-based validation:

```makefile
make mcp-spec-check
```

This check:
- Scans protected packages for `fmt.Printf` and related functions
- Excludes test files and allowed packages
- Provides clear error messages with remediation guidance

### 3. Comprehensive Validation

Run all validation checks together:

```makefile
make validate-all
```

This runs:
- `make test` - Unit and integration tests
- `make lint` - Code quality and linting (includes forbidigo checks)
- `make mcp-spec-check` - MCP specification compliance

## Proper Logging Patterns

### ✅ CORRECT - Structured Logging

In MCP server code (`internal/server`, `internal/tools`, `internal/sdk`, `internal/auth`):

```go
import "github.com/pingidentity/pingone-mcp-server/internal/logger"

// Get logger from context
log := logger.FromContext(ctx)

// Use structured logging with key-value pairs
log.Info("Operation completed", "resource_id", resourceID, "status", "success")
log.Debug("Processing request", "tool_name", toolName)
log.Error("Operation failed", "error", err)
log.Warn("Deprecated feature used", "feature", featureName)
```

### ❌ INCORRECT - Direct Output

Never use these in MCP server code:

```go
// These break MCP protocol communication
fmt.Printf("Operation completed\n")        // ❌ Breaks MCP protocol
fmt.Println("Processing...")                // ❌ Breaks MCP protocol
print("Debug info")                         // ❌ Breaks MCP protocol
println("Error occurred")                   // ❌ Breaks MCP protocol
```

### ✅ ALLOWED - CLI Commands

In `cmd/` packages (CLI commands for users):

```go
// These are fine for CLI output to users
fmt.Println("Session Information:")  // ✅ OK for CLI output
fmt.Printf("  Status: %s\n", status) // ✅ OK for CLI output
```

## Why This Matters

The MCP specification requires structured logging because:

1. **Protocol Communication**: MCP uses stdin/stdout for protocol messages. Direct `fmt.Printf` output interferes with protocol communication.

2. **Structured Data**: Structured logging provides consistent, parsable output that can be filtered, searched, and analyzed.

3. **Contextual Information**: Logger from context includes request tracking, session IDs, and transaction IDs automatically.

4. **Debug Control**: Structured logging can be controlled via log levels without code changes.

## CI/CD Integration

The compliance checks are integrated into the development workflow:

1. **Pre-commit**: Developers should run `make lint` before committing
2. **Pull Request**: CI runs `make validate-all` on all PRs
3. **Code Review**: Reviewers verify compliance with MCP patterns

## Frequently Asked Questions

### Q: What if I legitimately need to output to stdout?

**A:** The short answer is: **you don't need to in MCP server code**.

The linting rules already exempt the appropriate locations:

1. **CLI Commands (`cmd/` directory)**: Use `fmt.Printf` freely for user-facing output
   ```go
   // cmd/session/session.go - This is fine
   fmt.Println("Session Information:")
   fmt.Printf("  Status: %s\n", status)
   ```

2. **Test Files**: Use `fmt.Printf` for test debugging
   ```go
   // any_test.go - This is fine
   fmt.Printf("Test data: %+v\n", testData)
   ```

3. **MCP Server Code**: Use structured logging instead
   ```go
   // internal/tools/mytool.go - Use structured logging
   log := logger.FromContext(ctx)
   log.Info("Processing request", "data", data)
   ```

### Q: How do I output tool results to users?

**A:** Tool results should be returned through MCP responses, not printed:

```go
// ❌ Don't do this
fmt.Printf("Result: %v\n", result)

// ✅ Do this instead
return &mcp.CallToolResponse{
    Content: []mcp.Content{
        mcp.TextContent(fmt.Sprintf("Result: %v", result)),
    },
}, nil
```

### Q: What about debugging during development?

**A:** Use structured logging with Debug level:

```go
// ❌ Don't do this
fmt.Printf("Debug: processing user %s\n", userID)

// ✅ Do this instead
log.Debug("Processing user", "user_id", userID, "step", "validation")
```

You can control log verbosity through environment variables without changing code.

### Q: Can I temporarily bypass the check for testing?

**A:** During development only, you can use `nolint` directives:

```go
// TEMPORARY: Remove before committing
// nolint:forbidigo
fmt.Printf("Debug output\n")
```

**Warning**: These should NEVER be committed. PR reviewers will reject them.

### Q: What if I need to add a new exempt package?

**A:** Update `.golangci.yml` to add exemptions, but be very careful:

```yaml
issues:
  exclude-rules:
    - path: ^internal/mypackage/
      linters:
        - forbidigo
```

Most packages should follow MCP logging patterns. Only exempt packages that:
- Provide CLI interfaces to users (`cmd/`)
- Implement logging infrastructure (`internal/logger`)
- Format errors for logging (`internal/errs`)
- Are test utilities

## References

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [GitHub Copilot Instructions](.github/copilot-instructions.md)
- [PR Checklist](../contributing/pr-checklist.md)
- [golangci-lint Configuration](../.golangci.yml)
