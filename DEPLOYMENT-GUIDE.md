# Deployment Guide - GitHub Container Registry

## Prerequisites

1. **GitHub Personal Access Token (PAT)** with `read:packages` permission
2. **Docker** installed on deployment server

## Setup

### 1. Login to GitHub Container Registry

```bash
# Set environment variables
export GITHUB_USERNAME="your-github-username"
export GITHUB_TOKEN="your-github-pat"
export GITHUB_REPOSITORY="doanngocthanh9x/deepapp_golang_grpc_hub"

# Login to registry
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin
```

### 2. Choose Deployment Strategy

## Option A: All-in-One Deployment (Recommended for Simple Setup)

**Single container with all services**

```bash
# Set image tag (main, develop, or specific SHA)
export IMAGE_TAG=main

# Pull and run
docker-compose -f docker-compose.registry-all-in-one.yml pull
docker-compose -f docker-compose.registry-all-in-one.yml up -d

# Check logs
docker logs deepapp-hub-all-in-one

# Check services inside container
docker exec deepapp-hub-all-in-one ps aux
```

**Services included:**
- ✅ Hub (gRPC core)
- ✅ WebAPI (REST API)
- ✅ Python Worker (5 capabilities)
- ✅ Java Worker (2 capabilities)
- ✅ Node Worker (5 capabilities)
- ✅ Go Worker (3 capabilities)

**Ports:**
- `50051` - gRPC Hub
- `8081` - WebAPI

**Pros:**
- Simple deployment (1 container)
- Lower memory footprint
- Easier to manage
- Good for development/testing

**Cons:**
- Cannot scale individual services
- All services restart together
- Larger image size (~900MB)

## Option B: Microservices Deployment (Recommended for Production)

**Separate containers for each service**

```bash
# Set image tag
export IMAGE_TAG=main

# Pull all images
docker-compose -f docker-compose.registry.yml pull

# Start all services
docker-compose -f docker-compose.registry.yml up -d

# Check status
docker-compose -f docker-compose.registry.yml ps

# View logs
docker-compose -f docker-compose.registry.yml logs -f
```

**Services:**
- Hub (core)
- WebAPI
- Python Worker
- Java Worker  
- Node Worker
- Go Worker
- C++ Worker (separate build)

**Pros:**
- Independent scaling
- Service isolation
- Faster updates (rebuild only changed services)
- Better for production

**Cons:**
- More containers to manage
- Slightly higher memory usage
- More complex networking

## Available Image Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `main` | Latest stable from main branch | `ghcr.io/.../all-in-one:main` |
| `develop` | Development branch | `ghcr.io/.../all-in-one:develop` |
| `main-abc123` | Specific commit SHA | `ghcr.io/.../all-in-one:main-abc123` |
| `latest` | Latest from default branch | `ghcr.io/.../all-in-one:latest` |

## Testing Deployment

```bash
# Check WebAPI health
curl http://localhost:8081/health

# List available capabilities
curl http://localhost:8081/api/capabilities | jq

# Test a capability
curl -X POST http://localhost:8081/api/execute \
  -H "Content-Type: application/json" \
  -d '{
    "capability": "hello_python",
    "params": {"name": "World"}
  }'
```

## Updating Deployment

### All-in-One

```bash
# Pull latest image
docker-compose -f docker-compose.registry-all-in-one.yml pull

# Restart with new image
docker-compose -f docker-compose.registry-all-in-one.yml up -d

# Check logs for successful startup
docker logs deepapp-hub-all-in-one --tail 50
```

### Microservices

```bash
# Update specific service
docker-compose -f docker-compose.registry.yml pull python-worker
docker-compose -f docker-compose.registry.yml up -d python-worker

# Or update all
docker-compose -f docker-compose.registry.yml pull
docker-compose -f docker-compose.registry.yml up -d
```

## Monitoring

```bash
# All-in-One: Check service status inside container
docker exec deepapp-hub-all-in-one ps aux

# All-in-One: View supervisor logs
docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/*.log

# Microservices: Check all container status
docker-compose -f docker-compose.registry.yml ps

# Microservices: View specific service logs
docker-compose -f docker-compose.registry.yml logs -f python-worker
```

## Troubleshooting

### Image Pull Issues

```bash
# Verify login
docker login ghcr.io

# Check image exists
docker manifest inspect ghcr.io/${GITHUB_REPOSITORY}/all-in-one:main

# Force pull
docker-compose -f docker-compose.registry-all-in-one.yml pull --ignore-pull-failures
```

### Service Not Starting

```bash
# All-in-One: Check supervisor status
docker exec deepapp-hub-all-in-one supervisorctl status

# All-in-One: View error logs
docker exec deepapp-hub-all-in-one cat /var/log/supervisor/python-worker.err.log

# Microservices: Check container logs
docker logs python-worker
```

### Performance Issues

```bash
# Check resource usage
docker stats

# All-in-One: Check processes
docker exec deepapp-hub-all-in-one top

# Restart specific service (microservices)
docker-compose -f docker-compose.registry.yml restart python-worker
```

## CI/CD Workflow

### Automatic Builds

GitHub Actions automatically builds and pushes images on:
- ✅ Push to `main` or `develop` branches
- ✅ Pull requests to `main`
- ✅ Manual trigger via workflow dispatch

### Smart Caching

- Only changed services are rebuilt
- Layer caching via GitHub Container Registry
- Proto changes trigger all service rebuilds

### Build Times

| Scenario | Time |
|----------|------|
| First build (no cache) | ~8-10 minutes |
| Cached (no changes) | ~30-60 seconds |
| Single service change | ~1-3 minutes |
| Proto change (all services) | ~6-8 minutes |

## Security

```bash
# Use specific tags in production
export IMAGE_TAG=main-abc123def

# Scan images for vulnerabilities
docker scout cves ghcr.io/${GITHUB_REPOSITORY}/all-in-one:main

# Keep images updated
docker-compose -f docker-compose.registry-all-in-one.yml pull
```

## Backup and Data

```bash
# All-in-One: Backup data volume
docker cp deepapp-hub-all-in-one:/data ./backup/

# Restore data
docker cp ./backup/ deepapp-hub-all-in-one:/data/
```

## Production Recommendations

### All-in-One

✅ **Use when:**
- Small to medium workload
- Simple deployment requirements
- Limited infrastructure
- Development/staging environments

### Microservices

✅ **Use when:**
- Need to scale specific services
- High availability requirements
- Complex orchestration (Kubernetes)
- Production environments

## Support

- **Documentation**: `docs/` folder
- **Issues**: GitHub Issues
- **CI/CD Guide**: `docs/CI-CD-GUIDE.md`
