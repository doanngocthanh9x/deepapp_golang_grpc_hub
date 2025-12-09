# 🔌 Python Worker Plugin System

## Overview

Worker tự động load capabilities từ thư mục `plugins/` sử dụng decorator `@capability`.

## 📁 Cấu trúc

```
python-worker-v2/
├── decorators.py           # @capability decorator
├── plugin_loader.py        # Auto-load plugins
├── worker_plugin.py        # Main worker với plugin system
└── plugins/                # Plugin directory
    ├── text_handlers.py    # Text processing plugins
    ├── media_handlers.py   # Image/video plugins
    └── math_handlers.py    # Math calculation plugins
```

## 🎯 Cách tạo plugin mới

### 1. Tạo file trong `plugins/`

```bash
touch plugins/my_plugin.py
```

### 2. Viết handler với decorator

```python
"""
My custom plugin
"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from decorators import capability


@capability(
    name="my_feature",
    description="My awesome feature",
    input_schema={
        "type": "object",
        "properties": {
            "input": {"type": "string"}
        },
        "required": ["input"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "output": {"type": "string"}
        }
    }
)
def handle_my_feature(worker_context, payload):
    """Process the request"""
    input_data = payload.get("input", "")
    worker_id = worker_context.get("worker_id")
    
    return {
        "output": f"Processed: {input_data}",
        "processed_by": worker_id
    }
```

### 3. Restart worker - Done!

```bash
python worker_plugin.py
```

Worker tự động load plugin và register capability!

## 📦 Built-in Plugins

### text_handlers.py
- **hello**: Simple greeting
- **echo**: Echo back message
- **text_transform**: Transform text (uppercase, lowercase, title, reverse, count)

### media_handlers.py
- **image_analysis**: Analyze image metadata
- **video_metadata**: Extract video metadata

### math_handlers.py
- **calculator**: Math operations (add, subtract, multiply, divide, power, sqrt)
- **statistics**: Calculate stats (count, sum, mean, min, max, median)

## 🚀 Run Worker

```bash
# Default
python worker_plugin.py

# Custom configuration
HUB_ADDRESS=localhost:50052 \
WORKER_ID=my-worker \
PLUGINS_DIR=plugins \
python worker_plugin.py
```

## 🧪 Test Plugins

```bash
# Hello plugin
curl -X POST http://localhost:8082/api/v2/invoke/hello \
  -H "Content-Type: application/json" \
  -d '{"name":"Plugin System"}'

# Text transform
curl -X POST http://localhost:8082/api/v2/invoke/text_transform \
  -H "Content-Type: application/json" \
  -d '{"text":"hello world","operation":"uppercase"}'

# Calculator
curl -X POST http://localhost:8082/api/v2/invoke/calculator \
  -H "Content-Type: application/json" \
  -d '{"operation":"add","a":10,"b":5}'

# Statistics
curl -X POST http://localhost:8082/api/v2/invoke/statistics \
  -H "Content-Type: application/json" \
  -d '{"numbers":[1,2,3,4,5,10,20,30]}'
```

## 🎨 Decorator API

### @capability(name, description, input_schema, output_schema)

**Parameters:**
- `name` (str): Unique capability name
- `description` (str): Human-readable description
- `input_schema` (dict): JSON Schema for input validation
- `output_schema` (dict): JSON Schema for output format

**Handler signature:**
```python
def handler(worker_context: dict, payload: dict) -> dict:
    pass
```

**worker_context** contains:
- `worker_id`: Worker identifier
- `worker_type`: Worker type
- `version`: Worker version

## 📊 Benefits

### ✅ No Manual Registration
```python
# OLD WAY (worker_dynamic.py)
self.capabilities = {
    "hello": {...},
    "feature1": {...},
    "feature2": {...}
}

# NEW WAY (plugin system)
# Just add @capability decorator - automatic registration!
```

### ✅ Modular & Clean
- Mỗi plugin là một file riêng
- Dễ maintain và test
- Có thể enable/disable bằng cách rename file

### ✅ Zero Configuration
- Drop file vào `plugins/`
- Worker tự động discover
- API tự động expose endpoint

### ✅ Team Collaboration
- Developer A: `plugins/auth_handlers.py`
- Developer B: `plugins/payment_handlers.py`
- Developer C: `plugins/ai_handlers.py`
- Không conflict!

## 🔥 Advanced Usage

### Multiple handlers in one plugin

```python
@capability(name="func1", ...)
def handler1(ctx, payload):
    pass

@capability(name="func2", ...)
def handler2(ctx, payload):
    pass

@capability(name="func3", ...)
def handler3(ctx, payload):
    pass
```

### Shared utilities

```python
# plugins/utils.py (không có decorator → không được register)
def validate_email(email):
    return "@" in email

# plugins/user_handlers.py
from utils import validate_email

@capability(name="create_user", ...)
def handle_create_user(ctx, payload):
    email = payload.get("email")
    if not validate_email(email):
        return {"error": "Invalid email"}
    # ...
```

### Conditional loading

```python
# plugins/ai_handlers.py
import os

if os.getenv("ENABLE_AI") == "true":
    @capability(name="ai_analyze", ...)
    def handle_ai(ctx, payload):
        pass
```

## 🎓 Comparison

| Aspect | Manual (v1) | Plugin System (v2) |
|--------|-------------|-------------------|
| Add feature | Edit worker code | Create plugin file |
| Registration | Manual dict | Auto via decorator |
| Organization | Single file | Multiple plugin files |
| Team work | Merge conflicts | Independent files |
| Testing | Full worker | Individual plugin |
| Maintenance | Hard to scale | Clean & modular |

🚀 **Plugin system = Scalable, maintainable, team-friendly!**
