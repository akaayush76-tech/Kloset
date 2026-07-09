# Build stage — use native build platform so cross-compilation works on ARM Macs
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Cross-compile for the target platform (linux/amd64 on Fly.io)
# CGO_ENABLED=0 means pure Go — no C toolchain needed for cross-compilation
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -o bin/server ./cmd/server

# Final stage
FROM --platform=linux/amd64 alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/bin/server .

# Copy runtime config (recommendation color matrix — tunable without rebuild)
COPY --from=builder /app/config ./config

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Change ownership of app directory
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/api/products || exit 1

# Run the application
CMD ["./server"]
