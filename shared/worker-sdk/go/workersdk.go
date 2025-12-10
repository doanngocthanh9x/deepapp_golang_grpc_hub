// Package workersdk provides an easy-to-use SDK for creating gRPC workers
// with built-in worker-to-worker communication support
package workersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "deepapp_golang_grpc_hub/internal/proto"
)

// CapabilityHandler is a function that handles a capability request
type CapabilityHandler func(params map[string]interface{}) (map[string]interface{}, error)

// Capability represents a worker capability
type Capability struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	InputSchema   string `json:"input_schema"`
	OutputSchema  string `json:"output_schema"`
	HTTPMethod    string `json:"http_method"`
	AcceptsFile   bool   `json:"accepts_file"`
	FileFieldName string `json:"file_field_name,omitempty"`
}

// WorkerSDK provides the base SDK for creating workers
type WorkerSDK struct {
	workerID    string
	hubAddress  string
	workerType  string
	running     bool
	stream      pb.HubService_ConnectClient
	sendChan    chan *pb.Message
	
	// Capability registry
	capabilities map[string]*Capability
	handlers     map[string]CapabilityHandler
	
	// Worker-to-worker call tracking
	pendingCalls sync.Map
	mu           sync.RWMutex
}

// PendingCall tracks a pending worker-to-worker call
type PendingCall struct {
	responseChan chan *pb.Message
	timer        *time.Timer
}

// NewWorkerSDK creates a new worker SDK instance
func NewWorkerSDK(workerID, hubAddress, workerType string) *WorkerSDK {
	return &WorkerSDK{
		workerID:     workerID,
		hubAddress:   hubAddress,
		workerType:   workerType,
		sendChan:     make(chan *pb.Message, 100),
		capabilities: make(map[string]*Capability),
		handlers:     make(map[string]CapabilityHandler),
	}
}

// AddCapability registers a new capability handler
func (w *WorkerSDK) AddCapability(cap *Capability, handler CapabilityHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.capabilities[cap.Name] = cap
	w.handlers[cap.Name] = handler
	
	log.Printf("[%s] ✓ Registered capability: %s", w.workerID, cap.Name)
}

// CallWorker calls another worker's capability through the Hub
func (w *WorkerSDK) CallWorker(targetWorker, capability string, params map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	if !w.running {
		return nil, fmt.Errorf("worker not connected")
	}
	
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	log.Printf("[%s] 🔗 Calling %s.%s", w.workerID, targetWorker, capability)
	
	// Serialize params
	content, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	
	// Create worker call message
	callMsg := &pb.Message{
		Id:        requestID,
		From:      w.workerID,
		To:        targetWorker,
		Channel:   capability,
		Content:   string(content),
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_WORKER_CALL,
		Metadata:  map[string]string{"capability": capability},
	}
	
	// Create response channel
	responseChan := make(chan *pb.Message, 1)
	timer := time.NewTimer(timeout)
	
	// Register pending call
	w.pendingCalls.Store(requestID, &PendingCall{
		responseChan: responseChan,
		timer:        timer,
	})
	
	// Send the call
	w.sendChan <- callMsg
	
	// Wait for response or timeout
	select {
	case response := <-responseChan:
		// Parse response
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return result, nil
		
	case <-timer.C:
		// Timeout - cleanup
		w.pendingCalls.Delete(requestID)
		return nil, fmt.Errorf("no response from %s after %v", targetWorker, timeout)
	}
}

// handleWorkerCallResponse handles response from worker-to-worker call
func (w *WorkerSDK) handleWorkerCallResponse(msg *pb.Message) {
	requestID, ok := msg.Metadata["request_id"]
	if !ok {
		return
	}
	
	if val, ok := w.pendingCalls.Load(requestID); ok {
		pending := val.(*PendingCall)
		pending.timer.Stop()
		pending.responseChan <- msg
		w.pendingCalls.Delete(requestID)
	}
}

// processMessage processes an incoming message
func (w *WorkerSDK) processMessage(msg *pb.Message) (string, error) {
	w.mu.RLock()
	handler, ok := w.handlers[msg.Channel]
	w.mu.RUnlock()
	
	if !ok {
		return "", fmt.Errorf("unknown capability: %s", msg.Channel)
	}
	
	// Parse input
	var params map[string]interface{}
	if msg.Content != "" {
		if err := json.Unmarshal([]byte(msg.Content), &params); err != nil {
			return "", fmt.Errorf("failed to parse params: %w", err)
		}
	}
	
	// Call handler
	result, err := handler(params)
	if err != nil {
		return "", err
	}
	
	// Serialize result
	content, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return string(content), nil
}

