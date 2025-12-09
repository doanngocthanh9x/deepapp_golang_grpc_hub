# DeepApp gRPC Hub - Worker SDKs

This repository contains SDKs for creating workers that connect to the DeepApp gRPC Hub system. The Hub provides a centralized way to manage and orchestrate microservices with automatic API generation.

## Available SDKs

### 🚀 [Node.js SDK](./nodejs/)
- **Language**: JavaScript/TypeScript
- **Protocol**: gRPC with protobuf
- **Features**: Async/await support, automatic reconnection
- **Installation**: `npm install @grpc/grpc-js @grpc/proto-loader`

### ☕ [Java SDK](./java/)
- **Language**: Java 11+
- **Protocol**: gRPC with OkHttp transport
- **Features**: Annotation-based handlers, automatic JSON serialization
- **Installation**: Maven dependencies (gRPC, Jackson, SLF4J)

### 🐍 [Python SDK](./python/)
- **Language**: Python 3.8+
- **Protocol**: gRPC with protobuf
- **Features**: Threading support, automatic proto generation
- **Installation**: `pip install grpcio grpcio-tools`

### 🐹 [Go SDK](./go/)
- **Language**: Go 1.18+
- **Protocol**: gRPC with protobuf
- **Features**: Context support, structured error handling
- **Installation**: `go get google.golang.org/grpc`

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web API       │    │     Hub         │    │   Workers       │
│   (Go)          │◄──►│   (Go/gRPC)     │◄──►│   (Any Lang)    │
│                 │    │                 │    │                 │
│ • REST API      │    │ • Message       │    │ • Capabilities  │
│ • Swagger UI    │    │   Routing       │    │ • File Upload   │
│ • Dynamic       │    │ • Worker Reg    │    │ • Auto Discovery│
│   Endpoints     │    │ • Load Balance  │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Worker Capabilities

Workers register capabilities with the Hub, which automatically generates:

- **REST API endpoints** (`/api/call/{capability}`)
- **Swagger documentation** with proper schemas
- **File upload support** for capabilities that need it
- **Request routing** based on capability names

### Capability Definition

```json
{
  "name": "analyze_image",
  "description": "Analyze uploaded image",
  "input_schema": "{\"type\":\"object\",\"properties\":{\"file\":{\"type\":\"string\",\"format\":\"binary\"}}}",
  "output_schema": "{\"type\":\"object\",\"properties\":{\"result\":{\"type\":\"string\"}}}",
  "http_method": "POST",
  "accepts_file": true,
  "file_field_name": "file"
}
```

## Getting Started

1. **Choose your language** and navigate to the corresponding SDK directory
2. **Install dependencies** as described in the README
3. **Copy the example** and modify it for your use case
4. **Set environment variables**:
   ```bash
   WORKER_ID=my-custom-worker
   HUB_ADDRESS=localhost:50051
   ```
5. **Run your worker** and it will automatically register with the Hub

## File Upload Support

Workers can declare capabilities that accept file uploads:

- Set `accepts_file: true`
- Specify `file_field_name` (e.g., "file", "image", "document")
- Files are base64 encoded in the request
- Hub generates multipart/form-data endpoints automatically

## Message Flow

```
1. Worker → Hub: Register capabilities
2. Client → Web API: Call capability via REST
3. Web API → Hub: Route to appropriate worker
4. Hub → Worker: Execute capability
5. Worker → Hub: Return result
6. Hub → Web API → Client: Response
```

## Development

### Running the Hub

```bash
# Start all-in-one container
cd /path/to/deepapp_golang_grpc_hub
docker-compose -f docker-compose.all-in-one.yml up
```

### Testing Workers

```bash
# Check capabilities
curl http://localhost:8080/api/capabilities

# Call a capability
curl -X POST http://localhost:8080/api/call/hello

# View Swagger UI
open http://localhost:8080/api/docs
```

## Examples

Each SDK includes complete examples demonstrating:

- Basic capability registration
- File upload handling
- Error handling
- Environment variable configuration
- Graceful shutdown

## Contributing

1. Choose an SDK directory
2. Follow the existing patterns
3. Add comprehensive examples
4. Update documentation
5. Test with the Hub system

## License

MIT License - see individual SDK directories for details.