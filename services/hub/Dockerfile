# Dockerfile for Hub service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared code
COPY shared/ ./shared/
COPY internal/ ./internal/
COPY proto/ ./proto/

# Copy hub service
COPY services/hub/ ./services/hub/

# Build hub
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o hub ./services/hub/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/hub /usr/local/bin/hub

EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/hub"]
