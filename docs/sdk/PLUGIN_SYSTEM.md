# 🔌 Worker Plugin System

## Overview

Plugin system tự động cho gRPC Worker Hub - giúp bạn tạo capabilities mới chỉ bằng cách thêm file vào thư mục `plugins/`. Không cần phải register thủ công hay setup gì thêm!

## 🎯 Features

- **Auto-Discovery**: Tự động scan và load tất cả plugins
- **Zero Configuration**: Không cần register thủ công
- **Hot Reload Ready**: Dễ dàng thêm/sửa plugins
- **Web API Auto-Routing**: Hub tự động tạo endpoints
- **Worker-to-Worker**: Plugins có thể gọi capabilities của workers khác
- **Multi-Language**: Python, Java, Node.js, Go

## 📁 Project Structure

```
services/
├── python-worker/
│   ├── plugins/
│   │   ├── __init__.py
│   │   ├── base_plugin.py           # Base class
│   │   ├── hello_plugin.py          # Example plugin
│   │   ├── image_analysis_plugin.py # File upload example
│   │   └── composite_task_plugin.py # Worker-to-worker example
│   ├── plugin_manager.py            # Plugin loader
│   └── worker_plugin_system.py      # Worker with plugin system
│
├── java-simple-worker/
│   └── src/main/java/com/deepapp/worker/
│       ├── plugins/
│       │   ├── BasePlugin.java        # Base interface
│       │   ├── HelloWorldPlugin.java  # Example plugin
│       │   └── FileInfoPlugin.java    # File info example
│       └── PluginManager.java         # Plugin loader
│
├── node-worker/                       # Coming soon
└── go-worker/                         # Coming soon
```

## 🚀 Quick Start

### Python Worker

#### 1. Tạo Plugin Mới

```python
# plugins/my_plugin.py
from datetime import datetime
from plugins.base_plugin import BasePlugin

class MyPlugin(BasePlugin):
    @property
    def name(self) -> str:
        return "my_capability"  # Tên capability
    
    @property
    def description(self) -> str:
        return "Does something awesome"
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        # Logic của bạn ở đây
        result = {"message": "Hello!", "timestamp": datetime.now().isoformat()}
        return result
```

#### 2. Chạy Worker

```bash
cd services/python-worker
python3 worker_plugin_system.py
```

**Xong!** Plugin của bạn đã được:
- ✅ Auto-loaded
- ✅ Registered với Hub
- ✅ Web API endpoint tự động tạo: `POST /api/call/my_capability`

### Java Worker

#### 1. Tạo Plugin Mới

```java
// plugins/MyPlugin.java
package com.deepapp.worker.plugins;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.*;

public class MyPlugin implements BasePlugin {
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public String getName() {
        return "my_capability";
    }
    
    @Override
    public String getDescription() {
        return "Does something awesome";
    }
    
    @Override
    public String execute(String input, Object workerSDK) throws Exception {
        Map<String, Object> result = new HashMap<>();
        result.put("message", "Hello from Java!");
        result.put("timestamp", System.currentTimeMillis());
        
        return objectMapper.writeValueAsString(result);
    }
}
```

#### 2. Build & Run

```bash
cd services/java-simple-worker
mvn clean package
java -jar target/java-simple-worker-1.0.0.jar
```

**Xong!** Plugin tự động được load và expose qua Web API.

## 📝 Plugin Examples

### Simple Plugin

```python
class HelloPlugin(BasePlugin):
    @property
    def name(self) -> str:
        return "hello"
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        return {"message": "Hello World!"}
```

### File Upload Plugin

```python
class ImageAnalysisPlugin(BasePlugin):
    @property
    def name(self) -> str:
        return "analyze_image"
    
    @property
    def accepts_file(self) -> bool:
        return True  # Enable file upload
    
    @property
    def file_field_name(self) -> str:
        return "file"  # Form field name
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        filename = params.get('filename')
        size = params.get('size')
        return {
            "filename": filename,
            "size_mb": round(size / (1024 * 1024), 2)
        }
```

### Worker-to-Worker Plugin

```python
class CompositeTaskPlugin(BasePlugin):
    @property
    def name(self) -> str:
        return "composite_task"
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        # Step 1: Local processing
        result = {"step1": "done"}
        
        # Step 2: Call another worker
        if worker_sdk:
            java_response = worker_sdk.call_worker(
                target_worker='java-simple-worker',
                capability='read_file_info',
                params={'filePath': '/etc/hosts'},
                timeout=30
            )
            result["java_info"] = java_response
        
        return result
```

## 🔗 Worker-to-Worker Communication

Plugins có thể gọi capabilities của workers khác:

```python
def execute(self, params: dict, worker_sdk=None) -> dict:
    if worker_sdk:
        # Call Java worker
        response = worker_sdk.call_worker(
            target_worker='java-simple-worker',  # Worker ID
            capability='read_file_info',         # Capability name
            params={'filePath': '/tmp/file.txt'}, # Parameters
            timeout=30                            # Timeout in seconds
        )
        return {"java_result": response}
```

## 🌐 Web API Auto-Routing

Khi plugin được load, Web API tự động tạo endpoints:

### Example: Hello Plugin
```bash
# Plugin auto-registered as:
POST http://localhost:8081/api/call/hello

# Request:
curl -X POST http://localhost:8081/api/call/hello \
  -H "Content-Type: application/json" \
  -d '{}'

# Response:
{
  "from": "python-worker",
  "response": "{\"message\":\"Hello World!\"}",
  "status": "success"
}
```

