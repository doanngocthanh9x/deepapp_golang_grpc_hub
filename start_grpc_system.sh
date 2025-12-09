#!/bin/bash
# Start complete gRPC system: Hub + Python Worker + Web API

echo "🚀 Starting gRPC System"
echo "=" | tr '=' '-' | head -c 50 && echo

# Stop any existing processes
echo "Stopping existing processes..."
pkill -f "cmd/hub/main.go" 2>/dev/null
pkill -f "worker_grpc.py" 2>/dev/null  
pkill -f "web-api/main.go" 2>/dev/null
sleep 2

# Start Hub
echo "Starting gRPC Hub..."
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run cmd/hub/main.go > /tmp/hub.log 2>&1 &
HUB_PID=$!
echo "  Hub PID: $HUB_PID"
sleep 3

# Start Python Worker
echo "Starting Python Worker..."
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub/services/python-worker
python3 worker_grpc.py > /tmp/worker.log 2>&1 &
WORKER_PID=$!
echo "  Worker PID: $WORKER_PID"
sleep 2

# Start Web API
echo "Starting Web API..."
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run services/web-api/main.go > /tmp/webapi.log 2>&1 &
API_PID=$!
echo "  API PID: $API_PID"
sleep 2

echo ""
echo "=" | tr '=' '-' | head -c 50 && echo
echo "✓ System started!"
echo ""
echo "Services:"
echo "  - gRPC Hub:     localhost:50051"
echo "  - Web API:      http://localhost:8081"
echo "  - Python Worker: Connected via gRPC"
echo ""
echo "Logs:"
echo "  tail -f /tmp/hub.log"
echo "  tail -f /tmp/worker.log"
echo "  tail -f /tmp/webapi.log"
echo ""
echo "Test:"
echo "  curl -X POST http://localhost:8081/api/hello"
echo "  Open http://localhost:8081 in browser"
echo ""
echo "Stop:"
echo "  kill $HUB_PID $WORKER_PID $API_PID"
echo ""
