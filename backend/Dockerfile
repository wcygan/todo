# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

# Enable Go modules and caching
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Pre-copy go.mod and go.sum to leverage Docker layer cache
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the source
COPY backend/ .

# Build the server binary
RUN go build -o /out/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

# Copy configuration (optional, can be overridden with mounts)
COPY --chown=nonroot:nonroot backend/configs ./configs

# Copy database migrations
COPY --chown=nonroot:nonroot backend/internal/store/migrations ./internal/store/migrations

# Copy built binary
COPY --from=builder /out/server /app/server

# Expose default port (matches config; can be overridden)
EXPOSE 8080

USER nonroot

ENV TODO_ENV=production

# Pass through common envs used by the app
ENV PORT=8080

ENTRYPOINT ["/app/server"]
