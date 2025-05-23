# Build stage
FROM golang:1.21.6-alpine3.19 AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for private repos and HTTPS)
RUN apk update && apk upgrade && apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o k8s-diagnostics-mcp .

# Final stage
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/k8s-diagnostics-mcp /k8s-diagnostics-mcp

# Expose port (if needed for your MCP setup)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/k8s-diagnostics-mcp"]