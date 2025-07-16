# Build stage
FROM golang:1.23-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o kuadrant-mcp-server .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S mcp && \
    adduser -u 1001 -S mcp -G mcp

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/kuadrant-mcp-server .

# Change ownership
RUN chown -R mcp:mcp /app

# Switch to non-root user
USER mcp

# The MCP server uses stdio for communication
ENTRYPOINT ["./kuadrant-mcp-server"]