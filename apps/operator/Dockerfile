# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o operator ./main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S streamforge && \
    adduser -u 1001 -S streamforge -G streamforge

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/operator .

# Copy config files
COPY --from=builder /app/config ./config

# Change ownership to non-root user
RUN chown -R streamforge:streamforge /app

# Switch to non-root user
USER streamforge

# Expose ports
EXPOSE 8080 8081 9443

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/healthz || exit 1

# Run the application
ENTRYPOINT ["./operator"] 