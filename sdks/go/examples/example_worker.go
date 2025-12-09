package main

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "time"

    "deepapp/sdk"
)

type ExampleWorker struct {
    *sdk.Worker
}

func NewExampleWorker() *ExampleWorker {
    workerID := os.Getenv("WORKER_ID")
    if workerID == "" {
        workerID = "go-example-worker"
    }

    hubAddress := os.Getenv("HUB_ADDRESS")
    if hubAddress == "" {
        hubAddress = "localhost:50051"
    }

    worker := &ExampleWorker{
        Worker: sdk.NewWorker(sdk.Config{
            WorkerID:   workerID,
            HubAddress: hubAddress,
        }),
    }

    // Register capabilities
    worker.AddCapability(sdk.Capability{
        Name:        "hello",
        Description: "Returns a hello message",
        InputSchema: "{}",
        OutputSchema: `{"type":"object","properties":{"message":{"type":"string"},"timestamp":{"type":"string"},"worker_id":{"type":"string"},"status":{"type":"string"}}}`,        HTTPMethod:  "GET",
        AcceptsFile: false,
    })

    worker.AddCapability(sdk.Capability{
        Name:        "echo",
        Description: "Echoes back the input message",
        InputSchema: `{"type":"object","properties":{"message":{"type":"string"}}}`,        OutputSchema: `{"type":"object","properties":{"echo":{"type":"string"},"timestamp":{"type":"string"},"status":{"type":"string"}}}`,        HTTPMethod:  "POST",
        AcceptsFile: false,
    })

    worker.AddCapability(sdk.Capability{
        Name:        "reverse_text",
        Description: "Reverses the input text",
        InputSchema: `{"type":"object","properties":{"text":{"type":"string"}}}`,        OutputSchema: `{"type":"object","properties":{"original":{"type":"string"},"reversed":{"type":"string"},"timestamp":{"type":"string"},"status":{"type":"string"}}}`,        HTTPMethod:  "POST",
        AcceptsFile: false,
    })

    worker.AddCapability(sdk.Capability{
        Name:        "analyze_file",
        Description: "Analyze an uploaded file",
        InputSchema: `{"type":"object","properties":{"file":{"type":"string","format":"binary"},"filename":{"type":"string"}}}`,        OutputSchema: `{"type":"object","properties":{"filename":{"type":"string"},"size":{"type":"number"},"analysis":{"type":"object"},"timestamp":{"type":"string"},"status":{"type":"string"}}}`,        HTTPMethod:  "POST",
        AcceptsFile: true,
        FileFieldName: "file",
    })

    // Set handlers
    worker.SetHandler("hello", worker)
    worker.SetHandler("echo", worker)
    worker.SetHandler("reverse_text", worker)
    worker.SetHandler("analyze_file", worker)

    return worker
}

// Implement CapabilityHandler interface
func (w *ExampleWorker) Handle(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    switch msg.Channel {
    case "hello":
        return w.handleHello(ctx, msg)
    case "echo":
        return w.handleEcho(ctx, msg)
    case "reverse_text":
        return w.handleReverseText(ctx, msg)
    case "analyze_file":
        return w.handleAnalyzeFile(ctx, msg)
    default:
        return nil, fmt.Errorf("unknown capability: %s", msg.Channel)
    }
}

func (w *ExampleWorker) handleHello(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    fmt.Println("🔍 Processing hello request")

    return &sdk.Response{
        Data: map[string]interface{}{
            "message":   "Hello World from Go SDK Worker! 🐹",
            "timestamp": time.Now().Format(time.RFC3339),
            "worker_id": w.WorkerID(),
            "status":    "success",
        },
    }, nil
}

func (w *ExampleWorker) handleEcho(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    fmt.Println("🔍 Processing echo request")

    var input map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Content), &input); err != nil {
        return nil, fmt.Errorf("invalid JSON input: %w", err)
    }

    message, ok := input["message"].(string)
    if !ok {
        message = "No message provided"
    }

    return &sdk.Response{
        Data: map[string]interface{}{
            "echo":      message,
            "timestamp": time.Now().Format(time.RFC3339),
            "status":    "success",
        },
    }, nil
}

