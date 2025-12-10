# Dockerfile for Go Worker
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared code (including SDK)
COPY shared/ ./shared/
COPY internal/ ./internal/
COPY proto/ ./proto/

# Copy go-worker service
COPY services/go-worker/ ./services/go-worker/

# Build go-worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o go-worker ./services/go-worker/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/go-worker /usr/local/bin/go-worker

ENTRYPOINT ["/usr/local/bin/go-worker"]
