# Build frontend assets
FROM node:24-alpine AS frontend-builder

WORKDIR /app/frontend

RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./
RUN pnpm build

# Multi-stage build for Go application
FROM golang:alpine AS builder

# Install git and ca-certificates
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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o web-clipboard-go ./backend/cmd/web-clipboard

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/web-clipboard-go .

# Copy built frontend files
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Change ownership to non-root user
RUN mkdir -p /data && chown -R appuser:appuser /app /data

VOLUME ["/data"]

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 5000

# Health check (using wget which is available in alpine)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5000/ || exit 1

# Run the application
CMD ["./web-clipboard-go"]
