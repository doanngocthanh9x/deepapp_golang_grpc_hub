package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"deepapp_golang_grpc_hub/services/web-api/internal/client"
)

// JavaWorkerHandler handles Java worker endpoints
type JavaWorkerHandler struct {
	hubClient *client.HubClient
}

// NewJavaWorkerHandler creates a new Java worker handler
func NewJavaWorkerHandler(hubClient *client.HubClient) *JavaWorkerHandler {
	return &JavaWorkerHandler{hubClient: hubClient}
}

// HandleHello handles /api/worker/java/hello
func (h *JavaWorkerHandler) HandleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send request to Java simple worker
	response, err := h.hubClient.SendRequest("java-simple-worker", "hello_world", "{}")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"response":  response.Content,
		"from":      response.From,
		"timestamp": response.Timestamp,
	})
}

// HandleFileInfo handles /api/worker/java/file_info
func (h *JavaWorkerHandler) HandleFileInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	filePath, ok := req["filePath"].(string)
	if !ok || filePath == "" {
		http.Error(w, "filePath is required", http.StatusBadRequest)
		return
	}

	// Create request data
	requestData := map[string]string{
		"filePath": filePath,
	}
	requestJSON, _ := json.Marshal(requestData)

	// Send to Java simple worker
	response, err := h.hubClient.SendRequest("java-simple-worker", "read_file_info", string(requestJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"response":  response.Content,
		"from":      response.From,
		"timestamp": response.Timestamp,
	})
}