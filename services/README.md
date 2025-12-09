# Demo System: HTTP API → gRPC Hub → Python Worker

## Architecture

```
Client (Browser/Curl)
    ↓ HTTP
Web Go API (Port 8080)
    ↓ gRPC
gRPC Hub (Port 50051)
    ↓ gRPC
Python Worker
```

## Services

### 1. gRPC Hub Server
**Location**: `/cmd/hub/main.go`  
**Port**: 50051  
**Role**: Central message broker

### 2. Web Go API
**Location**: `/services/web-api/main.go`  
**Port**: 8080  
**Role**: HTTP to gRPC gateway

**Endpoints**:
- `GET /` - Web UI for testing
- `POST /api/hello` - Get hello message from Python
- `POST /api/analyze` - Analyze uploaded image
- `GET /api/status` - API status

### 3. Python Worker
**Location**: `/services/python-worker/worker.py`  
**Role**: Process tasks (hello, image analysis)

## Quick Start

### Terminal 1: Start gRPC Hub

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run cmd/hub/main.go
```

### Terminal 2: Start Python Worker

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub/services/python-worker
pip3 install -r requirements.txt
python3 worker.py
```

### Terminal 3: Start Web API

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run services/web-api/main.go
```

### Access Web UI

Open browser: http://localhost:8080

## Testing with cURL

### 1. Hello World

```bash
curl -X POST http://localhost:8080/api/hello
```

Expected response:
```json
{
  "status": "success",
  "response": "{\"message\":\"Hello World from Python Worker!\",\"status\":\"success\"}",
  "from": "python-worker",
  "timestamp": "2025-12-08T11:30:00Z"
}
```

### 2. Image Analysis

```bash
curl -X POST http://localhost:8080/api/analyze \
  -F "image=@/path/to/image.jpg"
```

Expected response:
```json
{
  "status": "success",
  "response": "{\"filename\":\"image.jpg\",\"size_bytes\":1024000,\"format\":\"JPG\",\"analysis\":{...}}",
  "from": "python-worker",
  "timestamp": "2025-12-08T11:30:00Z"
}
```

### 3. Status Check

```bash
curl http://localhost:8080/api/status
```

## Flow Diagram

```
1. Client Request
   ↓
2. Web API receives HTTP request
   ↓
3. Web API creates gRPC message
   {
     "id": "req-123",
     "from": "web-api-456",
     "to": "python-worker",
     "channel": "hello" or "analyze_image",
     "content": "{...}",
     "type": "DIRECT"
   }
   ↓
4. gRPC Hub routes message to Python Worker
   ↓
5. Python Worker processes request
   - hello: Returns greeting message
   - analyze_image: Analyzes image and returns metadata
   ↓
6. Python Worker sends response via gRPC Hub
   {
     "from": "python-worker",
     "to": "web-api-456",
     "content": "{result...}"
   }
   ↓
7. Web API receives response
   ↓
8. Web API converts to HTTP JSON response
   ↓
9. Client receives final result
```

## Message Format

### Request Message (Web API → Python)
```go
Message{
    Id:        "req-timestamp",
    From:      "web-api-client-id",
    To:        "python-worker",
    Channel:   "hello" | "analyze_image",
    Content:   "request data as JSON string",
    Type:      MessageType_DIRECT,
    Timestamp: "2025-12-08T11:30:00Z"
}
```

### Response Message (Python → Web API)
```go
Message{
    Id:        "resp-timestamp",
    From:      "python-worker",
    To:        "web-api-client-id",
    Content:   "response data as JSON string",
    Type:      MessageType_DIRECT,
    Timestamp: "2025-12-08T11:30:05Z"
}
```

## Features

### Web API Features
✅ HTTP to gRPC gateway  
✅ Async request/response pattern  
✅ File upload support  
✅ Timeout handling (30s)  
✅ Web UI for testing  

### Python Worker Features
✅ Hello World handler  
✅ Image analysis (mock)  
✅ JSON request/response  
✅ Error handling  
✅ Extensible handler system  

## Extending the System

### Add New Handler in Python Worker

```python
def handle_new_task(self, request_data):
    """Handle new task"""
    # Process request
    result = {
        "result": "...",
        "status": "success"
    }
    return result

# Register in process_request()
handlers = {
    'hello': self.handle_hello,
    'analyze_image': self.handle_image_analysis,
    'new_task': self.handle_new_task  # Add here
}
```

### Add New Endpoint in Web API

```go
http.HandleFunc("/api/newtask", func(w http.ResponseWriter, r *http.Request) {
    response, err := api.hubClient.SendRequest("python-worker", "new_task", data)
    // ... handle response
})
```

## Troubleshooting

### Connection Refused
- Ensure gRPC Hub is running on port 50051
- Check firewall settings

### Timeout Errors
- Increase timeout in `SendRequest()` method
- Check if Python worker is running
- Verify worker ID matches ("python-worker")

### No Response from Python
- Check Python worker logs
- Verify channel name matches
- Ensure worker is connected to hub

## Production Considerations

1. **Authentication**: Add JWT/API keys
2. **Rate Limiting**: Implement rate limits
3. **Monitoring**: Add metrics and logging
4. **Error Recovery**: Implement retry logic
5. **Load Balancing**: Multiple Python workers
6. **TLS**: Enable secure gRPC connections

## Performance

- **Latency**: ~50-200ms per request
- **Throughput**: Depends on worker count
- **Concurrent Requests**: Handled via goroutines

## Testing

```bash
# Test Python worker functions
cd services/python-worker
python3 worker.py test

# Load test Web API
ab -n 1000 -c 10 http://localhost:8080/api/hello
```
