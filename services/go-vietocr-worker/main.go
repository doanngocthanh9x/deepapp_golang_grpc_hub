package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "deepapp_golang_grpc_hub/internal/proto"
)

// VietOCRWorker - Go worker for VietOCR using ONNX Runtime via Python subprocess
type VietOCRWorker struct {
	workerID   string
	hubAddress string

	// gRPC
	conn   *grpc.ClientConn
	client pb.HubServiceClient
	stream pb.HubService_ConnectClient

	// State
	running   bool
	sendQueue chan *pb.Message
	mu        sync.RWMutex
}

// NewVietOCRWorker creates new VietOCR worker
func NewVietOCRWorker(workerID, hubAddress string) (*VietOCRWorker, error) {
	return &VietOCRWorker{
		workerID:   workerID,
		hubAddress: hubAddress,
		sendQueue:  make(chan *pb.Message, 100),
	}, nil
}

// Connect to Hub
func (w *VietOCRWorker) Connect() error {
	log.Printf("🔵 Connecting to Hub at %s...", w.hubAddress)

	conn, err := grpc.Dial(
		w.hubAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	w.conn = conn

	w.client = pb.NewHubServiceClient(conn)

	stream, err := w.client.Connect(context.Background())
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to start stream: %w", err)
	}
	w.stream = stream

	log.Println("✅ Connected to Hub")

	// Send registration
	if err := w.sendRegistration(); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	w.running = true

	// Start goroutines
	go w.sendLoop()
	go w.receiveLoop()

	return nil
}

// sendRegistration to Hub
func (w *VietOCRWorker) sendRegistration() error {
	capabilities := []map[string]interface{}{
		{
			"name":            "ocr_detect",
			"description":     "OCR nhận diện text từ ảnh (Vietnamese + English) - Go Worker",
			"input_schema":    `{"type":"object","properties":{"image":{"type":"string"}},"required":["image"]}`,
			"output_schema":   `{"type":"object","properties":{"text":{"type":"string"},"confidence":{"type":"number"},"processing_time_ms":{"type":"number"}}}`,
			"http_method":     "POST",
			"accepts_file":    true,
			"file_field_name": "image",
		},
		{
			"name":          "ocr_batch",
			"description":   "Batch OCR processing - Go Worker",
			"input_schema":  `{"type":"object","properties":{"images":{"type":"array","items":{"type":"string"}}},"required":["images"]}`,
			"output_schema": `{"type":"object","properties":{"results":{"type":"array"},"total_processing_time_ms":{"type":"number"}}}`,
			"http_method":   "POST",
			"accepts_file":  false,
		},
	}

	regData := map[string]interface{}{
		"worker_id":    w.workerID,
		"worker_type":  "go-vietocr",
		"capabilities": capabilities,
		"metadata": map[string]string{
			"version":     "1.0.0",
			"description": "VietOCR Worker - Go (high performance)",
			"language":    "Vietnamese + English",
			"engine":      "ONNX Runtime",
		},
	}

	contentBytes, _ := json.Marshal(regData)

	msg := &pb.Message{
		Id:        fmt.Sprintf("register-%d", time.Now().UnixNano()),
		From:      w.workerID,
		To:        "hub",
		Channel:   "system",
		Content:   string(contentBytes),
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_REGISTER,
	}

	if err := w.stream.Send(msg); err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	log.Println("📤 Sent registration")
	return nil
}

func (w *VietOCRWorker) sendLoop() {
	for msg := range w.sendQueue {
		if err := w.stream.Send(msg); err != nil {
			log.Printf("✗ Failed to send message: %v", err)
			return
		}
	}
}

func (w *VietOCRWorker) receiveLoop() {
	log.Println("📨 Listening for OCR requests...")

	for {
		msg, err := w.stream.Recv()
		if err != nil {
			log.Printf("✗ Stream error: %v", err)
			w.running = false
			return
		}

		go w.handleMessage(msg)
	}
}

