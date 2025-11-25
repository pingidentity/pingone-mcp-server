# PingOne MCP Server

A Model Context Protocol (MCP) server implementation for PingOne.

## Overview

This server provides MCP tools for working with PingOne, built using the official Go SDK for MCP.

## Usage

### Building local binary

```bash
make build
```

### Building binaries with GoReleaser locally

To build snapshot binaries for multiple platforms using GoReleaser, which are placed in the `dist/` directory:

```bash
docker run --rm --privileged \
  -v $PWD:/go/src/github.com/user/repo \
  -w /go/src/github.com/user/repo \
  goreleaser/goreleaser release --snapshot --clean
```

### Running

To start the MCP server binary:

```bash
./bin/pingone-mcp-server
```

### Help

To see all available commands:

```bash
./bin/pingone-mcp-server --help
```

## Docker

### Building the Docker image

```bash
docker build -t pingone-mcp-server .
```

### Running in Docker

The following environment variables are needed when running in docker:

```bash
PINGONE_TOP_LEVEL_DOMAIN=com
PINGONE_REGION_CODE=NA
PINGONE_MCP_ENVIRONMENT_ID=your-environment-id
PINGONE_DEVICE_CODE_CLIENT_ID=your-device-code-client-id
PINGONE_DEVICE_CODE_SCOPES=openid
# Optional - will add more debug console output
PINGONE_MCP_DEBUG=true
```

Run the container (the --disable-read-only argument is optional):

```bash
docker run -i --rm \
  --env-file ./mcp.env \
  pingone-mcp-server \
  --disable-read-only
```

Alternatively, pass environment variables directly:

```bash
docker run -i --rm \
  -e PINGONE_TOP_LEVEL_DOMAIN=com \
  -e PINGONE_REGION_CODE=NA \
  -e PINGONE_MCP_ENVIRONMENT_ID=your-environment-id \
  -e PINGONE_DEVICE_CODE_CLIENT_ID=your-device-code-client-id \
  -e PINGONE_DEVICE_CODE_SCOPES=openid \
  -e PINGONE_MCP_DEBUG=true \
  pingone-mcp-server \
  --disable-read-only
```

