package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/deepapp/go-worker/plugins"
	pb "github.com/deepapp/go-worker/proto"
)

type GRPCWorker struct {
	workerID   string
	hubAddress string
	stream     pb.HubService_ConnectClient
	conn       *grpc.ClientConn
	plugins    map[string]plugins.Plugin
	connected  bool
}

func NewGRPCWorker(workerID, hubAddress string) *GRPCWorker {
	return &GRPCWorker{
		workerID:   workerID,
		hubAddress: hubAddress,
		plugins:    make(map[string]plugins.Plugin),
		connected:  false,
	}
}

func (w *GRPCWorker) RegisterPlugin(plugin plugins.Plugin) {
	w.plugins[plugin.GetName()] = plugin
	log.Printf("✅ Registered plugin: %s", plugin.GetName())
}

func (w *GRPCWorker) Connect() error {
	log.Printf("🔵 Connecting to Hub at %s...", w.hubAddress)

	conn, err := grpc.Dial(w.hubAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	w.conn = conn

	client := pb.NewHubServiceClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to start stream: %w", err)
	}
	w.stream = stream
	w.connected = true

	log.Println("✅ Connected to Hub")

	// Send registration
	if err := w.sendRegistration(); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	// Start receiving messages
	go w.receiveMessages()

	// Start heartbeat
	go w.startHeartbeat()

	return nil
}

func (w *GRPCWorker) sendRegistration() error {
	capabilities := make([]map[string]interface{}, 0, len(w.plugins))
	
	for _, plugin := range w.plugins {
		cap := map[string]interface{}{
			"name":             plugin.GetName(),
			"description":      plugin.GetDescription(),
			"http_method":      plugin.GetHttpMethod(),
			"accepts_file":     plugin.AcceptsFile(),
			"file_field_name":  plugin.GetFileFieldName(),
		}
		capabilities = append(capabilities, cap)
	}

	regData := map[string]interface{}{
		"worker_id":    w.workerID,
		"worker_type":  "go",
		"capabilities": capabilities,
		"metadata": map[string]interface{}{
			"version":      "1.0.0",
			"description":  "Go worker with plugin system",
			"plugin_count": len(w.plugins),
		},
	}

	content, err := json.Marshal(regData)
	if err != nil {
		return err
	}

	msg := &pb.Message{
		Type:      pb.MessageType_REGISTER,
		From:      w.workerID,
		To:        "hub",
		Content:   string(content),
		Action:    "register",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := w.stream.Send(msg); err != nil {
		return err
	}

	log.Printf("✅ Registration sent with %d capabilities", len(w.plugins))
	return nil
}

func (w *GRPCWorker) receiveMessages() {
	for {
		msg, err := w.stream.Recv()
		if err == io.EOF {
			log.Println("⚠️  Stream ended")
			w.connected = false
			return
		}
		if err != nil {
			log.Printf("❌ Receive error: %v", err)
			w.connected = false
			return
		}

		w.handleMessage(msg)
	}
}

func (w *GRPCWorker) handleMessage(msg *pb.Message) {
	// Convert MessageType to string for comparison
	msgType := msg.Type.String()
	
	switch msgType {
	case "REQUEST":
		w.handleRequest(msg)
	case "RESPONSE":
		w.handleResponse(msg)
	case "REGISTER":
		log.Println("✓ Registration acknowledged")
	default:
		// Silently ignore heartbeats and unknown types
	}
}

func (w *GRPCWorker) handleRequest(msg *pb.Message) {
	requestID := msg.Id
	originalSender := msg.From
	capability := msg.Channel
	if capability == "" && msg.Metadata != nil {
		capability = msg.Metadata["capability"]
	}

	log.Printf("📥 Request: %s (ID: %s)", capability, requestID)

	// Parse parameters from content
	var params map[string]interface{}
	if msg.Content != "" && msg.Content != "{}" {
		if err := json.Unmarshal([]byte(msg.Content), &params); err != nil {
			log.Printf("❌ Failed to parse content: %v", err)
			w.sendErrorResponse(requestID, originalSender, "Invalid request format")
			return
		}
	} else {
		params = make(map[string]interface{})
	}

	// Get plugin
	plugin, exists := w.plugins[capability]
	if !exists {
		log.Printf("❌ Unknown capability: %s", capability)
		w.sendErrorResponse(requestID, originalSender, fmt.Sprintf("Unknown capability: %s", capability))
		return
	}

	// Execute plugin
	ctx := &plugins.ExecutionContext{
		WorkerID:   w.workerID,
		CallWorker: w.callWorker,
	}

	result, err := plugin.Execute(params, ctx)
	if err != nil {
		log.Printf("❌ Plugin execution error: %v", err)
		w.sendErrorResponse(requestID, originalSender, err.Error())
		return
	}

	// Convert result to map[string]interface{}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		// Try to marshal and unmarshal to convert
		data, _ := json.Marshal(result)
		json.Unmarshal(data, &resultMap)
	}

	// Send response
	w.sendResponse(requestID, originalSender, resultMap)
	log.Printf("✅ Response sent for: %s", capability)
}

func (w *GRPCWorker) sendResponse(requestID, targetClient string, result map[string]interface{}) {
	content, err := json.Marshal(result)
	if err != nil {
		log.Printf("❌ Failed to marshal response: %v", err)
		return
	}

	msg := &pb.Message{
		Id:        requestID,
		Type:      pb.MessageType_RESPONSE,
		From:      w.workerID,
		To:        targetClient,
		Content:   string(content),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := w.stream.Send(msg); err != nil {
		log.Printf("❌ Failed to send response: %v", err)
	}
}

func (w *GRPCWorker) sendErrorResponse(requestID, targetClient, errMsg string) {
	result := map[string]interface{}{
		"error":  errMsg,
		"status": "error",
	}
	w.sendResponse(requestID, targetClient, result)
}

func (w *GRPCWorker) handleResponse(msg *pb.Message) {
	// Handle responses from other workers (for worker-to-worker calls)
	// Implementation depends on your needs
	log.Printf("📬 Received response from %s", msg.From)
}

func (w *GRPCWorker) callWorker(targetWorker, capability string, data map[string]interface{}, timeout int) (map[string]interface{}, error) {
	// Implementation for calling other workers
	return nil, fmt.Errorf("worker-to-worker calls not yet implemented")
}

func (w *GRPCWorker) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !w.connected {
			return
		}

		msg := &pb.Message{
			Type:      pb.MessageType_DIRECT,
			From:      w.workerID,
			To:        "",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := w.stream.Send(msg); err != nil {
			log.Printf("❌ Heartbeat failed: %v", err)
			return
		}
	}
}

func (w *GRPCWorker) Close() error {
	w.connected = false
	if w.stream != nil {
		w.stream.CloseSend()
	}
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
