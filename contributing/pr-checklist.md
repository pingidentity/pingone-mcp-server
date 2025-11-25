# Pull Request Checklist

The following provides the steps to check/run to prepare for creating a PR to the `main` branch. PRs that follow these checklists will merge faster than PRs that do not.

*Note: This checklist is designed to support both human contributors and automated code review tools.*

## For Automated Code Review

This checklist includes specific verification criteria marked with *Verification* that can be programmatically checked to support both manual and automated review processes.

## PR Planning & Structure

- [ ] **PR Scope**. To ensure maintainer reviews are as quick and efficient as possible, please separate functionality into logical PRs. For example, support for a new MCP tool and related utilities can go in the same PR, however support for multiple unrelated features should be separated. It's acceptable to merge related changes into the same PR where structural improvements are being made.
  - *Verification*: Check that files modified are logically related (same package directory, related functionality)

- [ ] **PR Title**. To assist the maintainers in assessing PRs for priority, please provide a descriptive title of the functionality being supported. For example: `Add MCP tool for user management` or `Fix OAuth2 token refresh in session manager`
  - *Verification*: Title should be descriptive and match the type of changes (Add/Update/Fix/Remove)

- [ ] **PR Description**. Please follow the provided PR description template and check relevant boxes. Include a clear description of:
  - What functionality is being added/changed
  - Why the change is needed (e.g., to fix an issue - include the issue number as reference)
  - Any breaking changes
  - *Verification*: Check that PR description template boxes are completed and description sections are filled

## Code Development

### Architecture & Design

- [ ] **Code implementation**. New code should follow established patterns in the codebase:
  - *Verification*: 
    - New packages follow the established server architecture (`cmd/`, `internal/`)
    - Code follows Go best practices and package organization standards
    - MCP tools are implemented in appropriate packages
    - Changes follow consistent patterns with existing code

- [ ] **MCP Tool Implementation**. If adding or modifying MCP tools:
  - *Verification*:
    - Tools are properly registered with the MCP server
    - Tool schemas are well-defined and documented
    - Tool handlers include proper error handling
    - Tool implementations follow MCP protocol standards

### Code Quality

- [ ] **Dependencies Check**. Ensure go.mod and go.sum are properly maintained:

```shell
go mod tidy
```
*Verification*: Run command and verify no changes to go.mod/go.sum files

- [ ] **Build**. Verify the server builds successfully with your changes:

```shell
make build
```
*Verification*: Run command and verify exit code 0

- [ ] **Code Linting**. Run all linting checks to ensure code quality and consistency:

```shell
make lint
```
*Verification*: Command must exit with code 0

This includes:
- Go vet checks
- golangci-lint
- Import organization checks
- Go code formatting

## Testing

### Unit Tests

- [ ] **Unit Tests**. Where a code function performs work internally to a package, but has an external scope (i.e., a function with an initial capital letter `func MyFunction`), unit tests should ideally be created. Not all functions require a unit test, if in doubt please ask:

```shell
make test
```
*Verification*: Run command and verify exit code 0

### Integration Tests

- [ ] **Manual Testing**. New MCP tools or significant changes should be manually tested with an MCP client:
  - *Verification*:
    - Document the MCP client used for testing (e.g., Claude Desktop, MCP Inspector)
    - Describe the test scenario and expected behavior
    - Include screenshots or logs if helpful
    - Verify the tool works end-to-end with actual PingOne services

- [ ] **Test Environment**. Ensure you have access to a PingOne environment for testing:
  - *Verification*:
    - Test credentials are configured properly
    - No sensitive data (credentials, tokens) is committed to the repository
    - Tests work against a real PingOne environment

## Documentation

### Code Documentation

- [ ] **Package Documentation**. Each package should have proper package-level documentation following Go conventions:
  - *Verification*: 
    - All packages have package comments that describe their purpose
    - Public functions and types have appropriate doc comments
    - Doc comments follow Go documentation standards (start with function/type name)
    - Documentation is clear and includes usage examples where helpful

