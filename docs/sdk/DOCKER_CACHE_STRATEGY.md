# Docker Cache Prevention Strategy

## Problem
Docker layer caching can cause stale code to persist even after rebuilds, leading to bugs where code changes don't take effect.

## Solution
All Dockerfiles include cache-busting mechanisms to ensure fresh builds:

### 1. CACHEBUST Build Argument
Every Dockerfile includes:
```dockerfile
ARG CACHEBUST=1
```

This argument forces Docker to invalidate cache when the value changes.

### 2. Timestamp-Based Builds
Build commands include timestamps to ensure uniqueness:

**Go Services (Hub, Web API):**
```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.BuildTime=$(date +%s)" -o binary cmd/main.go
```

**Java Worker:**
```dockerfile
RUN echo "Building at $(date +%s)" && mvn clean package -DskipTests
```

**Python Worker:**
```dockerfile
RUN echo "Building at $(date +%s)"
COPY services/python-worker/*.py ./
```

### 3. Docker Compose Build Args
`docker-compose.yml` passes CACHEBUST to all services:
```yaml
services:
  hub:
    build:
      context: .
      dockerfile: Dockerfile.hub
      args:
        CACHEBUST: ${CACHEBUST:-1}
```

## Usage

### Quick Rebuild (Recommended)
Use the provided script which automatically sets CACHEBUST to current timestamp:
```bash
./rebuild.sh
```

### Manual Rebuild with Cache Busting
```bash
export CACHEBUST=$(date +%s)
docker-compose build
docker-compose down
docker-compose up -d
```

### Force Complete Rebuild (Nuclear Option)
If you still suspect cache issues:
```bash
docker-compose build --no-cache
docker-compose down
docker-compose up -d
```

## How It Works

1. **Before Code Copy**: `ARG CACHEBUST=1` is placed before COPY commands
2. **During Build**: When CACHEBUST value changes, all layers after it are invalidated
3. **Timestamp in Binary**: Build timestamps are embedded in binaries/logs for verification
4. **Default Value**: `${CACHEBUST:-1}` uses environment variable or defaults to 1

## Verification

Check that new code is running:

**Hub/Web API:**
```bash
# Check build timestamp in logs
docker-compose logs hub | grep "BuildTime"
docker-compose logs webapi | grep "BuildTime"
```

**Java Worker:**
```bash
# Check Maven build timestamp
docker-compose logs java-simple-worker | grep "Building at"
```

**Python Worker:**
```bash
# Check Docker build timestamp
docker-compose logs worker | grep "Building at"
```

## Benefits

✅ **No Manual Flags**: No need to remember `--no-cache`  
✅ **Automatic**: `rebuild.sh` handles everything  
✅ **Verifiable**: Timestamps prove fresh builds  
✅ **Consistent**: All services use same pattern  
✅ **Safe**: Preserves dependency caching (go mod, Maven dependencies, pip packages)

## Troubleshooting

### Code changes not visible after rebuild
```bash
# 1. Run rebuild script
./rebuild.sh

# 2. If still not working, check CACHEBUST was passed
docker-compose config | grep CACHEBUST

# 3. Nuclear option - remove all containers and images
docker-compose down
docker system prune -af --volumes
./rebuild.sh
```

### Performance concerns
- **Dependency layers still cached**: go mod download, Maven dependencies, pip packages remain cached
- **Only source code rebuilt**: CACHEBUST is placed after dependency installation
- **Build time**: ~30-60s for full rebuild (acceptable for reliability)

## Files Modified

- `Dockerfile.hub`: Added CACHEBUST + build timestamp
- `Dockerfile.webapi`: Added CACHEBUST + build timestamp  
- `Dockerfile.worker-simple`: Added CACHEBUST + build timestamp
- `Dockerfile.worker`: Added CACHEBUST + build timestamp
- `docker-compose.yml`: Added args.CACHEBUST to all services
- `rebuild.sh`: Automated rebuild script with timestamp

## References

- Root cause: Docker cached old Hub binary preventing metadata fix from working
- Solution tested: Manual `--no-cache` + `rm -f` resolved issue
- Prevention: This cache-busting strategy prevents recurrence
