package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "deepapp_golang_grpc_hub/internal/proto"
)

type WebAPI struct {
	hubClient *HubClient
}

type HubClient struct {
	conn     *grpc.ClientConn
	client   pb.HubServiceClient
	stream   pb.HubService_ConnectClient
	clientID string
	responses chan *pb.Message
}

func NewHubClient(serverAddr string) (*HubClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := pb.NewHubServiceClient(conn)
	stream, err := client.Connect(context.Background())
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	hc := &HubClient{
		conn:      conn,
		client:    client,
		stream:    stream,
		clientID:  fmt.Sprintf("web-api-%d", time.Now().UnixNano()),
		responses: make(chan *pb.Message, 100),
	}

	// Start receiving messages
	go hc.receiveMessages()

	return hc, nil
}

func (hc *HubClient) receiveMessages() {
	for {
		msg, err := hc.stream.Recv()
		if err != nil {
			log.Printf("Receive error: %v", err)
			return
		}
		hc.responses <- msg
	}
}

func (hc *HubClient) SendRequest(targetWorker, capability, data string) (*pb.Message, error) {
	msg := pb.Message{
		Id:        fmt.Sprintf("req-%d", time.Now().UnixNano()),
		From:      hc.clientID,
		To:        targetWorker,
		Content:   data,
		Channel:   capability, // Use capability as channel for worker routing
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_REQUEST,
		Action:    "request",
		Metadata: map[string]string{
			"capability": capability,
		},
	}

	log.Printf("📤 Sending request: Type=%v (%d), Action='%s', Capability='%s', To='%s'", 
		msg.Type, msg.Type, msg.Action, capability, targetWorker)

	if err := hc.stream.Send(&msg); err != nil {
		return nil, err
	}

	// Wait for response with timeout
	select {
	case response := <-hc.responses:
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

func (api *WebAPI) handleHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send request to Python worker
	response, err := api.hubClient.SendRequest("python-worker", "hello", "")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"response": response.Content,
		"from":     response.From,
		"timestamp": response.Timestamp,
	})
}

func (api *WebAPI) handleImageAnalysis(w http.ResponseWriter, r *http.Request) {
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
	response, err := api.hubClient.SendRequest("python-worker", "analyze_image", string(requestJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"response": response.Content,
		"from":     response.From,
		"timestamp": response.Timestamp,
	})
}

func (api *WebAPI) handleJavaHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Send request to Java simple worker
	response, err := api.hubClient.SendRequest("java-simple-worker", "hello_world", "{}")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"response": response.Content,
		"from":     response.From,
		"timestamp": response.Timestamp,
	})
}

func (api *WebAPI) handleJavaFileInfo(w http.ResponseWriter, r *http.Request) {
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
	response, err := api.hubClient.SendRequest("java-simple-worker", "read_file_info", string(requestJSON))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"response": response.Content,
		"from":     response.From,
		"timestamp": response.Timestamp,
	})
}

func (api *WebAPI) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "running",
		"service":     "web-api",
		"hub_address": "localhost:50051", // Default hub address
		"endpoints": []string{
			"/api/hello",
			"/api/analyze",
			"/api/java/hello",
			"/api/java/file-info",
			"/api/status",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

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
	log.Printf("Connecting to gRPC Hub at %s...", hubAddress)
	hubClient, err := NewHubClient(hubAddress)
	if err != nil {
		log.Fatalf("Failed to connect to hub: %v", err)
	}
	log.Printf("Connected to hub with client ID: %s", hubClient.clientID)

	api := &WebAPI{hubClient: hubClient}

	// Setup HTTP routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Web API - gRPC Hub</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        button { background: #4CAF50; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        button:hover { background: #45a049; }
        #result { background: #e8f5e9; padding: 15px; margin-top: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Web API - gRPC Hub Demo</h1>
    
    <div class="endpoint">
        <h3>1. Hello World</h3>
        <button onclick="testHello()">Test Hello</button>
    </div>
    
    <div class="endpoint">
        <h3>2. Image Analysis</h3>
        <input type="file" id="imageFile" accept="image/*">
        <button onclick="testImageAnalysis()">Analyze Image</button>
    </div>
    
    <div class="endpoint">
        <h3>3. Status</h3>
        <button onclick="testStatus()">Check Status</button>
    </div>
    
    <div class="endpoint">
        <h3>4. Java Hello World</h3>
        <button onclick="testJavaHello()">Test Java Hello</button>
    </div>
    
    <div class="endpoint">
        <h3>5. Java File Info</h3>
        <input type="text" id="filePath" placeholder="Enter file path (e.g., /etc/hosts)" style="width: 300px; margin-right: 10px;">
        <button onclick="testJavaFileInfo()">Get File Info</button>
    </div>
    
    <div id="result"></div>
    
    <script>
        function testHello() {
            fetch('/api/hello', { method: 'POST' })
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').innerHTML = 
                        '<h3>Result:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(err => alert('Error: ' + err));
        }
        
        function testImageAnalysis() {
            const file = document.getElementById('imageFile').files[0];
            if (!file) {
                alert('Please select an image');
                return;
            }
            
            const formData = new FormData();
            formData.append('image', file);
            
            fetch('/api/analyze', { method: 'POST', body: formData })
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').innerHTML = 
                        '<h3>Result:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(err => alert('Error: ' + err));
        }
        
        function testStatus() {
            fetch('/api/status')
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').innerHTML = 
                        '<h3>Result:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(err => alert('Error: ' + err));
        }
        
        function testJavaHello() {
            fetch('/api/java/hello', { method: 'POST' })
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').innerHTML = 
                        '<h3>Java Hello Result:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(err => alert('Error: ' + err));
        }
        
        function testJavaFileInfo() {
            const filePath = document.getElementById('filePath').value;
            if (!filePath) {
                alert('Please enter a file path');
                return;
            }
            
            fetch('/api/java/file-info', { 
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ filePath: filePath })
            })
                .then(r => r.json())
                .then(data => {
                    document.getElementById('result').innerHTML = 
                        '<h3>Java File Info Result:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(err => alert('Error: ' + err));
        }
    </script>
</body>
</html>
		`)
	})
	
	// Static endpoints
	http.HandleFunc("/api/hello", api.handleHello)
	http.HandleFunc("/api/analyze", api.handleImageAnalysis)
	http.HandleFunc("/api/java/hello", api.handleJavaHello)
	http.HandleFunc("/api/java/file-info", api.handleJavaFileInfo)
	http.HandleFunc("/api/status", api.handleStatus)

	// Dynamic endpoints
	http.HandleFunc("/api/capabilities", api.handleCapabilities)
	http.HandleFunc("/api/swagger.json", api.handleSwagger)
	http.HandleFunc("/api/docs", api.handleSwaggerUI)
	http.HandleFunc("/api/call/", api.handleDynamicCall)

	// Start HTTP server
	portAddr := ":" + port
	log.Printf("Starting Web API on http://localhost%s", portAddr)
	log.Printf("Open http://localhost:%s in your browser", port)
	
	if err := http.ListenAndServe(portAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
