# 🚀 Dynamic Web API Architecture

## Overview

Web API hiện tại **100% dynamic** - không cần hard-code endpoints cho từng worker capability. Tất cả routes được tự động discovered từ Hub registry.

## Architecture Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web API Server                           │
│                                                                   │
│  1. Connect to Hub                                               │
│  2. Register dynamic handler: /api/call/*                        │
│  3. Query Hub registry for capabilities                          │
│  4. Auto-log all discovered endpoints                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         gRPC Hub                                 │
│                                                                   │
│  • Registry of all worker capabilities                           │
│  • Routes requests to appropriate workers                        │
│  • Returns capability metadata                                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌──────────────┬──────────────┬──────────────┐
        │              │              │              │
┌───────▼──────┐ ┌─────▼──────┐ ┌────▼──────┐ ┌────▼──────┐
│   Python     │ │    Java    │ │   Node.js │ │     Go    │
│   Worker     │ │   Worker   │ │   Worker  │ │   Worker  │
│              │ │            │ │           │ │           │
│  • Plugins   │ │  • Plugins │ │  • Plugins│ │  • Plugins│
│  • Auto-load │ │  • Auto-   │ │  • Auto-  │ │  • Auto-  │
│              │ │    load    │ │    load   │ │    load   │
└──────────────┘ └────────────┘ └───────────┘ └───────────┘
```

## How It Works

### 1. Worker Registration (Plugin System)

Workers auto-discover và register plugins:

```python
# Python Worker
plugins/
├── hello_plugin.py          → /api/call/hello
├── calculator_plugin.py     → /api/call/calculate
├── image_analysis_plugin.py → /api/call/analyze_image
└── composite_task_plugin.py → /api/call/composite_task
```

Mỗi plugin tự động được:
- Loaded bởi PluginManager
- Registered với Hub
- Exposed qua Web API

### 2. Web API Dynamic Routing

Web API chỉ có **1 catch-all route**:

```go
// Single dynamic route handles ALL capabilities
http.HandleFunc("/api/call/", dynamicHandler.HandleDynamicCall)
```

**Cách hoạt động:**
```
Request: POST /api/call/calculate
         ↓
1. Extract capability name: "calculate"
2. Send to Hub with params
3. Hub routes to appropriate worker (Python/Java/Node/Go)
4. Worker executes plugin
5. Return response
```

### 3. Automatic Endpoint Discovery

Khi Web API start, nó tự động:

```go
// Query Hub registry
response := hubClient.SendRequest("hub", "capability_discovery", ...)

// Parse và log tất cả endpoints
for capability := range capabilities {
    log.Printf("POST /api/call/%s - %s", capName, description)
}
```

**Output example:**
```
🎯 Auto-discovered API Endpoints:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  POST  /api/call/hello                Returns a hello message
  POST  /api/call/calculate            Performs math operations
  POST  /api/call/analyze_image        Analyzes uploaded images
  POST  /api/call/composite_task       Demo worker-to-worker call
  POST  /api/call/hello_world           Java hello message
  POST  /api/call/read_file_info        Reads file information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Total: 6 dynamic endpoints from 2 workers
```

## Benefits

### ✅ Zero Configuration
- **Thêm plugin mới**: Chỉ cần tạo file trong `plugins/`
- **Không cần update Web API**: Endpoint tự động có
- **Không cần update Swagger**: Tự động generate

### ✅ Multi-Language Support
- Python, Java, Node.js, Go workers
- Tất cả dùng chung 1 Web API
- Hub tự động routing

### ✅ Scalability
- Thêm worker mới → endpoints mới tự động
- Remove worker → endpoints tự biến mất
- Hot reload plugins (future)

### ✅ Self-Documenting
- `/api/capabilities` - List tất cả capabilities
- `/api/swagger.json` - Auto-generated OpenAPI spec
- `/api/docs` - Interactive Swagger UI

## API Endpoints

### Core Endpoints (Static)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Main UI Dashboard |
| GET | `/api/capabilities` | List all capabilities |
| GET | `/api/swagger.json` | OpenAPI specification |
| GET | `/api/docs` | Swagger UI |
| GET | `/api/status` | API health check |

### Dynamic Endpoints (Auto-discovered)

| Method | Pattern | Description |
|--------|---------|-------------|
| POST | `/api/call/{capability}` | Call any worker capability |

**Example calls:**
```bash
# Python plugin: hello
curl -X POST http://localhost:8081/api/call/hello \
  -H "Content-Type: application/json" \
  -d '{}'

# Python plugin: calculate
curl -X POST http://localhost:8081/api/call/calculate \
  -H "Content-Type: application/json" \
  -d '{"operation":"add","a":5,"b":3}'

# Java plugin: read_file_info
curl -X POST http://localhost:8081/api/call/read_file_info \
  -H "Content-Type: application/json" \
  -d '{"filePath":"/etc/hosts"}'

# Python plugin with file upload
curl -X POST http://localhost:8081/api/call/analyze_image \
  -F "file=@image.jpg"
```

## Adding New Capabilities

### Option 1: Python Plugin

```bash
# 1. Create plugin
cd services/python-worker/plugins
nano my_plugin.py
```

```python
from plugins.base_plugin import BasePlugin

class MyPlugin(BasePlugin):
    @property
    def name(self) -> str:
        return "my_capability"
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        return {"result": "success"}
```

```bash
# 2. Rebuild
docker-compose -f docker-compose.all-in-one.yml up -d --build

# 3. Test
curl -X POST http://localhost:8081/api/call/my_capability -d '{}'
```

**Done!** Endpoint `/api/call/my_capability` automatically available.

### Option 2: Java Plugin

```bash
cd services/java-simple-worker/src/main/java/com/deepapp/worker/plugins
nano MyPlugin.java
```

```java
public class MyPlugin implements BasePlugin {
    @Override
    public String getName() { return "my_capability"; }
    
    @Override
    public String execute(String input, Object sdk) {
        return "{\"result\":\"success\"}";
    }
}
```

```bash
mvn clean package
docker-compose -f docker-compose.all-in-one.yml up -d --build
```

**Done!** Same endpoint automatically available.

## Code Structure

```
services/web-api/
├── main.go                          # Dynamic routing setup
└── internal/
    ├── client/
    │   └── hub_client.go           # gRPC Hub client
    └── handlers/
        ├── dynamic.go               # Dynamic capability handler
        └── status.go               # Status handler

# Old handlers removed (no longer needed):
# ❌ python.go    - Removed
# ❌ java.go      - Removed
```

## Migration from Old Architecture

### Before (Hard-coded):
```go
// Had to register every capability manually
pythonHandler := handlers.NewPythonWorkerHandler(hubClient)
javaHandler := handlers.NewJavaWorkerHandler(hubClient)

http.HandleFunc("/api/worker/python/hello", pythonHandler.HandleHello)
http.HandleFunc("/api/worker/python/analyze_image", pythonHandler.HandleAnalyzeImage)
http.HandleFunc("/api/worker/java/hello", javaHandler.HandleHello)
http.HandleFunc("/api/worker/java/file_info", javaHandler.HandleFileInfo)
// ... must add more for each new capability
```

### After (100% Dynamic):
```go
// Single dynamic handler for ALL capabilities
dynamicHandler := handlers.NewDynamicHandler(hubClient)
http.HandleFunc("/api/call/", dynamicHandler.HandleDynamicCall)

// That's it! 🎉
```

## How Dynamic Handler Works

```go
func (h *DynamicHandler) HandleDynamicCall(w http.ResponseWriter, r *http.Request) {
    // 1. Extract capability from URL
    capability := r.URL.Path[len("/api/call/"):]
    
    // 2. Parse request (JSON or multipart)
    requestData := parseRequest(r)
    
    // 3. Send to Hub (Hub knows which worker has this capability)
    response := h.hubClient.SendRequest("", capability, requestData)
    
    // 4. Return response
    json.NewEncoder(w).Encode(response)
}
```

Hub handles all the routing logic:
- Knows which worker registered which capability
- Routes request to correct worker
- Returns response

## Future Enhancements

1. **WebSocket Support** - Real-time capability updates
2. **Rate Limiting** - Per-capability rate limits
3. **Authentication** - API keys, OAuth2
4. **Caching** - Cache capability discovery
5. **Metrics** - Track usage per capability
6. **Versioning** - Support multiple versions of same capability

## Debugging

### Check what endpoints are available:
```bash
curl http://localhost:8081/api/capabilities | python3 -m json.tool
```

### View auto-generated Swagger docs:
```bash
open http://localhost:8081/api/docs
```

### Check Web API logs:
```bash
docker-compose -f docker-compose.all-in-one.yml logs webapi
```

Look for:
```
🎯 Auto-discovered API Endpoints:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  POST  /api/call/hello                Returns a hello message
  ...
```

## Summary

🎯 **Web API is now 100% dynamic!**

- ✅ No hard-coded endpoints
- ✅ Auto-discovers from Hub registry
- ✅ Works with any worker language
- ✅ Plugin system for workers
- ✅ Self-documenting via Swagger
- ✅ Zero configuration needed

**Add new capability → Restart → It just works! 🚀**
