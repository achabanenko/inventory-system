# Build stage
FROM oven/bun:1.1-alpine AS builder

WORKDIR /app

# Copy package files for better layer caching
COPY package.json bun.lockb* ./

# Install dependencies with Bun (much faster than npm)
RUN bun install --frozen-lockfile --no-save

# Copy source code
COPY . .

# Build the application with optimizations
RUN bun run build

# Create a minimal static server
FROM golang:1.24-alpine AS server-builder
WORKDIR /app
COPY <<EOF main.go
package main

import (
    "log"
    "net/http"
    "os"
    "path/filepath"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }
    
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        path := filepath.Join("./static", r.URL.Path)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            http.ServeFile(w, r, "./static/index.html")
            return
        }
        fs.ServeHTTP(w, r)
    }))
    
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
EOF

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o server main.go

# Runtime stage - minimal with distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the static server
COPY --from=server-builder /app/server /app/server

# Copy built files
COPY --from=builder /app/dist ./static

# Expose port
EXPOSE 3000

# Use nonroot user for security
USER nonroot:nonroot

# Start static file server (compatible with Traefik)
ENTRYPOINT ["/app/server"]