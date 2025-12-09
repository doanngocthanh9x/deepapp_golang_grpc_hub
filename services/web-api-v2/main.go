package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	pb "deepapp_golang_grpc_hub/internal/proto"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "deepapp_golang_grpc_hub/services/web-api-v2/docs" // Swagger docs
)

// @title gRPC Hub Dynamic API
// @version 2.0
// @description Dynamic API Gateway với auto-discovery từ workers
// @host localhost:8082
// @BasePath /api/v2

type APIServer struct {
	hubClient      pb.HubServiceClient
	hubStream      pb.HubService_ConnectClient
	clientID       string
	capabilities   map[string]interface{} // Dynamic capabilities từ workers
	responseChans  map[string]chan string
	swaggerGen     *SwaggerGenerator // Dynamic Swagger generator
}

func main() {
	hubAddr := os.Getenv("HUB_ADDRESS")
	if hubAddr == "" {
		hubAddr = "localhost:50051"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Connect to gRPC Hub
	conn, err := grpc.Dial(hubAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to hub: %v", err)
	}
	defer conn.Close()

	client := pb.NewHubServiceClient(conn)
	
	server := &APIServer{
		hubClient:     client,
		clientID:      fmt.Sprintf("web-api-v2-%d", time.Now().UnixNano()),
		capabilities:  make(map[string]interface{}),
		responseChans: make(map[string]chan string),
		swaggerGen:    NewSwaggerGenerator(),
	}

	// Start gRPC stream
	if err := server.connectToHub(); err != nil {
		log.Fatalf("Failed to connect stream: %v", err)
	}

	// Start listening for responses
	go server.listenForResponses()

	// Discover capabilities
	time.Sleep(1 * time.Second)
	go server.discoverCapabilities()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Swagger endpoint - Static fallback
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// Dynamic Swagger UI
	r.Static("/static", "./static")
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static/swagger.html")
	})
	
	// Dynamic Swagger from workers
	//r.GET("/swagger/doc.json", server.getDynamicSwagger)

	// API routes
	api := r.Group("/api/v2")
	{
		api.GET("/status", server.getStatus)
		api.GET("/capabilities", server.getCapabilities)
		api.POST("/invoke/:capability", server.invokeCapability)
	}

	log.Printf("🚀 API Server v2 started on port %s", port)
	log.Printf("📚 Swagger UI: http://localhost:%s/swagger/index.html", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
func (s *APIServer) connectToHub() error {
	stream, err := s.hubClient.Connect(context.Background())
	if err != nil {
		return err
	}
	s.hubStream = stream

	// Send registration message
	regMsg := &pb.Message{
		Id:        fmt.Sprintf("reg-%d", time.Now().UnixNano()),
		From:      s.clientID,
		To:        "hub",
		Type:      pb.MessageType_REGISTER,
		Timestamp: time.Now().Format(time.RFC3339),
		Content:   `{"type":"api_gateway","version":"2.0"}`,
	}

	return stream.Send(regMsg)
}

func (s *APIServer) listenForResponses() {
	for {
		msg, err := s.hubStream.Recv()
		if err != nil {
			log.Printf("Stream error: %v", err)
			return
		}

		// Route response to waiting channel
		if ch, exists := s.responseChans[msg.Id]; exists {
			ch <- msg.Content
			close(ch)
			delete(s.responseChans, msg.Id)
		}
		
		// Check if this is capability discovery response
		if strings.Contains(msg.Content, "capabilities") {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Content), &response); err == nil {
				if caps, ok := response["capabilities"].(map[string]interface{}); ok {
					s.capabilities = caps
					s.swaggerGen.UpdateCapabilities(caps)
					log.Printf("📚 Updated Swagger with %d capabilities", len(caps))
				}
			}
		}
	}
}

func (s *APIServer) discoverCapabilities() {
	// Request capability list from hub
	msgID := fmt.Sprintf("discover-%d", time.Now().UnixNano())
	msg := &pb.Message{
		Id:        msgID,
		From:      s.clientID,
		To:        "hub",
		Type:      pb.MessageType_REQUEST,
		Content:   `{"action":"list_capabilities"}`,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.hubStream.Send(msg)
	log.Println("🔍 Discovering capabilities from workers...")
}

// getStatus godoc
// @Summary Get API status
// @Description Get current API gateway status and info
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /status [get]
func (s *APIServer) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"client_id":  s.clientID,
		"status":     "running",
		"version":    "2.0",
		"timestamp":  time.Now().Format(time.RFC3339),
		"capabilities_count": len(s.capabilities),
	})
}

// getCapabilities godoc
// @Summary List all available capabilities
// @Description Get all service capabilities registered by workers
// @Tags discovery
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /capabilities [get]
func (s *APIServer) getCapabilities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"capabilities": s.capabilities,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// getDynamicSwagger serves auto-generated Swagger spec from worker capabilities
func (s *APIServer) getDynamicSwagger(c *gin.Context) {
	spec := s.swaggerGen.GenerateSpec()
	c.JSON(http.StatusOK, spec)
}

// invokeCapability godoc
// @Summary Invoke a service capability
// @Description Dynamically invoke any registered service capability
// @Tags services
// @Accept json
// @Produce json
// @Param capability path string true "Capability name" example(hello)
// @Param request body map[string]interface{} true "Request payload"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /invoke/{capability} [post]
func (s *APIServer) invokeCapability(c *gin.Context) {
	capability := c.Param("capability")

	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create request message
	msgID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	requestData := map[string]interface{}{
		"capability": capability,
		"payload":    payload,
	}

	content, _ := json.Marshal(requestData)
	msg := &pb.Message{
		Id:        msgID,
		From:      s.clientID,
		To:        "worker", // Hub will route to appropriate worker
		Type:      pb.MessageType_REQUEST,
		Content:   string(content),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Create response channel
	responseChan := make(chan string, 1)
	s.responseChans[msgID] = responseChan

	// Send request
	if err := s.hubStream.Send(msg); err != nil {
		delete(s.responseChans, msgID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(response), &result); err != nil {
			c.JSON(http.StatusOK, gin.H{"data": response})
		} else {
			c.JSON(http.StatusOK, result)
		}
	case <-time.After(30 * time.Second):
		delete(s.responseChans, msgID)
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout waiting for worker response"})
	}
}
