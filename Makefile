.PHONY: build docker-build test lint

default: build

build:
	go mod tidy
	@mkdir -p bin
	go build -o bin/pingone-mcp-server .

docker-build:
	docker build -t pingone-mcp-server:dev .

test:
	go test -v -timeout 1m ./...

lint:
	go tool golangci-lint run --timeout 2m
