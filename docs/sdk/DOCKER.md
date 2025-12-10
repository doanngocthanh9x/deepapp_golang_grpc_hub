# gRPC Hub System - Docker Deployment

Complete containerized gRPC-based microservices system.

## Architecture

```
Browser → Web API (Go:8081) ←→ gRPC Hub (Go:50051) ←→ Python Worker
                                      ↕
                              Bidirectional gRPC
```

## Services

### 1. gRPC Hub (`hub`)
- **Port**: 50051
- **Function**: Central message routing server
- **Technology**: Go + gRPC
- **Image**: Built from `Dockerfile.hub`

### 2. Web API (`webapi`)
- **Port**: 8081
- **Function**: HTTP to gRPC gateway
- **Technology**: Go + gRPC Client
- **Image**: Built from `Dockerfile.webapi`
- **UI**: http://localhost:8081

### 3. Python Worker (`worker`)
- **Function**: Task processor (hello, image analysis)
- **Technology**: Python + gRPC Client
- **Image**: Built from `Dockerfile.worker`

## Quick Start

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+

### 1. Build and Start

```bash
# Build all images
docker-compose build

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f
```

### 2. Test the System

```bash
# Test hello endpoint
curl -X POST http://localhost:8081/api/hello

# Test with image upload
curl -X POST http://localhost:8081/api/analyze \
  -F "image=@test.jpg"

# Check status
curl http://localhost:8081/api/status

# Open Web UI
open http://localhost:8081
```

### 3. Monitor

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f hub
docker-compose logs -f webapi
docker-compose logs -f worker

# Check service status
docker-compose ps
```

### 4. Stop

```bash
# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## Using Makefile

```bash
# View all commands
make -f Makefile.docker help

# Build images
make -f Makefile.docker build

# Start services
make -f Makefile.docker up

# View logs
make -f Makefile.docker logs

# Test system
make -f Makefile.docker test

# Clean everything
make -f Makefile.docker clean

# Rebuild from scratch
make -f Makefile.docker rebuild
```

## Configuration

### Environment Variables

#### Hub Service
- `PORT`: gRPC port (default: 50051)
- `LOG_LEVEL`: Logging level (default: info)
- `DB_PATH`: Database path (default: /data/hub.db)

#### Web API Service
- `HUB_ADDRESS`: Hub address (default: hub:50051)
- `PORT`: HTTP port (default: 8081)

#### Worker Service
- `HUB_ADDRESS`: Hub address (default: hub:50051)
- `WORKER_ID`: Worker identifier (default: python-worker)

### Customize docker-compose.yml

```yaml
services:
  webapi:
    environment:
      - HUB_ADDRESS=hub:50051
      - PORT=8081
    ports:
      - "8081:8081"  # Change external port here
```

## Development

### Local Development with Docker

```bash
# Build and start with live reload
docker-compose up --build

# Rebuild single service
docker-compose up --build webapi

# Execute commands in container
docker-compose exec hub sh
docker-compose exec worker python -c "import sys; print(sys.version)"
```

### Build Individual Images

```bash
# Build Hub
docker build -f Dockerfile.hub -t grpc-hub:latest .

# Build Web API
docker build -f Dockerfile.webapi -t web-api:latest .

# Build Worker
docker build -f Dockerfile.worker -t python-worker:latest .
```

### Run Individual Containers

```bash
# Create network
docker network create grpc-network

# Run Hub
docker run -d --name hub --network grpc-network -p 50051:50051 grpc-hub:latest

# Run Worker
docker run -d --name worker --network grpc-network \
  -e HUB_ADDRESS=hub:50051 python-worker:latest

# Run Web API
docker run -d --name webapi --network grpc-network -p 8081:8081 \
  -e HUB_ADDRESS=hub:50051 web-api:latest
```

## Troubleshooting

### Check Service Health

```bash
# All services
docker-compose ps

# Individual service logs
docker-compose logs hub
docker-compose logs webapi
docker-compose logs worker
```

### Common Issues

**1. Port already in use**
```bash
# Find process using port
lsof -ti:8081 | xargs kill -9
lsof -ti:50051 | xargs kill -9
```

**2. Services can't connect**
```bash
# Check network
docker network ls
docker network inspect deepapp_golang_grpc_hub_grpc-network

# Restart services
docker-compose restart
```

**3. Build errors**
```bash
# Clean build
docker-compose down
docker system prune -f
docker-compose build --no-cache
docker-compose up
```

## Production Deployment

### Use Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml grpc-hub

# Scale worker
docker service scale grpc-hub_worker=3

# Remove stack
docker stack rm grpc-hub
```

### Use Kubernetes

```bash
# Convert docker-compose to k8s
kompose convert -f docker-compose.yml

# Apply to cluster
kubectl apply -f .

# Scale deployment
kubectl scale deployment worker --replicas=3
```

## API Endpoints

### Web API (http://localhost:8081)

- `GET /` - Web UI
- `POST /api/hello` - Test hello world
- `POST /api/analyze` - Image analysis (multipart/form-data)
- `GET /api/status` - System status

### Examples

```bash
# Hello
curl -X POST http://localhost:8081/api/hello

# Upload image
curl -X POST http://localhost:8081/api/analyze \
  -F "image=@photo.jpg"

# Status
curl http://localhost:8081/api/status
```

## Monitoring

### Prometheus + Grafana (Optional)

Add to `docker-compose.yml`:

```yaml
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- Restart services: `docker-compose restart`
- Rebuild: `make -f Makefile.docker rebuild`
