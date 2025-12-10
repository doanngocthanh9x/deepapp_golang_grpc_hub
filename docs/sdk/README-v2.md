# Dynamic gRPC Hub System v2

## 🎯 Tính năng mới

### 1. **Dynamic Service Registration**
Workers tự động đăng ký capabilities khi kết nối với Hub:
- Không cần hardcode API endpoints
- Workers declare capabilities với schema
- Hub tự động route requests đến workers phù hợp

### 2. **Auto-Discovery API**
Web API tự động discover và expose capabilities từ workers:
- GET `/api/v2/capabilities` - List tất cả services available
- POST `/api/v2/invoke/{capability}` - Gọi bất kỳ service nào dynamically

### 3. **Swagger/OpenAPI Documentation**
- Tự động generate API docs từ code annotations
- Interactive API testing tại `/swagger/index.html`
- Schema validation cho inputs/outputs

### 4. **Load Balancing & Health Check**
- Hub tracks worker status (online/offline)
- Simple round-robin routing
- Health monitoring

## 🏗️ Architecture

```
┌─────────────────┐
│   Client/User   │
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐
│  Web API v2     │◄────┐
│  (Gin+Swagger)  │     │
└────────┬────────┘     │
         │ gRPC         │ Auto-discovery
         ▼              │
┌─────────────────┐     │
│   gRPC Hub      │─────┘
│ (Service        │
│  Registry)      │
└────────┬────────┘
         │ gRPC (bidirectional streaming)
         │
    ┌────┴────┬────────┬────────┐
    ▼         ▼        ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│Worker 1│ │Worker 2│ │Worker 3│ │Worker N│
│(Python)│ │(Python)│ │(Go)    │ │(...)   │
└────────┘ └────────┘ └────────┘ └────────┘
```

## 🚀 Quick Start

### Prerequisites
- Go 1.18+
- Python 3.10+
- Docker & Docker Compose (optional)

### 1. Install Swagger Tool
```bash
make -f Makefile.v2 install-swag
```

### 2. Build Services
```bash
make -f Makefile.v2 build
```

### 3. Run Services (Local)

**Terminal 1: Hub**
```bash
make -f Makefile.v2 run-hub
```

**Terminal 2: Python Worker**
```bash
make -f Makefile.v2 run-worker-v2
```

**Terminal 3: Web API**
```bash
make -f Makefile.v2 run-api-v2
```

### 4. Access Swagger UI
```
http://localhost:8082/swagger/index.html
```

## 🐳 Docker Deployment

### Build and Start
```bash
make -f Makefile.v2 docker-build
make -f Makefile.v2 docker-up
```

### Test
```bash
make -f Makefile.v2 test-api
```

### View Logs
```bash
make -f Makefile.v2 docker-logs
```

### Stop
```bash
make -f Makefile.v2 docker-down
```

## 📝 Adding New Capabilities

### Python Worker Example

```python
# In worker_dynamic.py

# 1. Define handler
def handle_my_feature(self, payload):
    text = payload.get("text", "")
    return {
        "result": text.upper(),
        "length": len(text)
    }

# 2. Register in __init__
self.capabilities = {
    # ... existing capabilities ...
    "my_feature": {
        "name": "my_feature",
        "description": "My awesome feature",
        "input_schema": json.dumps({
            "type": "object",
            "properties": {
                "text": {"type": "string"}
            }
        }),
        "output_schema": json.dumps({
            "type": "object",
            "properties": {
                "result": {"type": "string"},
                "length": {"type": "integer"}
            }
        }),
        "handler": self.handle_my_feature
    }
}
```

**That's it!** API tự động expose endpoint mới:
```bash
curl -X POST http://localhost:8082/api/v2/invoke/my_feature \
  -H "Content-Type: application/json" \
  -d '{"text":"hello world"}'
```

## 🔧 API Endpoints

### System Endpoints

**GET /api/v2/status**
```json
{
  "client_id": "web-api-v2-...",
  "status": "running",
  "version": "2.0",
  "capabilities_count": 3
}
```

