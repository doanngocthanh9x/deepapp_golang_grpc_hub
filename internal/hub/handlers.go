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
