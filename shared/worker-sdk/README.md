# Worker SDK - Plugin System for Easy Worker-to-Worker Communication

Hệ thống SDK chung giúp các worker kết nối và truyền dữ liệu với nhau một cách dễ dàng qua gRPC Hub.

## 🎯 Tính năng

- ✅ **Đơn giản hóa việc tạo worker**: Chỉ cần extend base class và implement handlers
- ✅ **Worker-to-worker communication**: Gọi capability của worker khác qua `callWorker()` 
- ✅ **Tự động đăng ký với Hub**: SDK tự động gửi registration và xử lý connection
- ✅ **Thread-safe / Goroutine-safe**: Xử lý concurrent requests an toàn
- ✅ **Timeout và error handling**: Tự động xử lý timeout và lỗi
- ✅ **Multi-language support**: Python, Node.js, Go (và có thể mở rộng thêm)

## 📁 Cấu trúc

```
shared/worker-sdk/
├── python/
│   └── worker_sdk.py          # Python SDK base class
├── nodejs/
│   ├── package.json
│   └── index.js               # Node.js SDK base class  
└── go/
    └── workersdk.go           # Go SDK package
```

## 🐍 Python SDK

### Installation

```bash
# Copy SDK file to your project or add to PYTHONPATH
cp shared/worker-sdk/python/worker_sdk.py services/your-worker/

# Generate proto files
cd services/your-worker
python3 -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. ../../proto/hub.proto
```

### Usage

```python
from worker_sdk import WorkerSDK

class MyWorker(WorkerSDK):
    def register_capabilities(self):
        # Register your capabilities
        self.add_capability(
            name="my_task",
            handler=self.handle_my_task,
            description="Does something useful",
            http_method="POST"
        )
    
    def handle_my_task(self, params: dict) -> dict:
        # Call another worker if needed
        result = self.call_worker(
            target_worker="other-worker",
            capability="other_task",
            params={"key": "value"},
            timeout=30
        )
        
        return {
            "status": "success",
            "result": result
        }

# Run worker
worker = MyWorker("my-worker", "localhost:50051")
worker.run()
```

## 🟢 Node.js SDK

### Installation

```bash
cd services/your-worker
npm install ../../../shared/worker-sdk/nodejs
# or
npm install @grpc/grpc-js @grpc/proto-loader uuid
```

### Usage

```javascript
const { WorkerSDK } = require('@deepapp/worker-sdk-nodejs');

class MyWorker extends WorkerSDK {
  registerCapabilities() {
    this.addCapability('my_task', this.handleMyTask.bind(this), {
      description: 'Does something useful',
      httpMethod: 'POST'
    });
  }
  
  async handleMyTask(params) {
    // Call another worker if needed
    const result = await this.callWorker(
      'other-worker',
      'other_task',
      { key: 'value' },
      30000
    );
    
    return {
      status: 'success',
      result
    };
  }
}

// Run worker
const worker = new MyWorker('my-worker', 'localhost:50051');
worker.run();
```

## 🔵 Go SDK

### Installation

```bash
# SDK is already in shared/worker-sdk/go/
# Just import it in your worker
```

### Usage

```go
package main

import (
    "time"
    workersdk "deepapp_golang_grpc_hub/shared/worker-sdk/go"
)

type MyWorker struct {
    sdk *workersdk.WorkerSDK
}

func NewMyWorker(workerID, hubAddress string) *MyWorker {
    sdk := workersdk.NewWorkerSDK(workerID, hubAddress, "golang")
    worker := &MyWorker{sdk: sdk}
    worker.registerCapabilities()
    return worker
}

func (w *MyWorker) registerCapabilities() {
    w.sdk.AddCapability(&workersdk.Capability{
        Name:        "my_task",
        Description: "Does something useful",
        HTTPMethod:  "POST",
    }, w.handleMyTask)
}

func (w *MyWorker) handleMyTask(params map[string]interface{}) (map[string]interface{}, error) {
    // Call another worker if needed
    result, err := w.sdk.CallWorker(
        "other-worker",
        "other_task",
        map[string]interface{}{"key": "value"},
        30*time.Second,
    )
    
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "status": "success",
        "result": result,
    }, nil
}

func main() {
    worker := NewMyWorker("my-worker", "localhost:50051")
    worker.sdk.Run()
}
```