**GET /api/v2/capabilities**
```json
{
  "capabilities": {
    "hello": {
      "name": "hello",
      "description": "Simple hello greeting service",
      "input_schema": "{...}",
      "output_schema": "{...}"
    },
    "image_analysis": {...},
    "text_processing": {...}
  },
  "timestamp": "2025-12-08T08:00:00Z"
}
```

### Dynamic Service Invocation

**POST /api/v2/invoke/{capability}**

Example: Hello service
```bash
curl -X POST http://localhost:8082/api/v2/invoke/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"Dynamic User"}'
```

Response:
```json
{
  "message": "Hello, Dynamic User! From worker py-worker-dynamic"
}
```

Example: Text processing
```bash
curl -X POST http://localhost:8082/api/v2/invoke/text_processing \
  -H "Content-Type: application/json" \
  -d '{"text":"hello world","operation":"uppercase"}'
```

Response:
```json
{
  "result": "HELLO WORLD",
  "length": 11,
  "operation": "uppercase",
  "processed_by": "py-worker-dynamic"
}
```

## 📊 Service Registry

Hub maintains a service registry:

```go
type ServiceRegistry struct {
    workers      map[string]*WorkerInfo
    capabilities map[string][]string  // capability -> workers
}
```

**Features:**
- Auto-registration on worker connect
- Auto-cleanup on worker disconnect
- Load balancing (round-robin)
- Health tracking

## 🔄 Message Flow

1. **Registration:**
```
Worker → Hub: REGISTER message with capabilities
Hub → Registry: Store worker info
Hub → Worker: Confirmation
```

2. **Discovery:**
```
API → Hub: List capabilities request
Hub → Registry: Get all capabilities
Hub → API: Capabilities list
```

3. **Service Call:**
```
Client → API: POST /invoke/hello
API → Hub: REQUEST with capability name
Hub → Registry: Find worker for "hello"
Hub → Worker: Route request
Worker → Handler: Process
Worker → Hub: RESPONSE
Hub → API: Route response
API → Client: JSON response
```

## 🎓 Benefits

1. **Scalability:**
   - Add workers without changing API code
   - Multiple workers per capability (load balancing)
   - Horizontal scaling

2. **Maintainability:**
   - One place to add features (worker)
   - No API code changes needed
   - Self-documenting via Swagger

3. **Flexibility:**
   - Workers in any language (Python, Go, Node.js, etc.)
   - Dynamic schema validation
   - Runtime service discovery

4. **Developer Experience:**
   - Interactive Swagger UI for testing
   - Auto-generated API docs
   - Type-safe schemas

## 📦 Project Structure

```
deepapp_golang_grpc_hub/
├── cmd/hub/                    # Hub server
├── internal/
│   ├── hub/
│   │   ├── registry.go        # NEW: Service registry
│   │   ├── handlers.go        # NEW: Registration handlers
│   │   └── server.go          # Updated with registry
│   └── proto/                 # Updated proto definitions
├── services/
│   ├── web-api-v2/            # NEW: Dynamic API with Swagger
│   │   ├── main.go
│   │   ├── docs/              # Auto-generated Swagger docs
│   │   └── go.mod
│   └── python-worker-v2/      # NEW: Dynamic worker
│       ├── worker_dynamic.py
│       └── requirements.txt
├── Dockerfile.webapi-v2       # NEW
├── Dockerfile.worker-v2       # NEW
├── docker-compose-v2.yml      # NEW
├── Makefile.v2                # NEW: Build & run commands
└── README-v2.md               # This file
```

## 🚧 Next Steps

- [ ] Add authentication/authorization
- [ ] Implement rate limiting
- [ ] Add metrics & monitoring (Prometheus)
- [ ] Health check endpoints for all services
- [ ] Implement retry logic
- [ ] Add request validation middleware
- [ ] Support for streaming responses
- [ ] gRPC web for browser clients
