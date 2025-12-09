package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"deepapp_golang_grpc_hub/services/web-api/internal/client"
)

// DynamicHandler handles dynamic capability discovery and Swagger
type DynamicHandler struct {
	hubClient *client.HubClient
}

// NewDynamicHandler creates a new dynamic handler
func NewDynamicHandler(hubClient *client.HubClient) *DynamicHandler {
	return &DynamicHandler{hubClient: hubClient}
}

// HandleCapabilities returns all available capabilities from Hub
func (h *DynamicHandler) HandleCapabilities(w http.ResponseWriter, r *http.Request) {
	// Send discovery request to Hub
	discoveryMsg := map[string]interface{}{
		"action": "discover",
	}
	discoveryJSON, _ := json.Marshal(discoveryMsg)

	response, err := h.hubClient.SendRequest("hub", "capability_discovery", string(discoveryJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error discovering capabilities: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		http.Error(w, "Failed to parse capabilities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleSwagger generates dynamic Swagger documentation
func (h *DynamicHandler) HandleSwagger(w http.ResponseWriter, r *http.Request) {
	// Get capabilities from Hub
	discoveryMsg := map[string]interface{}{
		"action": "discover",
	}
	discoveryJSON, _ := json.Marshal(discoveryMsg)

	response, err := h.hubClient.SendRequest("hub", "capability_discovery", string(discoveryJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	var discoveryResult struct {
		Capabilities map[string]interface{} `json:"capabilities"`
		Workers      []interface{}          `json:"workers"`
	}
	json.Unmarshal([]byte(response.Content), &discoveryResult)

	// Build server URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8081"
	}
	serverURL := fmt.Sprintf("%s://%s", scheme, host)

	// Generate Swagger spec
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "DeepApp gRPC Hub API",
			"description": "Dynamic API generated from worker capabilities",
			"version":     "1.0.0",
		},
		"servers": []map[string]interface{}{
			{
				"url":         serverURL,
				"description": "API Server",
			},
		},
		"paths": make(map[string]interface{}),
	}

	paths := spec["paths"].(map[string]interface{})

	// Add dynamic endpoints based on capabilities
	for capName, capData := range discoveryResult.Capabilities {
		capMap, ok := capData.(map[string]interface{})
		if !ok {
			continue
		}

		description := ""
		if desc, ok := capMap["description"].(string); ok {
			description = desc
		}

		httpMethod := "post"
		if method, ok := capMap["http_method"].(string); ok {
			httpMethod = strings.ToLower(method) // Normalize to lowercase for OpenAPI spec
		}

		acceptsFile := false
		if af, ok := capMap["accepts_file"].(bool); ok {
			acceptsFile = af
		}

		fileFieldName := "file"
		if ffn, ok := capMap["file_field_name"].(string); ok && ffn != "" {
			fileFieldName = ffn
		}

		// Create path
		path := fmt.Sprintf("/api/call/%s", capName)

		requestBody := map[string]interface{}{
			"required": true,
			"content":  map[string]interface{}{},
		}

		content := requestBody["content"].(map[string]interface{})

		if acceptsFile {
			// Multipart form data for file upload
			content["multipart/form-data"] = map[string]interface{}{
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						fileFieldName: map[string]interface{}{
							"type":   "string",
							"format": "binary",
						},
						"params": map[string]interface{}{
							"type":        "object",
							"description": "Additional parameters as JSON",
						},
					},
				},
			}
		} else {
			// JSON request body
			content["application/json"] = map[string]interface{}{
				"schema": map[string]interface{}{
					"type":        "object",
					"description": "Request parameters",
				},
			}
		}

		operation := map[string]interface{}{
			"summary":     fmt.Sprintf("Call %s capability", capName),
			"description": description,
			"requestBody": requestBody,
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"status":    map[string]interface{}{"type": "string"},
									"response":  map[string]interface{}{"type": "string"},
									"from":      map[string]interface{}{"type": "string"},
									"timestamp": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
			},
		}

		paths[path] = map[string]interface{}{
			httpMethod: operation,
		}
	}

	// Add static endpoints
	paths["/api/capabilities"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get all available capabilities",
			"description": "Returns list of all registered worker capabilities",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of capabilities",
				},
			},
		},
	}

	paths["/api/status"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "API Status",
			"description": "Check API health status",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Status information",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

// HandleSwaggerUI serves Swagger UI HTML
func (h *DynamicHandler) HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DeepApp Hub API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/api/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// HandleDynamicCall handles dynamic API calls to any capability
func (h *DynamicHandler) HandleDynamicCall(w http.ResponseWriter, r *http.Request) {
	// Extract capability name from path: /api/call/{capability}
	capabilityName := r.URL.Path[len("/api/call/"):]

	if capabilityName == "" {
		http.Error(w, "Capability name required", http.StatusBadRequest)
		return
	}

	var requestData string

	// Check if request has file upload
	contentType := r.Header.Get("Content-Type")
	if len(contentType) > 19 && contentType[:19] == "multipart/form-data" {
		// Handle file upload
		err := r.ParseMultipartForm(100 << 20) // 100 MB max
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			// Try other common field names
			file, header, err = r.FormFile("image")
			if err != nil {
				file, header, err = r.FormFile("document")
			}
		}

		if err != nil {
			http.Error(w, "File required for this capability", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file
		fileData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Get additional params
		paramsStr := r.FormValue("params")
		var params map[string]interface{}
		if paramsStr != "" {
			json.Unmarshal([]byte(paramsStr), &params)
		} else {
			params = make(map[string]interface{})
		}

		// Add file info to params
		params["filename"] = header.Filename
		params["size"] = len(fileData)
		params["content_type"] = header.Header.Get("Content-Type")

		// Encode file as base64 if needed
		// For now, just send file info
		requestJSON, _ := json.Marshal(params)
		requestData = string(requestJSON)
	} else {
		// Handle JSON request
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		requestJSON, _ := json.Marshal(body)
		requestData = string(requestJSON)
	}

	// Send to Hub (let Hub route to appropriate worker)
	response, err := h.hubClient.SendRequest("", capabilityName, requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"response":  response.Content,
		"from":      response.From,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}