func (w *ExampleWorker) handleReverseText(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    fmt.Println("🔍 Processing reverse_text request")

    var input map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Content), &input); err != nil {
        return nil, fmt.Errorf("invalid JSON input: %w", err)
    }

    text, ok := input["text"].(string)
    if !ok {
        return nil, fmt.Errorf("text field is required")
    }

    // Reverse the text
    reversed := ""
    for i := len(text) - 1; i >= 0; i-- {
        reversed += string(text[i])
    }

    return &sdk.Response{
        Data: map[string]interface{}{
            "original":  text,
            "reversed":  reversed,
            "timestamp": time.Now().Format(time.RFC3339),
            "status":    "success",
        },
    }, nil
}

func (w *ExampleWorker) handleAnalyzeFile(ctx context.Context, msg *sdk.Message) (*sdk.Response, error) {
    fmt.Println("🔍 Processing file analysis request")

    var input map[string]interface{}
    if err := json.Unmarshal([]byte(msg.Content), &input); err != nil {
        return nil, fmt.Errorf("invalid JSON input: %w", err)
    }

    filename, _ := input["filename"].(string)
    if filename == "" {
        filename = "unknown"
    }

    fileData, ok := input["file"].(string)
    if !ok {
        return nil, errors.New("no file data provided")
    }

    // Decode base64 file data
    fileBytes, err := base64.StdEncoding.DecodeString(fileData)
    if err != nil {
        return nil, fmt.Errorf("invalid base64 file data: %w", err)
    }

    fileSize := len(fileBytes)

    // Basic file analysis
    analysis := map[string]interface{}{
        "size_bytes": fileSize,
        "size_kb":    float64(fileSize) / 1024,
        "size_mb":    float64(fileSize) / (1024 * 1024),
    }

    // Try to detect MIME type from filename
    if filename != "unknown" {
        analysis["filename"] = filename
        // Simple extension-based MIME type detection
        mimeTypes := map[string]string{
            ".txt":  "text/plain",
            ".json": "application/json",
            ".xml":  "application/xml",
            ".html": "text/html",
            ".css":  "text/css",
            ".js":   "application/javascript",
            ".png":  "image/png",
            ".jpg":  "image/jpeg",
            ".jpeg": "image/jpeg",
            ".gif":  "image/gif",
            ".pdf":  "application/pdf",
            ".zip":  "application/zip",
        }

        for ext, mimeType := range mimeTypes {
            if len(filename) > len(ext) && filename[len(filename)-len(ext):] == ext {
                analysis["mime_type"] = mimeType
                break
            }
        }
    }

    if analysis["mime_type"] == nil {
        analysis["mime_type"] = "application/octet-stream"
    }

    // Try to detect if it's text
    isText := true
    for _, b := range fileBytes {
        if b < 32 && b != 9 && b != 10 && b != 13 {
            isText = false
            break
        }
    }
    analysis["is_text"] = isText

    if isText {
        textContent := string(fileBytes)
        analysis["line_count"] = len(fmt.Sprintf("%s", textContent)) // Simple line count
        analysis["char_count"] = len(textContent)
    }

    fmt.Printf("📁 Analyzed file: %s (%d bytes)\n", filename, fileSize)

    return &sdk.Response{
        Data: map[string]interface{}{
            "filename":  filename,
            "size":      fileSize,
            "analysis":  analysis,
            "timestamp": time.Now().Format(time.RFC3339),
            "status":    "success",
        },
    }, nil
}

func main() {
    worker := NewExampleWorker()

    fmt.Printf("🚀 Starting Example Go Worker: %s\n", worker.WorkerID())

    if err := worker.Start(); err != nil {
        fmt.Printf("❌ Failed to start worker: %v\n", err)
        os.Exit(1)
    }
}