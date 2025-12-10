package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"deepapp_golang_grpc_hub/services/web-api/internal/client"
	"deepapp_golang_grpc_hub/services/web-api/internal/handlers"
	"deepapp_golang_grpc_hub/services/web-api/internal/ui"
)

func main() {
	// Get Hub address from environment or use default
	hubAddress := os.Getenv("HUB_ADDRESS")
	if hubAddress == "" {
		hubAddress = "localhost:50051"
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Connect to gRPC Hub
	log.Printf("🌐 Connecting to gRPC Hub at %s...", hubAddress)
	hubClient, err := client.NewHubClient(hubAddress)
	if err != nil {
		log.Fatalf("❌ Failed to connect to hub: %v", err)
	}
	defer hubClient.Close()
	log.Printf("✅ Connected to hub with client ID: %s", hubClient.ClientID)

	// Initialize handlers
	dynamicHandler := handlers.NewDynamicHandler(hubClient)
	statusHandler := handlers.NewStatusHandler()
	indexHandler := ui.NewIndexHandler()

	// Setup HTTP routes (100% Dynamic - No hard-coded endpoints!)
	log.Println("🔌 Setting up dynamic routes from Hub registry...")

	// Main UI
	http.HandleFunc("/", indexHandler.HandleIndex)

	// Core API endpoints
	http.HandleFunc("/api/capabilities", dynamicHandler.HandleCapabilities)
	http.HandleFunc("/api/swagger.json", dynamicHandler.HandleSwagger)
	http.HandleFunc("/api/docs", dynamicHandler.HandleSwaggerUI)
	http.HandleFunc("/api/status", statusHandler.HandleStatus)

	// Dynamic worker-specific routes
	// Pattern: /api/{worker_id}/call/{capability}
	// Examples:
	//   /api/python-worker/call/hello
	//   /api/java-simple-worker/call/read_file_info
	//   /api/node-worker/call/process_data
	http.HandleFunc("/api/", dynamicHandler.HandleWorkerCall)

	log.Println("✅ All routes registered dynamically from Hub")

	// Discover and log all available capabilities
	log.Println("\n📡 Discovering available capabilities from Hub...")
	go func() {
		// Give workers time to register
		time.Sleep(2 * time.Second)
		
		// Query Hub for capabilities
		discoveryMsg := map[string]interface{}{
			"action": "discover",
		}
		discoveryJSON, _ := json.Marshal(discoveryMsg)
		
		response, err := hubClient.SendRequest("hub", "capability_discovery", string(discoveryJSON))
		if err != nil {
			log.Printf("⚠️  Could not discover capabilities: %v", err)
			return
		}
		
		var result struct {
			Capabilities map[string]interface{} `json:"capabilities"`
			Workers      []map[string]interface{} `json:"workers"`
		}
		
		if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
			log.Printf("⚠️  Could not parse capabilities: %v", err)
			return
		}
		
		log.Println("\n🎯 Auto-discovered API Endpoints:")
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		
		count := 0
		for capName, capData := range result.Capabilities {
			if capMap, ok := capData.(map[string]interface{}); ok {
				description := "No description"
				if desc, ok := capMap["description"].(string); ok {
					description = desc
				}
				
				httpMethod := "POST"
				if method, ok := capMap["http_method"].(string); ok {
					httpMethod = method
				}
				
				endpoint := fmt.Sprintf("/api/call/%s", capName)
				log.Printf("  %s  %-30s  %s", httpMethod, endpoint, description)
				count++
			}
		}
		
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Printf("✅ Total: %d dynamic endpoints from %d workers\n", count, len(result.Workers))
		
		// Log worker info
		log.Println("\n👷 Active Workers:")
		for _, worker := range result.Workers {
			workerID := worker["id"]
			workerType := worker["type"]
			capCount := len(worker["capabilities"].([]interface{}))
			log.Printf("  • %s (%s) - %d capabilities", workerID, workerType, capCount)
		}
		log.Println()
	}()

	// Start HTTP server
	portAddr := ":" + port
	log.Printf("🚀 Starting Web API on http://localhost%s", portAddr)
	log.Printf("📱 Open http://localhost:%s in your browser", port)
	log.Printf("📚 API Docs: http://localhost:%s/api/docs", port)
	log.Printf("🔍 Capabilities: http://localhost:%s/api/capabilities", port)

	if err := http.ListenAndServe(portAddr, nil); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}