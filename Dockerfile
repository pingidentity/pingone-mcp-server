FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
# Download dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY . .
# Build static Linux binary
WORKDIR /src/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# Final image
FROM gcr.io/distroless/static:nonroot
WORKDIR /

# Copy binary from builder
COPY --from=builder /app/server /server

# Set default environment variables (can be overridden at runtime)
ENV PINGONE_MCP_TRANSPORT=stdio
ENV PINGONE_MCP_DEBUG_API=false
ENV PINGONE_MCP_ALLOW_MUTATION=false
ENV PINGONE_MCP_ALLOW_INSECURE=false
ENV PINGONE_MCP_SERVER_PORT=8080
ENV PINGONE_MCP_API_KEY_PATH=/tmp/pingone-mcp-server-api.key
ENV PINGONE_REGION=com

# Use non-root user
USER nonroot:nonroot

# Run the server (accepts CLI arguments)
ENTRYPOINT ["/server"]