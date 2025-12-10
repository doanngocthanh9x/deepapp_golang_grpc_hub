# Dockerfile for Hub service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Copy all source to ensure go mod tidy works
COPY . .

# Tidy and download modules
RUN go mod tidy && go mod download

# Build hub from cmd/hub
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o hub ./cmd/hub/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/hub /usr/local/bin/hub

EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/hub"]