func (w *VietOCRWorker) handleMessage(msg *pb.Message) {
	log.Printf("📬 Received message:")
	log.Printf("   ID: %s", msg.Id)
	log.Printf("   From: %s", msg.From)
	log.Printf("   Type: %s", msg.Type)
	log.Printf("   Channel: %s", msg.Channel)

	var responseContent string

	// Get capability
	capability := msg.Channel
	if cap, ok := msg.Metadata["capability"]; ok {
		capability = cap
	}

	switch capability {
	case "ocr_detect":
		responseContent = w.handleOCRDetect(msg)
	case "ocr_batch":
		responseContent = w.handleOCRBatch(msg)
	default:
		responseContent = fmt.Sprintf(`{"status":"error","error":"Unknown capability: %s"}`, capability)
	}

	// Send response
	response := &pb.Message{
		Id:        fmt.Sprintf("resp-%d", time.Now().UnixNano()),
		From:      w.workerID,
		To:        msg.From,
		Channel:   msg.Channel,
		Content:   responseContent,
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      pb.MessageType_RESPONSE,
	}

	w.sendQueue <- response
	log.Println("   ✓ Queued response")
}

func (w *VietOCRWorker) handleOCRDetect(msg *pb.Message) string {
	log.Println("  → Processing OCR detect")

	start := time.Now()

	// Parse request
	var req struct {
		Image string `json:"image"`
	}

	if err := json.Unmarshal([]byte(msg.Content), &req); err != nil {
		return fmt.Sprintf(`{"status":"error","error":"Invalid request: %v"}`, err)
	}

	// Decode base64 image
	_, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		return fmt.Sprintf(`{"status":"error","error":"Invalid base64: %v"}`, err)
	}

	// TODO: Call ONNX inference (for now, demo response)
	text := "Văn bản tiếng Việt từ Go Worker"
	confidence := 0.93

	processingTime := time.Since(start).Milliseconds()

	result := map[string]interface{}{
		"text":               text,
		"confidence":         confidence,
		"processing_time_ms": processingTime,
		"worker_id":          w.workerID,
		"status":             "success",
	}

	resultBytes, _ := json.Marshal(result)
	log.Printf("  ✓ OCR result: '%s' (conf: %.3f, time: %dms)", text, confidence, processingTime)

	return string(resultBytes)
}

func (w *VietOCRWorker) handleOCRBatch(msg *pb.Message) string {
	log.Println("  → Processing OCR batch")

	start := time.Now()

	// Parse request
	var req struct {
		Images []string `json:"images"`
	}

	if err := json.Unmarshal([]byte(msg.Content), &req); err != nil {
		return fmt.Sprintf(`{"status":"error","error":"Invalid request: %v"}`, err)
	}

	log.Printf("  📦 Processing %d images...", len(req.Images))

	results := make([]map[string]interface{}, len(req.Images))
	for i := range req.Images {
		// Process each image
		results[i] = map[string]interface{}{
			"text":       fmt.Sprintf("Text %d từ Go", i+1),
			"confidence": 0.90 + float64(i)*0.01,
			"index":      i,
		}
	}

	processingTime := time.Since(start).Milliseconds()

	result := map[string]interface{}{
		"results":                  results,
		"total_images":             len(req.Images),
		"successful":               len(results),
		"total_processing_time_ms": processingTime,
		"worker_id":                w.workerID,
		"status":                   "success",
	}

	resultBytes, _ := json.Marshal(result)
	log.Printf("  ✓ Batch complete: %d images in %dms", len(results), processingTime)

	return string(resultBytes)
}

func (w *VietOCRWorker) Close() error {
	w.running = false
	close(w.sendQueue)

	if w.conn != nil {
		return w.conn.Close()
	}

	return nil
}

func main() {
	workerID := flag.String("worker-id",
		getEnv("WORKER_ID", "go-vietocr-worker"),
		"Worker ID")

	hubAddress := flag.String("hub-address",
		getEnv("HUB_ADDRESS", "localhost:50051"),
		"Hub address")

	flag.Parse()

	log.Println(strings.Repeat("=", 60))
	log.Println("VietOCR Worker (Go)")
	log.Println(strings.Repeat("=", 60))
	log.Printf("Worker ID: %s", *workerID)
	log.Printf("Hub Address: %s", *hubAddress)
	log.Println(strings.Repeat("=", 60))

	// Create worker
	worker, err := NewVietOCRWorker(*workerID, *hubAddress)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Close()

	// Connect to Hub
	if err := worker.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("\nWorker running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("\n✗ Shutting down...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
