package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// StatusHandler handles status endpoint
type StatusHandler struct{}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{}
}

// HandleStatus handles /api/status
func (h *StatusHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "running",
		"service":     "web-api",
		"hub_address": "localhost:50051",
		"endpoints": []string{
			"/api/worker/python/hello",
			"/api/worker/python/analyze_image",
			"/api/worker/java/hello",
			"/api/worker/java/file_info",
			"/api/status",
			"/api/capabilities",
			"/api/swagger.json",
			"/api/docs",
			"/api/call/{capability}",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}