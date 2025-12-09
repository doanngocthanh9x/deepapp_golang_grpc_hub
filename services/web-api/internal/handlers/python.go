package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"deepapp_golang_grpc_hub/services/web-api/internal/client"
)

// PythonWorkerHandler handles Python worker endpoints
type PythonWorkerHandler struct {
	hubClient *client.HubClient
}

// NewPythonWorkerHandler creates a new Python worker handler
func NewPythonWorkerHandler(hubClient *client.HubClient) *PythonWorkerHandler {
	return &PythonWorkerHandler{hubClient: hubClient}
}

// HandleHello handles /api/worker/python/hello
func (h *PythonWorkerHandler) HandleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send request to Python worker
	response, err := h.hubClient.SendRequest("python-worker", "hello", "")
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

// HandleAnalyzeImage handles /api/worker/python/analyze_image
func (h *PythonWorkerHandler) HandleAnalyzeImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create request data
	requestData := map[string]interface{}{
		"filename": header.Filename,
		"size":     len(imageData),
		"image":    base64Image[:100] + "...", // Send truncated for demo
	}

	requestJSON, _ := json.Marshal(requestData)

	// Send to Python worker
	response, err := h.hubClient.SendRequest("python-worker", "analyze_image", string(requestJSON))
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