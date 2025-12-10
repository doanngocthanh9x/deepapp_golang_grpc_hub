# 🚀 Deployment Guide - All-in-One Image

## Quick Start

### 1. Pull từ GitHub Container Registry

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin

# Pull image
docker pull ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest

# Run container
docker run -d \
  --name deepapp-hub \
  -p 8081:8081 \
  -p 50051:50051 \
  ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest
```

### 2. Sử dụng Docker Compose (Recommended)

```bash
# Set environment variables
export GITHUB_REPOSITORY=doanngocthanh9x/deepapp_golang_grpc_hub
export IMAGE_TAG=main  # or 'latest', 'develop', specific SHA

# Pull and start
docker-compose -f docker-compose.registry.yml pull
docker-compose -f docker-compose.registry.yml up -d

# Check logs
docker logs -f deepapp-hub-all-in-one

# Check status
docker-compose -f docker-compose.registry.yml ps
```

## 📦 Available Services

All-in-one image bao gồm:

- **Hub**: gRPC Hub (port 50051)
- **WebAPI**: REST API Gateway (port 8081)
- **Python Worker**: 4 capabilities
- **Node.js Worker**: 5 capabilities
- **Go Worker**: 3 capabilities
- **Java Worker**: 2 capabilities

**Total: 14 capabilities** across 4 languages!

## 🌐 Access Points

```bash
# WebAPI Home
http://localhost:8081

# Health Check
http://localhost:8081/health

# Swagger UI (Interactive API Docs)
http://localhost:8081/api/docs

# Get All Capabilities
http://localhost:8081/api/capabilities

# Get Registered Workers
http://localhost:8081/workers
```

## 🧪 Test APIs

### Test Python Worker - Calculator

```bash
curl -X POST http://localhost:8081/api/python-worker/call/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation": "add", "a": 10, "b": 5}'
```

### Test Go Worker - Hello

```bash
curl -X POST http://localhost:8081/api/go-worker/call/hello_go \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Test Node.js Worker - String Operations

```bash
curl -X POST http://localhost:8081/api/node-worker/call/string_ops \
  -H "Content-Type: application/json" \
  -d '{"text": "hello", "operation": "uppercase"}'
```

### Test Java Worker - File Info

```bash
curl -X POST http://localhost:8081/api/java-simple-worker/call/read_file_info \
  -H "Content-Type: application/json" \
  -d '{"filePath": "/etc/hosts"}'
```

## 🔄 CI/CD Workflow

GitHub Actions tự động build và push image khi:

- Push to `main` or `develop` branch
- Any changes in `services/`, `shared/`, `proto/`, `Dockerfile.all-in-one`

### Available Tags

```bash
# Latest stable (from main branch)
ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest

# Branch-based
ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:main
ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:develop

# Commit-based (specific version)
ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:main-sha1234567

# PR preview
ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:pr-123
```

## 🛠️ Management Commands

```bash
# Start
docker-compose -f docker-compose.registry.yml up -d

# Stop
docker-compose -f docker-compose.registry.yml down

# Restart
docker-compose -f docker-compose.registry.yml restart

# View logs
docker-compose -f docker-compose.registry.yml logs -f

# Update to latest
docker-compose -f docker-compose.registry.yml pull
docker-compose -f docker-compose.registry.yml up -d

# Execute command in container
docker exec -it deepapp-hub-all-in-one bash

# Check running processes
docker exec deepapp-hub-all-in-one ps aux
```

## 📊 Resource Usage

- **Image Size**: ~1.2GB (Ubuntu-based)
- **Memory**: ~300MB (all 6 services running)
- **CPU**: Minimal at idle, scales with load
- **Startup Time**: ~10 seconds

## 🔐 Production Setup

### Environment Variables

```bash
# .env file
GITHUB_REPOSITORY=doanngocthanh9x/deepapp_golang_grpc_hub
IMAGE_TAG=latest
HUB_ADDRESS=localhost:50051
PORT=8081
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker Compose Production

```yaml
version: '3.8'

services:
  deepapp-hub:
    image: ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest
    container_name: deepapp-hub-all-in-one
    ports:
      - "127.0.0.1:8081:8081"  # Only localhost
      - "127.0.0.1:50051:50051"
    restart: always
    environment:
      - HUB_ADDRESS=localhost:50051
      - PORT=8081
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8081/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 🐛 Troubleshooting

### Container not starting

```bash
# Check logs
docker logs deepapp-hub-all-in-one

# Check if ports are available
netstat -tuln | grep -E '8081|50051'

# Remove and recreate
docker-compose -f docker-compose.registry.yml down -v
docker-compose -f docker-compose.registry.yml up -d
```

### Workers not registering

```bash
# Check supervisor status inside container
docker exec deepapp-hub-all-in-one ps aux

# Check individual worker logs
docker exec deepapp-hub-all-in-one cat /var/log/supervisor/python-worker.out.log
docker exec deepapp-hub-all-in-one cat /var/log/supervisor/go-worker.out.log
```

### Cannot pull image (Permission denied)

```bash
# Make sure you're logged in
docker login ghcr.io -u YOUR_USERNAME

# Check image exists
docker pull ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest
```

## 📈 Monitoring

### Health Checks

```bash
# Simple health check
curl http://localhost:8081/health

# Get all workers status
curl http://localhost:8081/api/capabilities | jq '.workers'

# Check specific worker
curl http://localhost:8081/workers | jq '.[] | select(.id=="python-worker")'
```

### Prometheus Metrics (Future)

```bash
# Metrics endpoint (if enabled)
curl http://localhost:8081/metrics
```

## 🔄 Update Strategy

### Zero-Downtime Update

```bash
# Pull new image
docker pull ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest

# Start new container with different name
docker run -d \
  --name deepapp-hub-new \
  -p 8082:8081 \
  -p 50052:50051 \
  ghcr.io/doanngocthanh9x/deepapp_golang_grpc_hub/all-in-one:latest

# Test new container
curl http://localhost:8082/health

# Switch traffic (update nginx/load balancer)
# Then remove old container
docker stop deepapp-hub-all-in-one
docker rm deepapp-hub-all-in-one

# Rename new container
docker rename deepapp-hub-new deepapp-hub-all-in-one
```

---

**Last Updated**: December 10, 2025  
**Maintained By**: DeepApp Team
