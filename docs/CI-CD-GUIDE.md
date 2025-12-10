# CI/CD & Container Registry Guide

## GitHub Actions Workflow

### Automatic Builds

The project uses GitHub Actions to automatically build and push Docker images when code changes:

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main`
- Manual workflow dispatch

**Smart Change Detection:**
The workflow only builds images for services that actually changed:

```yaml
services/hub/**          → builds hub image
services/go-worker/**    → builds go-worker image
services/node-worker/**  → builds node-worker image
proto/**                 → builds all images (proto changes affect all)
```

### Image Tags

Images are tagged with multiple tags for flexibility:

- `main` - latest from main branch
- `develop` - latest from develop branch
- `pr-123` - pull request builds
- `main-sha1234567` - specific commit SHA

### Registry Locations

All images are pushed to GitHub Container Registry (ghcr.io):

```
ghcr.io/<username>/deepapp_golang_grpc_hub/hub:main
ghcr.io/<username>/deepapp_golang_grpc_hub/webapi:main
ghcr.io/<username>/deepapp_golang_grpc_hub/go-worker:main
ghcr.io/<username>/deepapp_golang_grpc_hub/java-worker:main
ghcr.io/<username>/deepapp_golang_grpc_hub/node-worker:main
ghcr.io/<username>/deepapp_golang_grpc_hub/python-worker:main
ghcr.io/<username>/deepapp_golang_grpc_hub/cpp-worker:main
ghcr.io/<username>/deepapp_golang_grpc_hub/all-in-one:main
```

## Using Registry Images

### Pull Pre-built Images

```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull specific service
docker pull ghcr.io/<username>/deepapp_golang_grpc_hub/go-worker:main

# Or use docker-compose with registry images
export GITHUB_REPOSITORY=<username>/deepapp_golang_grpc_hub
export IMAGE_TAG=main
docker-compose -f docker-compose.registry.yml up -d
```

### Build Caching

Each service uses Docker layer caching to speed up builds:

- First build: ~3-5 minutes per service
- Subsequent builds (no changes): ~30 seconds (cache hit)
- Subsequent builds (code changes): ~1-2 minutes (partial cache)

**Cache locations:**
```
ghcr.io/<username>/deepapp_golang_grpc_hub/hub:buildcache
ghcr.io/<username>/deepapp_golang_grpc_hub/go-worker:buildcache
...
```

## Local Development

### Build Single Service

```bash
# Build specific service
docker build -f services/go-worker/Dockerfile -t go-worker:dev .

# Build with cache from registry
docker build \
  --cache-from ghcr.io/<username>/deepapp_golang_grpc_hub/go-worker:buildcache \
  -f services/go-worker/Dockerfile \
  -t go-worker:dev .
```

### Build All-in-One (Current Method)

```bash
# Build the all-in-one image (includes all services)
docker-compose -f docker-compose.all-in-one.yml build

# Run
docker-compose -f docker-compose.all-in-one.yml up -d
```

## Deployment Strategies

### Strategy 1: All-in-One Container
**Pros:** Simple, single container
**Cons:** Rebuild entire image for any change
**Use case:** Development, small deployments

```bash
docker run -p 8081:8081 -p 50051:50051 \
  ghcr.io/<username>/deepapp_golang_grpc_hub/all-in-one:main
```

### Strategy 2: Separate Service Containers
**Pros:** Only rebuild changed services, better scalability
**Cons:** More containers to manage
**Use case:** Production, Kubernetes

```bash
export GITHUB_REPOSITORY=<username>/deepapp_golang_grpc_hub
docker-compose -f docker-compose.registry.yml up -d
```

### Strategy 3: Kubernetes (Future)
Deploy each service as separate pods with horizontal scaling.

## Performance Comparison

### Build Time Comparison

| Scenario | All-in-One | Separate Images |
|----------|-----------|-----------------|
| First build | 8-10 min | 8-10 min total |
| Go worker change | 8-10 min | 1-2 min |
| Node worker change | 8-10 min | 30-60 sec |
| Proto change | 8-10 min | 8-10 min total |

### CI/CD Workflow Example

When you change only `services/go-worker/main.go`:

1. GitHub Actions detects change in go-worker
2. Skips building hub, webapi, java-worker, node-worker, python-worker, cpp-worker
3. Only builds go-worker image (~1-2 min with cache)
4. Pushes go-worker:main to registry
5. Other services continue using cached images

## Monitoring Builds

View build status:
- GitHub Actions tab in repository
- Container packages: https://github.com/USERNAME?tab=packages

Check image sizes:
```bash
# List images with sizes
docker images | grep deepapp_golang_grpc_hub

# Typical sizes:
# hub:          ~20MB (Go binary)
# go-worker:    ~15MB (Go binary)
# node-worker:  ~150MB (Node + deps)
# java-worker:  ~200MB (JRE + JAR)
# python-worker:~80MB (Python + deps)
# cpp-worker:   ~100MB (Ubuntu + libs)
# all-in-one:   ~800MB (everything)
```

## Troubleshooting

### Build Cache Issues

If builds are slow or caching isn't working:

```bash
# Clear local build cache
docker builder prune -a

# Re-run workflow to rebuild caches
# Go to Actions → Build and Push → Re-run jobs
```

### Authentication Issues

```bash
# Create GitHub Personal Access Token with packages:write permission
# Then login:
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

### Image Not Found

Make sure package is public:
1. Go to GitHub repository → Packages
2. Click on package
3. Package settings → Change visibility → Public

## Future Enhancements

- [ ] Add image vulnerability scanning (Trivy)
- [ ] Add multi-arch builds (AMD64, ARM64)
- [ ] Add deployment to Kubernetes cluster
- [ ] Add staging environment builds
- [ ] Add rollback mechanism
- [ ] Add performance benchmarks in CI
