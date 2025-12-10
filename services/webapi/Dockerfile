# Dockerfile for Web API service
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

# Build webapi from services/web-api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o webapi ./services/web-api/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/webapi /usr/local/bin/webapi

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/webapi"]