## 🚀 Examples

Xem các example workers:

- **Python**: `services/python-worker/worker_sdk_example.py`
- **Node.js**: `services/node-worker/worker.js`
- **Go**: `services/go-worker/main.go`

## 📝 API Reference

### Base Methods (All Languages)

#### `addCapability(name, handler, options)`
Đăng ký một capability mới

**Parameters:**
- `name`: Tên capability (e.g., "process_data")
- `handler`: Function xử lý (params -> result)
- `options`: Configuration (description, httpMethod, acceptsFile, etc.)

#### `callWorker(targetWorker, capability, params, timeout)`
Gọi capability của worker khác

**Parameters:**
- `targetWorker`: ID của worker cần gọi
- `capability`: Tên capability trên worker đó
- `params`: Parameters để gửi
- `timeout`: Timeout (seconds/milliseconds)

**Returns:** Response từ worker khác

**Throws:** TimeoutError nếu không có response

#### `run()`
Khởi động worker và kết nối với Hub

#### `stop()`
Dừng worker và ngắt kết nối

## 🔄 Worker-to-Worker Communication Flow

```
┌─────────────┐                ┌──────┐                ┌─────────────┐
│  Worker A   │───WORKER_CALL─→│ Hub  │───WORKER_CALL─→│  Worker B   │
│  (Python)   │                │      │                │   (Java)    │
└─────────────┘                └──────┘                └─────────────┘
       ↑                           │                           │
       │                           │                           │
       └──────────RESPONSE─────────┴──────────RESPONSE─────────┘
```

1. Worker A gọi `callWorker("worker-b", "task", params)`
2. SDK tạo WORKER_CALL message và gửi đến Hub
3. Hub route message đến Worker B
4. Worker B xử lý và trả RESPONSE
5. Hub route response về Worker A
6. SDK của Worker A nhận response và trả về cho caller

## 🛠️ Testing

Test worker-to-worker communication:

```bash
# Start Hub
cd cmd/hub && go run main.go

# Terminal 1: Start Python worker
cd services/python-worker
WORKER_ID=python-worker python3 worker_sdk_example.py

# Terminal 2: Start Java worker  
cd services/java-simple-worker
WORKER_ID=java-simple-worker mvn exec:java

# Terminal 3: Test composite task (Python calls Java)
curl -X POST http://localhost:8081/api/call/composite_task \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/etc/hosts"}'
```

## 📚 Advanced Usage

### Custom Error Handling

```python
def handle_task(self, params: dict) -> dict:
    try:
        result = self.call_worker("other-worker", "task", params)
        return {"status": "success", "result": result}
    except TimeoutError as e:
        return {"status": "timeout", "error": str(e)}
    except Exception as e:
        return {"status": "error", "error": str(e)}
```

### Chaining Multiple Workers

```javascript
async handleChainedTask(params) {
  // Step 1: Python processing
  const pythonResult = await this.callWorker(
    'python-worker', 'process', params
  );
  
  // Step 2: Java processing
  const javaResult = await this.callWorker(
    'java-worker', 'analyze', pythonResult
  );
  
  // Step 3: Go processing
  const goResult = await this.callWorker(
    'go-worker', 'finalize', javaResult
  );
  
  return goResult;
}
```

## 🔧 Configuration

Environment variables:

- `WORKER_ID`: Unique worker identifier
- `HUB_ADDRESS`: Hub address (default: localhost:50051)
- `LOG_LEVEL`: Logging level (info, debug, error)

## 📄 License

MIT
