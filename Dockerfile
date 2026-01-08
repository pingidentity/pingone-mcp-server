FROM golang:1.25.1-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pingone-mcp-server .

# Use a minimal base image for the final stage
FROM alpine:latest

WORKDIR /root/

# Environment variables for PingOne MCP Server configuration
ENV PINGONE_TOP_LEVEL_DOMAIN="" \
    PINGONE_REGION_CODE="" \
    PINGONE_MCP_ENVIRONMENT_ID="" \
    PINGONE_DEVICE_CODE_CLIENT_ID="" \
    PINGONE_DEVICE_CODE_SCOPES="openid" \
    PINGONE_MCP_DEBUG="false"

# Copy the binary from the builder stage
COPY --from=builder /app/pingone-mcp-server .

# Copy the entrypoint script
COPY docker-entrypoint.sh .

RUN chmod +x ./pingone-mcp-server && \
    chmod +x ./docker-entrypoint.sh

ENTRYPOINT ["./docker-entrypoint.sh"]