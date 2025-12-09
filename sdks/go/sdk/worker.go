// Package sdk provides a simple SDK for creating workers in the DeepApp gRPC Hub system
package sdk

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    hubpb "deepapp/sdk/proto"
)

// Config holds worker configuration
type Config struct {
    WorkerID   string
    HubAddress string
}

// Capability defines a worker capability
type Capability struct {
    Name          string
    Description   string
    InputSchema   string
    OutputSchema  string
    HTTPMethod    string
    AcceptsFile   bool
    FileFieldName string
}

// Message represents a message from the hub
type Message struct {
    ID        string
    From      string
    To        string
    Channel   string
    Content   string
    Timestamp string
    Type      hubpb.MessageType
    Action    string
    Metadata  map[string]string
}

// Response represents a response to send back
type Response struct {
    Data map[string]interface{} `json:"data"`
}

// CapabilityHandler is the interface for handling capabilities
type CapabilityHandler interface {
    Handle(ctx context.Context, msg *Message) (*Response, error)
}

// Worker represents a worker instance
type Worker struct {
    config       Config
    capabilities []Capability
    handlers     map[string]CapabilityHandler
    conn         *grpc.ClientConn
    client       hubpb.HubServiceClient
    running      bool
    mu           sync.RWMutex
}

// NewWorker creates a new worker instance
func NewWorker(config Config) *Worker {
    if config.WorkerID == "" {
        config.WorkerID = fmt.Sprintf("go-worker-%d", time.Now().Unix())
    }
    if config.HubAddress == "" {
        config.HubAddress = "localhost:50051"
    }

    return &Worker{
        config:   config,
        handlers: make(map[string]CapabilityHandler),
        running:  false,
    }
}

// AddCapability adds a capability to the worker
func (w *Worker) AddCapability(cap Capability) {
    w.capabilities = append(w.capabilities, cap)
}

// SetHandler sets a handler for a capability
func (w *Worker) SetHandler(capabilityName string, handler CapabilityHandler) {
    w.handlers[capabilityName] = handler
}

// Start connects to the hub and starts processing messages
func (w *Worker) Start() error {
    log.Printf("🚀 Starting Go Worker: %s", w.config.WorkerID)
    log.Printf("📡 Connecting to Hub at: %s", w.config.HubAddress)

    // Create gRPC connection
    conn, err := grpc.Dial(w.config.HubAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return fmt.Errorf("failed to connect to hub: %w", err)
    }
    w.conn = conn
    w.client = hubpb.NewHubServiceClient(conn)

    log.Println("✓ Connected to Hub")

    // Start bidirectional stream
    stream, err := w.client.Connect(context.Background())
    if err != nil {
        return fmt.Errorf("failed to create stream: %w", err)
    }

    // Send registration
    if err := w.sendRegistration(stream); err != nil {
        return fmt.Errorf("failed to register: %w", err)
    }

    log.Printf("✓ Registered with Hub as '%s'", w.config.WorkerID)
    log.Println("📨 Listening for requests...\n")

    w.running = true

    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Println("\n🛑 Received shutdown signal")
        w.Stop()
    }()

    // Start message processing loop
    return w.processMessages(stream)
}

// Stop stops the worker
func (w *Worker) Stop() {
    w.mu.Lock()
    defer w.mu.Unlock()

    if !w.running {
        return
    }

    log.Println("🛑 Stopping worker...")
    w.running = false

    if w.conn != nil {
        w.conn.Close()
        log.Println("✓ Disconnected from Hub")
    }
}

// IsRunning returns whether the worker is running
func (w *Worker) IsRunning() bool {
    w.mu.RLock()
    defer w.mu.RUnlock()
    return w.running
}

// WorkerID returns the worker ID
func (w *Worker) WorkerID() string {
    return w.config.WorkerID
}

