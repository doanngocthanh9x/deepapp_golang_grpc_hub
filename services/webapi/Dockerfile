# Dockerfile for Web API service
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared code
COPY shared/ ./shared/
COPY internal/ ./internal/
COPY proto/ ./proto/

# Copy webapi service
COPY services/webapi/ ./services/webapi/

# Build webapi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o webapi ./services/webapi/main.go

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/webapi /usr/local/bin/webapi

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/webapi"]
