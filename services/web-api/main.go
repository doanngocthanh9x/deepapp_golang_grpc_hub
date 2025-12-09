package main

import (
	"log"
	"net/http"
	"os"

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
	pythonHandler := handlers.NewPythonWorkerHandler(hubClient)
	javaHandler := handlers.NewJavaWorkerHandler(hubClient)
	dynamicHandler := handlers.NewDynamicHandler(hubClient)
	statusHandler := handlers.NewStatusHandler()
	indexHandler := ui.NewIndexHandler()

	// Setup HTTP routes

	// Main UI
	http.HandleFunc("/", indexHandler.HandleIndex)

	// Python Worker endpoints
	http.HandleFunc("/api/worker/python/hello", pythonHandler.HandleHello)
	http.HandleFunc("/api/worker/python/analyze_image", pythonHandler.HandleAnalyzeImage)

	// Java Worker endpoints
	http.HandleFunc("/api/worker/java/hello", javaHandler.HandleHello)
	http.HandleFunc("/api/worker/java/file_info", javaHandler.HandleFileInfo)

	// Dynamic API endpoints
	http.HandleFunc("/api/capabilities", dynamicHandler.HandleCapabilities)
	http.HandleFunc("/api/swagger.json", dynamicHandler.HandleSwagger)
	http.HandleFunc("/api/docs", dynamicHandler.HandleSwaggerUI)
	http.HandleFunc("/api/call/", dynamicHandler.HandleDynamicCall)

	// Status endpoint
	http.HandleFunc("/api/status", statusHandler.HandleStatus)

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