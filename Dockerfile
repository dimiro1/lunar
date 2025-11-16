# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-w -s' -o faas-go ./cmd

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/faas-go .

# Expose port
EXPOSE 3000

# Environment variables (Railway will override PORT)
ENV DATA_DIR=/data \
    EXECUTION_TIMEOUT=300

# Run the application
CMD ["./faas-go"]