// sendRegistration sends registration message to Hub
func (w *WorkerSDK) sendRegistration() error {
	w.mu.RLock()
	capabilities := make([]*Capability, 0, len(w.capabilities))
	for _, cap := range w.capabilities {
		capabilities = append(capabilities, cap)
	}
	w.mu.RUnlock()
	
	regData := map[string]interface{}{
		"worker_id":   w.workerID,
		"worker_type": w.workerType,
		"capabilities": capabilities,
		"metadata": map[string]string{
			"version":     "1.0.0",
			"sdk_version": "2.0.0",
		},
	}
	
	content, err := json.Marshal(regData)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}
	
	regMsg := &pb.Message{
		Id:        fmt.Sprintf("register-%d", time.Now().UnixNano()),
		From:      w.workerID,
		To:        "hub",
		Channel:   "system",
		Content:   string(content),
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_REGISTER,
		Metadata:  make(map[string]string),
	}
	
	w.sendChan <- regMsg
	log.Printf("[%s] 📤 Sent registration", w.workerID)
	
	return nil
}

// receiveLoop handles incoming messages from Hub
func (w *WorkerSDK) receiveLoop() {
	for {
		if !w.running {
			break
		}
		
		msg, err := w.stream.Recv()
		if err != nil {
			log.Printf("[%s] ✗ Receive error: %v", w.workerID, err)
			w.running = false
			break
		}
		
		// Handle different message types
		switch msg.Type {
		case pb.MessageType_RESPONSE:
			// Response from worker-to-worker call
			w.handleWorkerCallResponse(msg)
			
		case pb.MessageType_WORKER_CALL, pb.MessageType_REQUEST:
			// Process and send response
			content, err := w.processMessage(msg)
			if err != nil {
				content = fmt.Sprintf(`{"error":"%s","status":"failed"}`, err.Error())
			}
			
			responseMsg := &pb.Message{
				Id:        fmt.Sprintf("resp-%d", time.Now().UnixNano()),
				From:      w.workerID,
				To:        msg.From,
				Channel:   msg.Channel,
				Content:   content,
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      pb.MessageType_RESPONSE,
				Metadata:  make(map[string]string),
			}
			
			// Add request_id for worker-to-worker calls
			if msg.Type == pb.MessageType_WORKER_CALL {
				responseMsg.Metadata["request_id"] = msg.Id
				responseMsg.Metadata["status"] = "success"
			}
			
			w.sendChan <- responseMsg
		}
	}
	
	log.Printf("[%s] Receive loop exited", w.workerID)
}

// sendLoop handles sending messages to Hub
func (w *WorkerSDK) sendLoop() {
	for msg := range w.sendChan {
		if !w.running {
			break
		}
		
		if err := w.stream.Send(msg); err != nil {
			log.Printf("[%s] ✗ Send error: %v", w.workerID, err)
		}
	}
	
	log.Printf("[%s] Send loop exited", w.workerID)
}

// Run starts the worker and connects to Hub
func (w *WorkerSDK) Run() error {
	log.Printf("[%s] 🚀 Starting Worker", w.workerID)
	log.Printf("[%s]    ID: %s", w.workerID, w.workerID)
	log.Printf("[%s]    Hub: %s", w.workerID, w.hubAddress)
	log.Printf("[%s] %s", w.workerID, "==================================================")
	
	log.Printf("[%s] ✓ Registered %d capabilities", w.workerID, len(w.capabilities))
	
	// Connect to Hub
	log.Printf("[%s] Connecting to Hub...", w.workerID)
	
	conn, err := grpc.Dial(w.hubAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()
	
	client := pb.NewHubServiceClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	
	w.stream = stream
	w.running = true
	
	log.Printf("[%s] ✓ Connected to Hub", w.workerID)
	
	// Send registration
	if err := w.sendRegistration(); err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}
	
	log.Printf("[%s] 📨 Listening for requests...\n", w.workerID)
	
	// Start send and receive loops
	go w.sendLoop()
	go w.receiveLoop()
	
	// Keep running
	for w.running {
		time.Sleep(1 * time.Second)
	}
	
	return nil
}

// Stop stops the worker
func (w *WorkerSDK) Stop() {
	w.running = false
	close(w.sendChan)
	log.Printf("[%s] ✗ Disconnected from Hub", w.workerID)
}
