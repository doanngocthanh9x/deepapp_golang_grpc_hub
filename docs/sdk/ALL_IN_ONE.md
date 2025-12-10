# All-in-One Container Architecture

## Overview
Single Docker container running all services:
- **Hub** (gRPC message router) - localhost:50051
- **Python Worker** - connects to Hub
- **Java Worker** - connects to Hub  
- **Web API** (HTTP gateway) - **PUBLIC** on port 8081

## Architecture Benefits

### ✅ Advantages
1. **Simplified Deployment**: One container to manage
2. **No Network Overhead**: Services communicate via localhost (faster)
3. **Easy Scaling**: Scale entire stack together
4. **Reduced Resource Usage**: Shared base image layers
5. **Atomic Updates**: All services updated together

### ⚠️ Trade-offs
1. **No Independent Scaling**: Can't scale individual services
2. **Larger Image Size**: Contains all dependencies
3. **Restart All**: One service crash requires full restart
4. **Less Isolation**: Services share same container

## File Structure

```
Dockerfile.all-in-one           # Multi-stage build for all services
docker-compose.all-in-one.yml   # Single service compose file
start-all-in-one.sh            # Build and start script
logs.sh                         # View logs easily
```

## Build Process

### Multi-Stage Build (Optimized)

**Stage 1: Go Builder**
- ✅ CACHED: `go mod download` (dependencies)
- ❌ NOT CACHED: Source code copy (uses CACHEBUST)
- Builds: Hub + Web API binaries

**Stage 2: Java Builder**  
- ✅ CACHED: Maven dependency download
- ❌ NOT CACHED: Source code copy (uses CACHEBUST)
- Builds: Java Worker JAR

**Stage 3: Python Builder**
- ✅ CACHED: `pip install` (dependencies)

**Stage 4: Final Runtime**
- Combines all binaries
- ❌ NOT CACHED: Python source code (uses CACHEBUST)
- Installs supervisor for process management

## Usage

### Quick Start
```bash
# Build and start all services
./start-all-in-one.sh
```

### Manual Build
```bash
# Force code rebuild (libraries cached)
export CACHEBUST=$(date +%s)
docker-compose -f docker-compose.all-in-one.yml build

# Start container
docker-compose -f docker-compose.all-in-one.yml up -d
```

### View Logs
```bash
# All service status
./logs.sh status

# Individual service logs (live)
./logs.sh hub
./logs.sh webapi
./logs.sh java
./logs.sh python

# All logs snapshot
./logs.sh all
```

### Stop/Restart
```bash
# Stop
docker-compose -f docker-compose.all-in-one.yml down

# Restart
docker-compose -f docker-compose.all-in-one.yml restart

# Rebuild and restart
./start-all-in-one.sh
```

## Service Management

### Supervisor Controls
```bash
# View all service status
docker exec deepapp-hub-all-in-one supervisorctl status

# Restart individual service
docker exec deepapp-hub-all-in-one supervisorctl restart hub
docker exec deepapp-hub-all-in-one supervisorctl restart webapi
docker exec deepapp-hub-all-in-one supervisorctl restart java-worker
docker exec deepapp-hub-all-in-one supervisorctl restart python-worker

# Stop/Start individual service
docker exec deepapp-hub-all-in-one supervisorctl stop java-worker
docker exec deepapp-hub-all-in-one supervisorctl start java-worker
```

## Testing

### Health Check
```bash
curl http://localhost:8081/api/status
```

### Test Java Worker
```bash
# Hello World capability
curl -X POST http://localhost:8081/api/request/java-simple-worker/hello_world \
  -H 'Content-Type: application/json' \
  -d '{"name":"World"}'

# File Info capability
curl -X POST http://localhost:8081/api/request/java-simple-worker/read_file_info \
  -H 'Content-Type: application/json' \
  -d '{"path":"/app/java-worker.jar"}'
```

### Test Python Worker
```bash
# Execute Python code
curl -X POST http://localhost:8081/api/request/python-worker/execute \
  -H 'Content-Type: application/json' \
  -d '{"code":"print(2 + 2)"}'
```

## Cache Strategy

### ✅ What Gets Cached (Fast Rebuilds)
- Go modules (`go mod download`)
- Maven dependencies (`mvn dependency:go-offline`)
- Python packages (`pip install`)
- Base OS packages

### ❌ What Doesn't Get Cached (Fresh Code)
- Go source code (Hub, Web API)
- Java source code (Java Worker)
- Python source code (Python Worker)

### How It Works
1. `ARG CACHEBUST=1` placed **after** dependency installation
2. `CACHEBUST` value changes → invalidates cache for source code
3. Dependencies installed **before** CACHEBUST → remain cached
4. Result: Fast builds (~30s) with guaranteed fresh code

## Network Architecture

