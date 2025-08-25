############################
# Build stage
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main cmd/api/main.go

############################
# Runtime stage (Alpine)
FROM alpine:3.19

# Add non-root user and certs
RUN addgroup -S app && adduser -S app -G app && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

# Copy the binary and certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/main /usr/local/bin/main

# Expose port
EXPOSE 8080

# Drop privileges
USER app

# Run the application
ENTRYPOINT ["/usr/local/bin/main"]