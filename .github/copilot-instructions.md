# GitHub Copilot Instructions for PingOne MCP Server

## Project Overview

This project implements a Model Context Protocol (MCP) server for PingOne identity and access management platform. The MCP specification defines standards for communication between AI assistants and external systems.

**Key Resources:**
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

## MCP Specification Compliance

**CRITICAL:** All code MUST comply with the Model Context Protocol specification. This is not optional.

### 1. Logging Standards (MCP Specification Requirement)

The MCP specification requires that servers use **structured logging only** to stderr. Direct output to stdout/stderr via `fmt.Printf`, `fmt.Println`, `print`, or `println` will break MCP protocol communication.

#### Required Logging Pattern:

**✅ CORRECT - Use structured logging:**
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

**❌ INCORRECT - Never use in internal/ packages:**
```go
// NEVER use these in internal/server, internal/tools, internal/sdk, or internal/auth:
fmt.Printf("Operation completed\n")        // Breaks MCP protocol
fmt.Println("Processing...")                // Breaks MCP protocol
print("Debug info")                         // Breaks MCP protocol
println("Error occurred")                   // Breaks MCP protocol
```

**Note:** The only exception is `cmd/` packages which provide CLI commands for users (not MCP server code).

#### Enforcement:

This is enforced by:
1. **golangci-lint** with `forbidigo` linter - checks at lint time
2. **Code review** - reviewers will reject PRs that violate this
3. **MCP specification compliance** - violations break protocol communication

### 2. Error Handling Standards

Follow the project's error handling patterns for consistent error reporting:

**For Tool Errors:**
```go
import "github.com/pingidentity/pingone-mcp-server/internal/errs"

// When a tool encounters an error
return errs.NewToolError("tool_name", err, "Additional context about what failed")
```

**For Command Errors:**
```go
import "github.com/pingidentity/pingone-mcp-server/internal/errs"

// When a CLI command encounters an error
return errs.NewCommandError("command_name", err)
```

**For API Errors:**
```go
import "github.com/pingidentity/pingone-mcp-server/internal/errs"

// When processing API responses
apiErr := errs.HandlePingOneAPIError(httpResponse, responseBody, err)
if apiErr != nil {
    return errs.NewToolError("tool_name", apiErr, "Failed to call PingOne API")
}
```

### 3. Context Management

Always pass context through the call chain for:
- Logging (logger is attached to context)
- Cancellation signals
- Transaction tracking
- Request lifecycle management

**Pattern:**
```go
func ProcessRequest(ctx context.Context, input string) error {
    log := logger.FromContext(ctx)  // Get logger from context
    
    // Pass context to downstream calls
    result, err := someOperation(ctx, input)
    if err != nil {
        log.Error("Operation failed", "error", err)
        return err
    }
    
    return nil
}
```

### 4. Tool Implementation Standards

When implementing MCP tools, follow these patterns:

**Tool Structure:**
```go
type MyTool struct {
    clientFactory sdk.ClientFactory
    // ... other dependencies
}

func (t *MyTool) Execute(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResponse, error) {
    log := logger.FromContext(ctx)
    log.Debug("Tool execution started")
    
    // 1. Parse and validate input arguments
    var args MyToolArgs
    if err := parseToolArguments(req, &args); err != nil {
        return nil, errs.NewToolError("my_tool", err, "Failed to parse arguments")
    }
    
    // 2. Perform operations
    result, err := t.performOperation(ctx, args)
    if err != nil {
        return nil, errs.NewToolError("my_tool", err, "Operation failed")
    }
    
    // 3. Return structured response
    return newToolResponse(result)
}
```

**Tool Registration:**
```go
func RegisterMyTools(ctx context.Context, server *mcp.Server, factory sdk.ClientFactory) error {
    log := logger.FromContext(ctx)
    
    tool := NewMyTool(factory)
    
    if err := server.AddTool(tool.Definition(), tool.Execute); err != nil {
        log.Error("Failed to register tool", "tool", "my_tool", "error", err)
        return err
    }
    
    log.Debug("Tool registered", "tool", "my_tool")
    return nil
}
```

### 5. Testing Standards

**Unit Tests:**
- Test all public functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies (SDK clients, token stores, etc.)
- Use `internal/testutils` for common test assertions

**Integration Tests:**
- Test tool execution end-to-end
- Use `testutils.NewTestServer()` for MCP server testing
- Verify proper error handling and logging
- Test both success and failure scenarios

