# Dockerfile for Go Worker
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Copy all source code
COPY . .

# Tidy and download modules
RUN go mod tidy && go mod download

# Build go-worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o go-worker ./services/go-worker/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/go-worker /usr/local/bin/go-worker

ENTRYPOINT ["/usr/local/bin/go-worker"]
