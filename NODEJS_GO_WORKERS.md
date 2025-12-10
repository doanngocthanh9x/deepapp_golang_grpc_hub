# Complete Worker Architecture with Node.js and Go Support

## Overview
DeepApp gRPC Hub now supports **4 worker types** with plugin system and worker-to-worker communication:
- Python Worker
- Java Worker  
- Node.js Worker (NEW)
- Go Worker (NEW)

## Worker Structure

### All Workers Follow Same Pattern:

```
services/{worker-name}/
├── plugins/                    # Standard capabilities
│   ├── base_plugin.*          # Base plugin class/interface
│   ├── plugin_manager.*       # Auto-discovery engine
│   ├── hello_plugin.*         # Hello world capability
│   └── ...                    # Other plugins
├── worker-to-worker/          # Cross-worker communication
│   ├── composite_plugin.*     # Calls multiple workers
│   └── ...                    # Other bridges
└── worker-plugin-system.*     # Main worker entry point
```

## Capabilities by Worker

### Python Worker (`python-worker`)
**Standard Plugins:**
- `hello` - Hello world message
- `calculate` - Math operations (add, subtract, multiply, divide)
- `analyze_image` - Image analysis (file upload)

**Worker-to-Worker:**
- `python_composite` - Calls Java, Node.js, Go workers

### Java Worker (`java-simple-worker`)
**Standard Plugins:**
- `hello_world` - Hello world message
- `read_file_info` - File information reader

**Worker-to-Worker:**
- `java_composite` - Calls Python, Node.js, Go workers

### Node.js Worker (`node-worker`) **NEW**
**Standard Plugins:**
- `hello_node` - Hello world from Node.js
- `string_ops` - String operations (uppercase, lowercase, reverse, length)
- `json_process` - JSON processing (validate, keys, values, pretty)

**Worker-to-Worker:**
- `node_composite` - Calls Python, Java, Go workers
- `python_calc_bridge` - Bridge to Python calculator

### Go Worker (`go-worker`) **NEW**
**Standard Plugins:**
- `hello_go` - Hello world from Go
- `hash_text` - Hash computation (MD5, SHA256)
- `base64_ops` - Base64 encode/decode

**Worker-to-Worker:**
- `go_composite` - Calls Python, Java, Node.js workers

## API URL Pattern

All capabilities are accessed via:
```
POST /api/{worker_id}/call/{capability}
```

Examples:
```bash
# Node.js Hello
curl -X POST http://localhost:8081/api/node-worker/call/hello_node \
  -H "Content-Type: application/json" \
  -d '{"name":"World"}'

# Go Hash
curl -X POST http://localhost:8081/api/go-worker/call/hash_text \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello","algorithm":"sha256"}'

# Node.js Composite (calls all other workers)
curl -X POST http://localhost:8081/api/node-worker/call/node_composite \
  -H "Content-Type: application/json" \
  -d '{}'

# Go Composite (calls all other workers)
curl -X POST http://localhost:8081/api/go-worker/call/go_composite \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Worker-to-Worker Communication

Each worker can call other workers via gRPC Hub:

```python
# Python example
result = call_worker("node-worker", "string_ops", {
    "text": "hello",
    "operation": "uppercase"
})
```

```javascript
// Node.js example
const result = await context.callWorker(
    'python-worker',
    'calculate',
    { operation: 'add', a: 10, b: 5 },
    10000
);
```

```go
// Go example
result, err := context.CallWorker(
    "java-simple-worker",
    "hello_world",
    map[string]interface{}{},
    10000
)
```

```java
// Java example
String result = callWorker
    .apply("go-worker:hash_text", params)
    .get(10, TimeUnit.SECONDS);
```

## Docker Build

All workers are built in a single Dockerfile with multi-stage build:

1. **go-builder** - Builds Hub, Web API, Go Worker
2. **node-builder** - Installs Node.js dependencies  
3. **java-builder** - Builds Java Worker JAR
4. **python-builder** - Installs Python dependencies
5. **Final image** - Alpine Linux with all runtimes

## Supervisor Configuration

All services managed by Supervisor:

```ini
[program:hub]           # Priority 10 - Starts first
[program:python-worker] # Priority 20
[program:java-worker]   # Priority 20
[program:node-worker]   # Priority 20 - NEW
[program:go-worker]     # Priority 20 - NEW
[program:webapi]        # Priority 30 - Starts last
```

## Testing Composite Workflows

### Python Composite
Calls Java, Node.js, Go workers:
```bash
curl -X POST http://localhost:8081/api/python-worker/call/python_composite \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response shows results from all 3 workers with summary.

