.PHONY: build test lint

default: build

build:
	go mod tidy
	@mkdir -p bin
	go build -o bin/pingone-mcp-server .

test:
	go test -v -timeout 1m ./...

lint:
	go tool golangci-lint run --timeout 2m
