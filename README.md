# PingOne MCP Server

A Model Context Protocol (MCP) server implementation for PingOne.

## Overview

This server provides MCP tools for working with PingOne, built using the official Go SDK for MCP.

## Install

### From GitHub release

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for artifact downloads, artifact signatures, and the checksum file. To verify package downloads, see the [Verify Section](#verify).

OR

Use the following single-line command to install the server into '/usr/local/bin' directly.

```shell
RELEASE_VERSION=$(basename $(curl -Ls -o /dev/null -w %{url_effective} https://github.com/pingidentity/pingone-mcp-server/releases/latest)); \
OS_NAME=$(uname -s); \
HARDWARE_PLATFORM=$(uname -m | sed s/aarch64/arm64/ | sed s/x86_64/amd64/); \
URL="https://github.com/pingidentity/pingone-mcp-server/releases/download/${RELEASE_VERSION}/pingone-mcp-server_${RELEASE_VERSION#v}_${OS_NAME}_${HARDWARE_PLATFORM}"; \
curl -Ls -o pingone-mcp-server "${URL}"; \
chmod +x pingone-mcp-server; \
sudo mv pingone-mcp-server /usr/local/bin/pingone-mcp-server;
```

### Verify

#### Checksums

See [the latest GitHub release](https://github.com/pingidentity/pingone-mcp-server/releases/latest) for the checksums.txt file. The checksums are in the format of SHA256.

## Local usage

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
