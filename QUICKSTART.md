# Quick Start Guide

## System Demo: HTTP → gRPC → Python Worker

Hệ thống demo hoàn chỉnh với 3 components:

1. **gRPC Hub** - Message broker (port 50051)
2. **Web API (Go)** - HTTP gateway (port 8080)
3. **Python Worker** - Task processor

## Cách 1: Chạy Tự Động (Recommended)

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
./start-demo.sh
```

Script sẽ tự động:
- Khởi động gRPC Hub
- Khởi động Python Worker
- Khởi động Web API
- Hiển thị status

## Cách 2: Chạy Thủ Công

### Terminal 1: gRPC Hub
```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run cmd/hub/main.go
```

### Terminal 2: Python Worker
```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub/services/python-worker
python3 worker.py
```

### Terminal 3: Web API
```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run services/web-api/main.go
```

## Truy Cập Web UI

Mở browser: **http://localhost:8080**

## Test với cURL

### 1. Hello World

```bash
curl -X POST http://localhost:8080/api/hello
```

Response:
```json
{
  "status": "success",
  "response": "{\"message\":\"Hello World from Python Worker! 🐍\",\"status\":\"success\"}",
  "from": "python-worker",
  "timestamp": "2025-12-08T11:30:00Z"
}
```

### 2. Analyze Image

```bash
# Create a test image first
convert -size 800x600 xc:blue test.jpg  # Or use any image

# Upload and analyze
curl -X POST http://localhost:8080/api/analyze \
  -F "image=@test.jpg"
```

Response:
```json
{
  "status": "success",
  "response": "{\"filename\":\"test.jpg\",\"size_bytes\":1024000,\"format\":\"JPG\",\"analysis\":{...}}",
  "from": "python-worker",
  "timestamp": "2025-12-08T11:30:05Z"
}
```

### 3. Check Status

```bash
curl http://localhost:8080/api/status
```

## Kiểm Tra Services

```bash
# Check if gRPC Hub is running
lsof -i :50051

# Check if Web API is running
lsof -i :8080

# Check Python process
ps aux | grep worker.py
```

## Flow Diagram

```
Browser/cURL
    ↓ HTTP POST /api/hello
[Web API :8080]
    ↓ gRPC Message
    {
      from: "web-api-xxx",
      to: "python-worker",
      channel: "hello",
      type: DIRECT
    }
    ↓
[gRPC Hub :50051]
    ↓ Route to worker
[Python Worker]
    ↓ Process & Response
    {
      message: "Hello World from Python! 🐍",
      status: "success"
    }
    ↓
[gRPC Hub]
    ↓ Route back
[Web API]
    ↓ HTTP Response
{
  "status": "success",
  "response": "...",
  "from": "python-worker"
}
```

## Features Implemented

✅ **Web API (Go)**
- HTTP server with REST endpoints
- gRPC client connection
- Request/response pattern with timeout
- File upload support
- Web UI for testing

✅ **Python Worker**
- Hello world handler
- Image analysis (mock)
- JSON serialization
- Error handling

✅ **gRPC Hub**
- Message routing
- Bidirectional streaming
- Connection management

## What You Can Test

1. **Hello World**: Simple request-response
2. **Image Upload**: File handling and analysis
3. **Status Check**: Service health monitoring

## Extending

### Add New Python Handler

Edit `services/python-worker/worker.py`:

```python
def handle_new_task(self, request_data):
    return {
        "result": "Your result",
        "status": "success"
    }
```

### Add New API Endpoint

Edit `services/web-api/main.go`:

```go
http.HandleFunc("/api/newtask", func(w http.ResponseWriter, r *http.Request) {
    response, err := api.hubClient.SendRequest(
        "python-worker",
        "new_task",
        "data"
    )
    // Handle response...
})
```

## Troubleshooting

**Port already in use:**
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9

# Kill process on port 50051
lsof -ti:50051 | xargs kill -9
```

**Python dependencies missing:**
```bash
cd services/python-worker
pip3 install -r requirements.txt
```

**Go dependencies missing:**
```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go mod tidy
```

## Stop Services

Press `Ctrl+C` in each terminal, or:

```bash
# Kill all related processes
pkill -f "cmd/hub/main.go"
pkill -f "services/web-api/main.go"
pkill -f "worker.py"
```

## Next Steps

- Add authentication
- Implement real image processing (OpenCV, TensorFlow)
- Add database for request logging
- Implement load balancing
- Add monitoring and metrics

## Documentation

- Full API docs: `/services/README.md`
- SDK documentation: `/docs/sdk/`
- Main README: `/README.md`
