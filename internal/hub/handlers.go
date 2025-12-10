package hub

import (
	"encoding/json"
	"fmt"
	"time"

	"deepapp_golang_grpc_hub/internal/proto"
)

// handleRegistration xử lý worker registration
func (s *Server) handleRegistration(msg *proto.Message) {
	fmt.Printf("📋 Processing registration from %s\n", msg.From)
	fmt.Printf("📄 Registration content: %s\n", msg.Content)

	var regData struct {
		WorkerID     string                   `json:"worker_id"`
		WorkerType   string                   `json:"worker_type"`
		Capabilities []ServiceCapability      `json:"capabilities"`
		Metadata     map[string]interface{}   `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(msg.Content), &regData); err != nil {
		fmt.Printf("❌ Failed to parse registration: %v\n", err)
		return
	}

	fmt.Printf("🔍 Received %d capabilities from %s\n", len(regData.Capabilities), regData.WorkerID)
	for i, cap := range regData.Capabilities {
		fmt.Printf("  Cap %d: %s (http_method=%s, accepts_file=%v, file_field=%s)\n", 
			i, cap.Name, cap.HTTPMethod, cap.AcceptsFile, cap.FileFieldName)
	}

	// Create worker info
	workerInfo := &WorkerInfo{
		ID:           regData.WorkerID,
		Type:         regData.WorkerType,
		Status:       "online",
		Capabilities: regData.Capabilities,
		Metadata:     regData.Metadata,
		RegisteredAt: time.Now().Format(time.RFC3339),
		LastSeen:     time.Now().Format(time.RFC3339),
	}

	// Register with registry
	s.registry.RegisterWorker(regData.WorkerID, workerInfo)

	capNames := make([]string, len(regData.Capabilities))
	for i, cap := range regData.Capabilities {
		capNames[i] = cap.Name
	}

	fmt.Printf("✅ Worker registered: %s [%s] with capabilities: %v\n",
		regData.WorkerID, regData.WorkerType, capNames)

	// Send confirmation back to worker
	confirmMsg := &proto.Message{
		Id:        fmt.Sprintf("confirm-%d", time.Now().UnixNano()),
		From:      "hub",
		To:        msg.From,
		Type:      proto.MessageType_RESPONSE,
		Content:   `{"status":"registered","message":"Worker registered successfully"}`,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.dispatcher.Dispatch(confirmMsg)
}

// handleCapabilityDiscovery xử lý yêu cầu discovery capabilities
func (s *Server) handleCapabilityDiscovery(msg *proto.Message) {
	fmt.Printf("🔍 Processing capability discovery from %s\n", msg.From)

	capabilities := s.registry.GetAllCapabilities()
	workers := s.registry.GetAllWorkers()

	response := map[string]interface{}{
		"capabilities": capabilities,
		"workers":      workers,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	responseJSON, _ := json.Marshal(response)

	responseMsg := &proto.Message{
		Id:        msg.Id,
		From:      "hub",
		To:        msg.From,
		Type:      proto.MessageType_RESPONSE,
		Content:   string(responseJSON),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.dispatcher.Dispatch(responseMsg)
	fmt.Printf("✅ Sent %d capabilities to %s\n", len(capabilities), msg.From)
}

// handleServiceRequest route request to appropriate worker
func (s *Server) handleServiceRequest(msg *proto.Message) {
	fmt.Printf("📨 Processing service request from %s to %s\n", msg.From, msg.To)

	// Check if capability is in metadata
	capability, hasCapability := msg.Metadata["capability"]
	
	// If not in metadata, try parsing from content (backward compatibility)
	if !hasCapability {
		var reqData struct {
			Capability string                 `json:"capability"`
			Payload    map[string]interface{} `json:"payload"`
		}

		if err := json.Unmarshal([]byte(msg.Content), &reqData); err != nil {
			fmt.Printf("❌ Failed to parse request and no capability in metadata: %v\n", err)
			return
		}
		capability = reqData.Capability
	}

	// If To field is already set, route directly
	if msg.To != "" && msg.To != "hub" {
		fmt.Printf("🎯 Routing request to specified worker: %s (capability: %s)\n", msg.To, capability)
		s.dispatcher.Dispatch(msg)
		return
	}

	// Find worker for capability
	workerID, found := s.registry.GetWorkerForCapability(capability)
	if !found {
		fmt.Printf("❌ No worker found for capability: %s\n", capability)
		
		// Send error response
		errorMsg := &proto.Message{
			Id:        msg.Id,
			From:      "hub",
			To:        msg.From,
			Type:      proto.MessageType_RESPONSE,
			Content:   fmt.Sprintf(`{"error":"No worker available for capability: %s"}`, capability),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		s.dispatcher.Dispatch(errorMsg)
		return
	}

	fmt.Printf("🎯 Routing %s request to worker: %s\n", capability, workerID)

	// Route to worker - preserve all message fields
	msg.To = workerID
	s.dispatcher.Dispatch(msg)
}

// handleWorkerCall routes worker-to-worker calls
func (s *Server) handleWorkerCall(msg *proto.Message) {
	fmt.Printf("🔗 Worker-to-Worker call: %s → %s (capability: %s)\n", msg.From, msg.To, msg.Channel)

	// Validate target worker exists
	targetWorker := msg.To
	if targetWorker == "" {
		fmt.Printf("❌ Worker call missing target worker\n")
		s.sendErrorResponse(msg, "Target worker not specified")
		return
	}

	// Check if target worker is registered
	if !s.connMgr.Has(targetWorker) {
		fmt.Printf("❌ Target worker not found: %s\n", targetWorker)
		s.sendErrorResponse(msg, fmt.Sprintf("Worker %s not found or offline", targetWorker))
		return
	}

	// Validate capability
	capability := msg.Channel
	if capability == "" {
		// Try to get from metadata
		if cap, ok := msg.Metadata["capability"]; ok {
			capability = cap
			msg.Channel = cap
		} else {
			fmt.Printf("❌ Worker call missing capability\n")
			s.sendErrorResponse(msg, "Capability not specified")
			return
		}
	}

	// Check if target worker has the capability
	workerForCap, found := s.registry.GetWorkerForCapability(capability)
	if !found || workerForCap != targetWorker {
		fmt.Printf("⚠️  Warning: Worker %s may not have capability %s\n", targetWorker, capability)
	}

	fmt.Printf("✅ Forwarding worker call to %s\n", targetWorker)

	// Forward the message to target worker
	s.dispatcher.Dispatch(msg)
}

// handleResponse routes responses back to original requester
func (s *Server) handleResponse(msg *proto.Message) {
	fmt.Printf("📬 Response: %s → %s\n", msg.From, msg.To)

	// Validate target
	if msg.To == "" {
		fmt.Printf("❌ Response missing target\n")
		return
	}

	// Check if target is connected
	if !s.connMgr.Has(msg.To) {
		fmt.Printf("❌ Response target not connected: %s\n", msg.To)
		return
	}

	// Forward response
	s.dispatcher.Dispatch(msg)
	fmt.Printf("✅ Response delivered to %s\n", msg.To)
}

// sendErrorResponse sends an error response back to requester
func (s *Server) sendErrorResponse(originalMsg *proto.Message, errorMessage string) {
	errorMsg := &proto.Message{
		Id:        fmt.Sprintf("error-%d", time.Now().UnixNano()),
		From:      "hub",
		To:        originalMsg.From,
		Type:      proto.MessageType_RESPONSE,
		Content:   fmt.Sprintf(`{"error":"%s"}`, errorMessage),
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata: map[string]string{
			"original_message_id": originalMsg.Id,
		},
	}
	s.dispatcher.Dispatch(errorMsg)
}
