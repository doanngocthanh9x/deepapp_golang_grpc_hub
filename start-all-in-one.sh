#!/bin/bash
# Build and run all-in-one container

echo "🔨 Building all-in-one container at $(date)"
echo "📝 Using CACHEBUST=$(date +%s) to force code rebuild (libraries cached)"

# Export CACHEBUST to force code rebuild while keeping library cache
export CACHEBUST=$(date +%s)

# Build the all-in-one image
echo "🏗️  Building all-in-one image..."
docker-compose -f docker-compose.all-in-one.yml build

# Stop and remove old container
echo "🛑 Stopping old container..."
docker-compose -f docker-compose.all-in-one.yml down

# Start new container
echo "🚀 Starting all-in-one container..."
docker-compose -f docker-compose.all-in-one.yml up -d

# Wait for services to start
echo "⏳ Waiting for services to start..."
sleep 5

# Show status
echo ""
echo "✅ All-in-one container started!"
echo ""
echo "📊 Service status:"
docker exec deepapp-hub-all-in-one supervisorctl status

echo ""
echo "📝 Logs:"
echo "  View all logs:    docker-compose -f docker-compose.all-in-one.yml logs -f"
echo "  View hub logs:    docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/hub.out.log"
echo "  View webapi logs: docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/webapi.out.log"
echo "  View java logs:   docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/java-worker.out.log"
echo "  View python logs: docker exec deepapp-hub-all-in-one tail -f /var/log/supervisor/python-worker.out.log"
echo ""
echo "🌐 Web API available at: http://localhost:8081"
echo "🧪 Test endpoints:"
echo "  curl http://localhost:8081/api/status"
echo "  curl -X POST http://localhost:8081/api/request/java-simple-worker/hello_world -H 'Content-Type: application/json' -d '{\"name\":\"World\"}'"
echo "  curl -X POST http://localhost:8081/api/request/java-simple-worker/read_file_info -H 'Content-Type: application/json' -d '{\"path\":\"/app/java-worker.jar\"}'"
