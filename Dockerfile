# ===================================
# Stage 1: Builder
# ===================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations
# -ldflags="-s -w" strips debug info and reduces binary size by ~30%
# CGO_ENABLED=0 creates a static binary (no dependencies)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -a -installsuffix cgo \
    -o omiro .

# ===================================
# Stage 2: Runtime (Distroless)
# ===================================
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data and CA certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary and static files
COPY --from=builder /build/omiro /app/omiro
COPY --from=builder /build/index.html /app/index.html

WORKDIR /app

# Use non-root user for security
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/omiro"]