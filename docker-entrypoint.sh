#!/bin/sh
set -e

exec ./pingone-mcp-server run --grant-type=device_code --store-type=file "$@"