**Example:**
```go
func TestMyTool_Execute(t *testing.T) {
    tests := []struct {
        name    string
        input   map[string]interface{}
        want    interface{}
        wantErr bool
    }{
        {
            name: "valid input",
            input: map[string]interface{}{"key": "value"},
            want: expectedResult,
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Code Organization

- **cmd/**: CLI commands for user interaction (fmt.Printf allowed here)
- **internal/server**: MCP server implementation (structured logging only)
- **internal/tools**: MCP tool implementations (structured logging only)
- **internal/sdk**: PingOne SDK client wrappers (structured logging only)
- **internal/auth**: Authentication and session management (structured logging only)
- **internal/logger**: Logging infrastructure (can use fmt for implementation)
- **internal/errs**: Error handling utilities (can use fmt for formatting)
- **internal/testutils**: Test utilities and helpers

## Git Commit Standards

Follow these commit message guidelines:
- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Reference issue numbers when applicable
- Keep first line under 72 characters

## Pull Request Checklist

Before submitting a PR, ensure:
- [ ] Code follows MCP specification (especially logging requirements)
- [ ] `make test` passes
- [ ] `make lint` passes (includes forbidigo checks)
- [ ] New tools have proper documentation
- [ ] Error handling follows project patterns
- [ ] All code uses structured logging (no fmt.Printf in internal/)
- [ ] Tests cover new functionality
- [ ] PR template is filled out completely

## Additional Notes

- **Copyright Headers**: All Go files must include the copyright header:
  ```go
  // Copyright © 2025 Ping Identity Corporation
  ```

- **Dependencies**: Prefer standard library when possible. New dependencies require approval.

- **Documentation**: Update relevant docs when adding new tools or features:
  - Update README.md for user-facing changes
  - Update docs/troubleshooting.md for common issues
  - Add inline godoc comments for exported functions

## Legitimate Stdout Output - When and How

### Where fmt.Printf IS Allowed

The linting rules **automatically exempt** these locations:

1. **`cmd/` directory** - CLI commands for user interaction:
   ```go
   // cmd/session/session.go - ✅ ALLOWED
   fmt.Println("Current Session Information:")
   fmt.Printf("  Session ID: %s\n", sessionID)
   ```

2. **Test files** (`*_test.go`):
   ```go
   // any_test.go - ✅ ALLOWED
   fmt.Printf("Debug test output: %v\n", testData)
   ```

3. **`internal/errs/` directory** - Error formatting:
   ```go
   // internal/errs/error.go - ✅ ALLOWED
   return fmt.Sprintf("error: %s", msg)
   ```

4. **`internal/logger/` directory** - Logger implementation:
   ```go
   // internal/logger/logger.go - ✅ ALLOWED
   fmt.Fprintf(os.Stderr, "[LOG] %s\n", msg)  // Writing to stderr for bootstrap logging
   ```

### Where fmt.Printf IS FORBIDDEN

**MCP server runtime code** - These packages must use structured logging only:
- `internal/server`
- `internal/tools`
- `internal/sdk`
- `internal/auth/client`
- `internal/auth/login`
- `internal/auth/logout`

### What If I Need Output in Server Code?

You should **never need direct stdout** in MCP server code. Use these alternatives:

| Need | ❌ Don't Use | ✅ Use Instead |
|------|-------------|----------------|
| Debug info | `fmt.Printf("Debug: %v", data)` | `log.Debug("Debug info", "data", data)` |
| User feedback | `fmt.Println("Processing...")` | Return through MCP response |
| Error messages | `fmt.Printf("Error: %v", err)` | `log.Error("Operation failed", "error", err)` |
| Progress updates | `fmt.Printf("50%% complete")` | Use MCP progress notifications |
| Tool results | `fmt.Printf("%v", result)` | Return via `mcp.CallToolResponse` |

### Temporarily Disabling Checks (Development Only)

If you need to temporarily disable checks during development:

```go
// TEMPORARY: Remove before committing
// nolint:forbidigo
fmt.Printf("Debug output: %v\n", data)
```

**Important**: These should NEVER be committed. The PR checklist will catch them.

### Adding New Exemptions

If you need to exempt a new package, update `.golangci.yml`:

```yaml
issues:
  exclude-rules:
    # Allow in new package
    - path: ^internal/mynewpackage/
      linters:
        - forbidigo
```

However, **be very careful** - most new packages should follow MCP logging patterns.

## Common Mistakes to Avoid

❌ **Using fmt.Printf in server/tool code** - Breaks MCP protocol  
❌ **Not passing context through call chains** - Breaks logging and cancellation  
❌ **Ignoring errors** - All errors must be handled or explicitly ignored with comment  
❌ **Missing structured logging** - Always use logger.FromContext(ctx)  
❌ **Hardcoding configuration** - Use environment variables or config files  
❌ **Not testing error paths** - Error handling must be tested

## Questions?

If you're unsure about MCP specification requirements or project patterns:
1. Check existing code in the same package for patterns
2. Review the MCP specification documentation
3. Ask in PR comments for clarification
4. Check `internal/logger` for logging examples
5. Check `internal/tools` for tool implementation examples