- [ ] **Function Documentation**. Public functions should have clear documentation describing their purpose, parameters, return values, and any important behavior:
  - *Verification*: 
    - All exported functions have doc comments
    - Comments describe what the function does, not how it does it
    - Parameter and return value descriptions are included where non-obvious
    - Error conditions are documented where applicable

- [ ] **MCP Tool Documentation**. MCP tools should be well-documented:
  - *Verification*:
    - Tool schemas include clear descriptions
    - Parameter requirements and types are documented
    - Expected return values are documented
    - Example usage is provided

### User Documentation

- [ ] **README Updates**. If adding new features or changing usage, update README.md:
  - *Verification*:
    - Installation instructions are current
    - Usage examples reflect new functionality
    - Command documentation is complete and accurate

- [ ] **CONTRIBUTING.md Updates**. If changing development workflows or adding new patterns:
  - *Verification*:
    - Development setup instructions are current
    - New patterns are documented
    - Examples are updated

## Security & Compliance

- [ ] **Sensitive Data**. Ensure no sensitive data (API keys, tokens, etc.) are committed to the repository:
  - *Verification*: 
    - No API keys, passwords, or tokens in code or test files
    - Sensitive test data uses environment variables
    - No `.env` files or similar containing credentials
    - Test examples use placeholder credentials or environment variable references

- [ ] **Input Validation**. Implement appropriate input validation for all user-provided data:
  - *Verification*: 
    - MCP tool inputs are validated
    - Error handling provides clear feedback for invalid inputs
    - No potential for injection attacks or unsafe operations

- [ ] **Credential Storage**. If changes affect authentication or session management:
  - *Verification*:
    - Credentials are stored securely using system keychain
    - Token lifecycle is managed properly
    - No credentials are logged or exposed in error messages

- [ ] **Dependency Security**. Ensure all dependencies are secure and up-to-date:
  - *Verification*:
    - `go mod tidy` produces no warnings about vulnerabilities
    - Dependencies are from trusted sources
    - Minimal dependency footprint is maintained

## Final Checks

- [ ] **All Development Checks**. Run the comprehensive development check:

```shell
make test
make lint
make build
```
*Verification*: All commands exit with code 0

- [ ] **CI Compatibility**. Verify your changes will pass automated CI checks by ensuring all the above steps pass locally:
  - *Verification*: All previous verification steps completed successfully

- [ ] **Breaking Changes**. If your PR introduces breaking changes, ensure they are:
  - Clearly documented in the PR description
  - Follow the project's versioning strategy
  - Include migration guidance for users
  - *Verification*: 
    - Breaking changes are documented in PR description
    - Migration guidance is provided
    - Backward compatibility considerations are addressed where possible

- [ ] **Version Information**. If this PR should trigger a version bump:
  - Document the type of version change (major/minor/patch)
  - Update any version-related documentation
  - Follow semantic versioning principles

## Additional Notes

- The maintainers may run additional tests in different PingOne regions
- Large PRs may take longer to review - consider breaking them into smaller, focused changes
- If you're unsure about any step, please ask questions in your PR or create an issue for discussion
- MCP tool implementations should be tested with real MCP clients when possible

---

## Documentation-Only Changes

If you are making documentation-only changes (guides, examples, or README updates), you can use this simplified checklist:

- [ ] **Documentation Updates**. New or updated documentation should be clear, well-structured, and include practical examples

- [ ] **Formatting**. Ensure documentation follows proper formatting and style guidelines

- [ ] **Accuracy**. Verify all commands, examples, and instructions are accurate and up-to-date

Documentation changes are generally merged quicker than code changes as there is less to review.

---

## Quick Reference

### Common Commands

```shell
# Build the server
make build

# Run tests
make test

# Run linter
make lint

# Test login flow
./bin/pingone-mcp-server login

# Test server run
./bin/pingone-mcp-server run

# Check version
./bin/pingone-mcp-server --version
```

### Testing with MCP Clients

**Claude Desktop:**
1. Update `~/Library/Application Support/Claude/claude_desktop_config.json`
2. Add server configuration
3. Restart Claude Desktop
4. Test MCP tools through chat interface

**MCP Inspector:**
1. Clone and run the inspector tool
2. Connect to your local MCP server
3. Test tools interactively
4. Inspect request/response payloads