### Node.js Composite  
Calls Python, Java, Go workers:
```bash
curl -X POST http://localhost:8081/api/node-worker/call/node_composite \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Go Composite
Calls Python, Java, Node.js workers:
```bash
curl -X POST http://localhost:8081/api/go-worker/call/go_composite \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Java Composite
Calls Python, Node.js, Go workers:
```bash
curl -X POST http://localhost:8081/api/java-simple-worker/call/java_composite \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Plugin Development

### Adding New Node.js Plugin

1. Create `services/node-worker/plugins/my-plugin.js`:
```javascript
const BasePlugin = require('./base-plugin');

class MyPlugin extends BasePlugin {
    getName() {
        return "my_capability";
    }
    
    getDescription() {
        return "My custom capability";
    }
    
    async execute(params, context) {
        return {
            result: "success",
            data: params
        };
    }
}

module.exports = MyPlugin;
```

2. Rebuild container - plugin auto-discovered!

### Adding New Go Plugin

1. Create `services/go-worker/plugins/my_plugin.go`:
```go
package plugins

type MyPlugin struct {
    BasePlugin
}

func (p *MyPlugin) GetName() string {
    return "my_capability"
}

func (p *MyPlugin) GetDescription() string {
    return "My custom capability"
}

func (p *MyPlugin) Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error) {
    return map[string]interface{}{
        "result": "success",
        "data": params,
    }, nil
}
```

2. Register in `worker-plugin-system.go`:
```go
w.registerPlugin(&plugins.MyPlugin{})
```

3. Rebuild container!

## File Structure

```
services/
├── node-worker/
│   ├── package.json
│   ├── worker-plugin-system.js
│   ├── plugins/
│   │   ├── base-plugin.js
│   │   ├── plugin-manager.js
│   │   ├── hello-plugin.js
│   │   ├── string-manip-plugin.js
│   │   └── json-processor-plugin.js
│   └── worker-to-worker/
│       ├── composite-plugin.js
│       └── python-calc-bridge-plugin.js
│
├── go-worker/
│   ├── go.mod
│   ├── worker-plugin-system.go
│   ├── plugins/
│   │   ├── plugin.go
│   │   ├── hello_plugin.go
│   │   ├── hash_plugin.go
│   │   └── base64_plugin.go
│   └── worker-to-worker/
│       └── composite_plugin.go
│
├── python-worker/
│   ├── requirements.txt
│   ├── worker_plugin_system.py
│   ├── plugin_manager.py
│   ├── plugins/
│   │   ├── base_plugin.py
│   │   ├── hello_plugin.py
│   │   ├── calculator_plugin.py
│   │   └── image_analysis_plugin.py
│   └── worker-to-worker/
│       └── composite_task_plugin.py
│
└── java-simple-worker/
    ├── pom.xml
    ├── src/main/java/com/deepapp/worker/
    │   ├── WorkerPluginSystem.java
    │   ├── PluginManager.java
    │   ├── plugins/
    │   │   ├── BasePlugin.java
    │   │   ├── HelloWorldPlugin.java
    │   │   └── FileInfoPlugin.java
    │   └── workertoworker/
    │       └── CompositePlugin.java
```

## Benefits

✅ **Zero Configuration** - All plugins auto-discovered
✅ **Language Agnostic** - Python, Java, Node.js, Go all work the same way
✅ **Worker-to-Worker** - Any worker can call any other worker
✅ **Dynamic API** - Web API automatically discovers all capabilities
✅ **Single Container** - All services in one Docker container
✅ **Hot Reload Ready** - Easy to add/remove plugins

## Next Steps

1. Add more Node.js plugins (file operations, HTTP requests, etc.)
2. Add more Go plugins (concurrency examples, system info, etc.)
3. Create plugin templates for each language
4. Add plugin marketplace/registry
5. Implement plugin versioning
6. Add plugin dependencies support