### Internal Communication (localhost)
```
┌─────────────────────────────────────────┐
│  Container: deepapp-hub-all-in-one      │
│                                         │
│  ┌──────────┐                          │
│  │   Hub    │ :50051                   │
│  └────┬─────┘                          │
│       │ localhost                       │
│  ┌────┴─────────┬─────────┐           │
│  │              │         │            │
│  ▼              ▼         ▼            │
│ Python       Java      WebAPI          │
│ Worker       Worker    :8081           │
│                         │               │
└─────────────────────────┼───────────────┘
                          │
                          ▼
                    Public: :8081
```

### Port Exposure
- **8081** (Web API) → PUBLIC
- **50051** (Hub) → INTERNAL ONLY
- Workers → INTERNAL ONLY

## Data Persistence

### Volume Mounts
```yaml
volumes:
  - hub-data:/data  # Hub database persisted
```

### Database Location
- Inside container: `/data/hub.db`
- Volume: `hub-data` (Docker managed)

### Backup Database
```bash
# Copy database out
docker cp deepapp-hub-all-in-one:/data/hub.db ./hub-backup.db

# Restore database
docker cp ./hub-backup.db deepapp-hub-all-in-one:/data/hub.db
docker exec deepapp-hub-all-in-one supervisorctl restart hub
```

## Troubleshooting

### Service Won't Start
```bash
# Check supervisor logs
docker exec deepapp-hub-all-in-one cat /var/log/supervisor/supervisord.log

# Check individual service
./logs.sh status
./logs.sh hub
```

### Hub Connection Issues
```bash
# Verify Hub is running
docker exec deepapp-hub-all-in-one netstat -tlnp | grep 50051

# Test Hub connection from inside container
docker exec deepapp-hub-all-in-one nc -zv localhost 50051
```

### Code Changes Not Applied
```bash
# Rebuild with fresh code
export CACHEBUST=$(date +%s)
docker-compose -f docker-compose.all-in-one.yml build --no-cache
docker-compose -f docker-compose.all-in-one.yml up -d

# Or use helper script
./start-all-in-one.sh
```

### Check Build Times
```bash
# Verify fresh code was built
docker exec deepapp-hub-all-in-one ls -la /app/
docker exec deepapp-hub-all-in-one java -version
./logs.sh all | grep "Building"
```

## Performance

### Build Times
- **First build**: 3-5 minutes (downloads all dependencies)
- **Code-only rebuild**: 30-60 seconds (dependencies cached)
- **Full rebuild**: 3-5 minutes (with `--no-cache`)

### Resource Usage
- **Memory**: ~500MB-1GB (all services combined)
- **CPU**: Minimal when idle
- **Disk**: ~800MB image size

### Optimization Tips
1. Use `CACHEBUST` for normal rebuilds (fast)
2. Only use `--no-cache` when dependencies change
3. Keep base images updated (`docker pull python:3.10-slim`)
4. Prune old images: `docker image prune -f`

## Comparison with Multi-Container

| Feature | All-in-One | Multi-Container |
|---------|------------|-----------------|
| Containers | 1 | 4 |
| Network | localhost | Docker network |
| Scaling | Together | Independent |
| Complexity | Low | Medium |
| Isolation | Low | High |
| Startup Time | Fast | Slower |
| Resource Usage | Lower | Higher |
| Debugging | Harder | Easier |

## When to Use All-in-One

### ✅ Good For
- Development environments
- Simple deployments
- Resource-constrained systems
- Tightly coupled services
- Quick prototyping

### ❌ Not Good For
- Production (prefer multi-container)
- Independent service scaling
- Service isolation requirements
- Complex debugging scenarios
- Multiple environments (dev/staging/prod)

## Migration Path

### From All-in-One to Multi-Container
```bash
# Stop all-in-one
docker-compose -f docker-compose.all-in-one.yml down

# Start multi-container
docker-compose up -d
```

### From Multi-Container to All-in-One
```bash
# Stop multi-container
docker-compose down

# Start all-in-one
./start-all-in-one.sh
```

## Environment Variables

All services configured via supervisor config:
- `PORT=50051` - Hub gRPC port
- `LOG_LEVEL=info` - Hub log level
- `DB_PATH=/data/hub.db` - Hub database path
- `HUB_ADDRESS=localhost:50051` - Worker/API connection
- `WORKER_ID=*` - Worker identification

## Security Considerations

1. **Internal Services**: Hub and workers not exposed externally
2. **Single Entry Point**: Only Web API is public
3. **No SSH**: Use `docker exec` for container access
4. **Volume Permissions**: Database persisted with proper permissions
5. **Network Isolation**: Can add to custom Docker network if needed

## Summary

✅ **Optimized for**: Fast rebuilds, simple deployment, development  
✅ **Cached**: All dependencies (Go modules, Maven, pip)  
❌ **Not Cached**: Source code (ensures fresh code every build)  
🚀 **Result**: 30-second rebuilds with guaranteed code freshness
