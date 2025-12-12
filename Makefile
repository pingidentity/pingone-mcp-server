.PHONY: build test lint mcp-spec-check validate-all

default: build

build:
	go mod tidy
	@mkdir -p bin
	go build -o bin/pingone-mcp-server .

test:
	go test -v -timeout 1m ./...

lint:
	go tool golangci-lint run --timeout 2m --config .golangci.yml

# MCP Specification compliance check - ensures no fmt.Printf/Println in server code
mcp-spec-check:
	@echo "Checking MCP specification compliance (no fmt.Printf/Println in internal/server, internal/tools, internal/sdk, internal/auth)..."
	@! grep -rn --include="*.go" --exclude="*_test.go" -E 'fmt\.(Printf|Println|Print|Fprint|Fprintf|Fprintln)' \
		internal/server internal/tools internal/sdk internal/auth/client internal/auth/login internal/auth/logout 2>/dev/null || \
		(echo "ERROR: Found fmt.Printf/Println in MCP server code. Use structured logging instead (logger.FromContext(ctx))." && \
		 echo "See .github/copilot-instructions.md for proper logging patterns." && exit 1)
	@echo "✓ MCP specification compliance verified"

# Run all validation checks
validate-all: test lint mcp-spec-check
	@echo "✓ All validation checks passed"
