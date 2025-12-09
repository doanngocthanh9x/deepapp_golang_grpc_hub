# DeepApp gRPC Hub - Go Worker SDK

This SDK provides a simple way to create workers that connect to the DeepApp gRPC Hub system using Go.

## Installation

```bash
go mod init your-worker
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

## Quick Start

```go
package main

import (
    "context"
    "encoding/json"
    "time"

    "deepapp/sdk"
)

type MyWorker struct {
    *sdk.Worker
}

func NewMyWorker() *MyWorker {
    worker := &MyWorker{
        Worker: sdk.NewWorker(sdk.Config{
            WorkerID:   "my-go-worker",
            HubAddress: "localhost:50051",
        }),
    }

    // Register capabilities
    worker.AddCapability(sdk.Capability{
        Name:        "hello",
        Description: "Returns a hello message",
        InputSchema: "{}",
        OutputSchema: `{"type":"object","properties":{"message":{"type":"string"}}}`,
        HTTPMethod:  "GET",
        AcceptsFile: false,
    })

    worker.AddCapability(sdk.Capability{
        Name:        "process_data",
        Description: "Process uploaded data",
        InputSchema:  `{"type":"object","properties":{"file":{"type":"string","format":"binary"}}}`,
        OutputSchema: `{"type":"object","properties":{"result":{"type":"string"}}}`,
        HTTPMethod:  "POST",
        AcceptsFile: true,
        FileFieldName: "file",
    })

    return worker
}

// Implement capability handlers
func (w *MyWorker) HandleHello(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    return &sdk.Response{
        Data: map[string]interface{}{
            "message":    "Hello from Go Worker! 🐹",
            "timestamp":  time.Now().Format(time.RFC3339),
            "worker_id": w.WorkerID(),
        },
    }, nil
}

func (w *MyWorker) HandleProcessData(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    var input map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Content), &input); err != nil {
        return nil, err
    }

    filename := "unknown"
    if f, ok := input["filename"].(string); ok {
        filename = f
    }

    return &sdk.Response{
        Data: map[string]interface{}{
            "filename":  filename,
            "processed": true,
            "result":    "Data processed successfully",
            "timestamp": time.Now().Format(time.RFC3339),
        },
    }, nil
}

func main() {
    worker := NewMyWorker()
    worker.Start()
}
```

## API Reference

### Worker Struct

#### Constructor

```go
func NewWorker(config Config) *Worker
```

#### Config

```go
type Config struct {
    WorkerID   string // Unique worker ID
    HubAddress string // Hub gRPC address
}
```

#### Methods

- `AddCapability(cap Capability)` - Add a capability
- `Start()` - Connect to hub and start processing
- `Stop()` - Disconnect from hub
- `HandleCapability(name string, handler CapabilityHandler)` - Register handler

### Capability Struct

```go
type Capability struct {
    Name          string // Unique capability name
    Description   string // Human readable description
    InputSchema   string // JSON Schema for input
    OutputSchema  string // JSON Schema for output
    HTTPMethod    string // HTTP method for web API
    AcceptsFile   bool   // Whether it accepts file uploads
    FileFieldName string // Field name for file uploads
}
```

### Handler Functions

Implement handler functions for each capability:

```go
func (w *MyWorker) HandleCapabilityName(ctx context.Context, msg *Message) (*Response, error) {
    // Your logic here
    return &Response{Data: map[string]interface{}{"result": "success"}}, nil
}
```

## Advanced Usage

### File Upload Handling

```go
func (w *MyWorker) HandleAnalyzeImage(ctx context.Context, msg *Message) (*Response, error) {
    var input map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Content), &input); err != nil {
        return nil, err
    }

    // File data is base64 encoded
    fileData, ok := input["file"].(string)
    if !ok {
        return nil, errors.New("no file data provided")
    }

    // Decode base64
    fileBytes, err := base64.StdEncoding.DecodeString(fileData)
    if err != nil {
        return nil, err
    }

    // Process the file...
    analysis := analyzeImage(fileBytes)

    return &Response{
        Data: map[string]interface{}{
            "analysis":  analysis,
            "timestamp": time.Now().Format(time.RFC3339),
        },
    }, nil
}
```

### Error Handling

```go
func (w *MyWorker) HandleCapability(ctx context.Context, msg *Message) (*Response, error) {
    // Your logic here
    if somethingWrong {
        return nil, errors.New("processing failed")
    }

    return &Response{Data: map[string]interface{}{"success": true}}, nil
}
```

### Environment Variables

```bash
WORKER_ID=my-custom-worker
HUB_ADDRESS=localhost:50051
```

## Complete Example

See `examples/` directory for complete working examples.