func (w *Worker) sendRegistration(stream hubpb.HubService_ConnectClient) error {
    // Convert capabilities to JSON format
    caps := make([]map[string]interface{}, len(w.capabilities))
    for i, cap := range w.capabilities {
        capMap := map[string]interface{}{
            "name":         cap.Name,
            "description":  cap.Description,
            "input_schema": cap.InputSchema,
            "output_schema": cap.OutputSchema,
            "http_method":  cap.HTTPMethod,
            "accepts_file": cap.AcceptsFile,
        }
        if cap.FileFieldName != "" {
            capMap["file_field_name"] = cap.FileFieldName
        }
        caps[i] = capMap
    }

    registrationData := map[string]interface{}{
        "worker_id":   w.config.WorkerID,
        "worker_type": "go-sdk",
        "capabilities": caps,
        "metadata": map[string]interface{}{
            "version":     "1.0.0",
            "description": "Go SDK Worker",
            "sdk":         "go-sdk",
        },
    }

    content, err := json.Marshal(registrationData)
    if err != nil {
        return err
    }

    msg := &hubpb.Message{
        Id:        fmt.Sprintf("register-%d", time.Now().UnixNano()),
        From:      w.config.WorkerID,
        To:        "hub",
        Channel:   "system",
        Content:   string(content),
        Timestamp: time.Now().Format(time.RFC3339),
        Type:      hubpb.MessageType_REGISTER,
        Action:    "register",
    }

    return stream.Send(msg)
}

func (w *Worker) processMessages(stream hubpb.HubService_ConnectClient) error {
    for w.IsRunning() {
        msg, err := stream.Recv()
        if err != nil {
            if !w.IsRunning() {
                break
            }
            return fmt.Errorf("stream receive error: %w", err)
        }

        // Convert protobuf message to SDK message
        sdkMsg := &Message{
            ID:        msg.Id,
            From:      msg.From,
            To:        msg.To,
            Channel:   msg.Channel,
            Content:   msg.Content,
            Timestamp: msg.Timestamp,
            Type:      msg.Type,
            Action:    msg.Action,
            Metadata:  msg.Metadata,
        }

        // Process message in goroutine
        go w.handleMessage(sdkMsg, stream)
    }
    return nil
}

func (w *Worker) handleMessage(msg *Message, stream hubpb.HubService_ConnectClient) {
    log.Printf("📬 Received request: %s from %s", msg.Channel, msg.From)

    capability := msg.Channel

    // Find handler
    handler, exists := w.handlers[capability]
    if !exists {
        log.Printf("⚠️  No handler for capability: %s", capability)
        w.sendErrorResponse(msg, stream, fmt.Sprintf("Unknown capability: %s", capability))
        return
    }

    // Handle the capability
    ctx := context.Background()
    response, err := handler.Handle(ctx, msg)
    if err != nil {
        log.Printf("❌ Error handling capability %s: %v", capability, err)
        w.sendErrorResponse(msg, stream, err.Error())
        return
    }

    // Send response
    w.sendResponse(msg, stream, response)
}

func (w *Worker) sendResponse(request *Message, stream hubpb.HubService_ConnectClient, response *Response) {
    responseData := map[string]interface{}{
        "status": "success",
    }

    // Merge response data
    for k, v := range response.Data {
        responseData[k] = v
    }

    content, err := json.Marshal(responseData)
    if err != nil {
        log.Printf("❌ Failed to marshal response: %v", err)
        return
    }

    msg := &hubpb.Message{
        Id:        fmt.Sprintf("resp-%d", time.Now().UnixNano()),
        From:      w.config.WorkerID,
        To:        request.From,
        Channel:   request.Channel,
        Content:   string(content),
        Timestamp: time.Now().Format(time.RFC3339),
        Type:      hubpb.MessageType_DIRECT,
        Action:    "response",
    }

    if err := stream.Send(msg); err != nil {
        log.Printf("❌ Failed to send response: %v", err)
    } else {
        log.Println("✓ Sent response")
    }
}

func (w *Worker) sendErrorResponse(request *Message, stream hubpb.HubService_ConnectClient, errorMsg string) {
    responseData := map[string]interface{}{
        "error":  errorMsg,
        "status": "failed",
    }

    content, _ := json.Marshal(responseData)

    msg := &hubpb.Message{
        Id:        fmt.Sprintf("resp-%d", time.Now().UnixNano()),
        From:      w.config.WorkerID,
        To:        request.From,
        Channel:   request.Channel,
        Content:   string(content),
        Timestamp: time.Now().Format(time.RFC3339),
        Type:      hubpb.MessageType_DIRECT,
        Action:    "response",
    }

    stream.Send(msg)
}