### Example: File Upload Plugin
```bash
POST http://localhost:8081/api/call/analyze_image

# With file upload:
curl -X POST http://localhost:8081/api/call/analyze_image \
  -F "file=@image.jpg"
```

## 📦 Plugin Properties

### Required Properties

```python
@property
def name(self) -> str:
    return "capability_name"  # REQUIRED: Unique capability name

def execute(self, params: dict, worker_sdk=None) -> dict:
    return {"result": "data"}  # REQUIRED: Execute logic
```

### Optional Properties

```python
@property
def description(self) -> str:
    return "What this plugin does"  # Human-readable description

@property
def input_schema(self) -> str:
    return '{"type":"object",...}'  # JSON Schema for validation

@property
def output_schema(self) -> str:
    return '{"type":"object",...}'  # JSON Schema for output

@property
def http_method(self) -> str:
    return "POST"  # HTTP method: GET/POST/PUT/DELETE

@property
def accepts_file(self) -> bool:
    return False  # Enable file upload

@property
def file_field_name(self) -> str:
    return "file"  # Form field name for file upload
```

### Lifecycle Hooks

```python
def on_load(self):
    """Called when plugin is loaded"""
    print(f"Plugin {self.name} loaded!")
    # Initialize resources, connections, etc.

def on_unload(self):
    """Called when plugin is unloaded"""
    print(f"Plugin {self.name} unloaded!")
    # Cleanup resources, close connections, etc.
```

## 🎨 Plugin Naming Convention

### Python
- File name: `*_plugin.py` (e.g., `hello_plugin.py`)
- Class name: `*Plugin` (e.g., `HelloPlugin`)
- Must inherit from `BasePlugin`

### Java
- File name: `*Plugin.java` (e.g., `HelloWorldPlugin.java`)
- Class name: `*Plugin` (e.g., `HelloWorldPlugin`)
- Must implement `BasePlugin` interface
- Must be in `com.deepapp.worker.plugins` package

## 🔥 Hot Reload (Development)

Để thêm plugin mới khi worker đang chạy:

1. Tạo file plugin mới
2. Restart worker (auto-reload coming soon)

```bash
# Python
pkill -f worker_plugin_system.py
python3 worker_plugin_system.py

# Java (with Maven)
mvn clean package && java -jar target/*.jar
```

## 📊 Monitoring

Check loaded plugins:

```bash
# Get all capabilities
curl http://localhost:8081/api/capabilities

# Response shows all auto-discovered plugins:
{
  "capabilities": {
    "hello": {...},
    "analyze_image": {...},
    "composite_task": {...}
  }
}
```

## 🐛 Debugging

### Plugin not loading?

**Python:**
```bash
# Check worker logs
tail -f /var/log/supervisor/python-worker.out.log

# Look for:
🔌 Auto-discovering plugins from: /app/python-worker/plugins
📦 Found X plugin modules
✓ Loaded plugin: HelloPlugin → capability 'hello'
```

**Java:**
```bash
# Check worker logs
tail -f /var/log/supervisor/java-worker.out.log

# Look for:
🔌 Auto-discovering plugins from package: com.deepapp.worker.plugins
📦 Found X plugin classes
✓ Loaded plugin: HelloWorldPlugin → capability 'hello_world'
```

### Common Issues

1. **File naming**: Must end with `_plugin.py` (Python) or `Plugin.java` (Java)
2. **Class naming**: Must inherit/implement `BasePlugin`
3. **Package**: Java plugins must be in `com.deepapp.worker.plugins`
4. **Syntax errors**: Check logs for exceptions

## 🚢 Docker Deployment

Plugin system works seamlessly in Docker:

```bash
# Build with plugin system
docker-compose -f docker-compose.all-in-one.yml build

# Run
docker-compose -f docker-compose.all-in-one.yml up -d

# Logs
docker-compose -f docker-compose.all-in-one.yml logs -f python-worker
```

Plugins are automatically copied and loaded in the container!

## 🎯 Best Practices

1. **One Plugin, One Capability**: Mỗi plugin nên implement 1 capability
2. **Error Handling**: Always handle exceptions trong `execute()`
3. **Timeouts**: Set reasonable timeouts cho worker-to-worker calls
4. **Validation**: Use input_schema để validate parameters
5. **Testing**: Test plugin locally trước khi deploy
6. **Logging**: Use print() hoặc logger để debug
7. **Cleanup**: Implement `on_unload()` để cleanup resources

## 📚 Advanced Topics

### Custom Worker SDK

```python
class MyWorkerSDK:
    def call_worker(self, target_worker, capability, params, timeout):
        # Custom implementation
        pass

# Pass to plugin
result = plugin.execute(params, worker_sdk=my_sdk)
```

### Async Plugins (Coming Soon)

```python
async def execute(self, params, worker_sdk=None):
    result = await some_async_operation()
    return result
```

### Plugin Dependencies (Coming Soon)

```python
class MyPlugin(BasePlugin):
    @property
    def dependencies(self):
        return ['other_plugin']  # Ensure load order
```

## 🤝 Contributing

Để contribute plugins mới:

1. Fork repo
2. Tạo plugin trong `plugins/`
3. Test locally
4. Submit PR

## 📄 License

MIT License - see LICENSE file

## 🆘 Support

- 📧 Email: support@example.com
- 💬 Discord: [Join our server](#)
- 📖 Docs: [Full documentation](#)

---

**Happy Plugin Development! 🎉**
