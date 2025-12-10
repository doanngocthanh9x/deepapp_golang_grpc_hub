package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	workersdk "deepapp_golang_grpc_hub/shared/worker-sdk/go"
)

type GoWorker struct {
	sdk *workersdk.WorkerSDK
}

func NewGoWorker(workerID, hubAddress string) *GoWorker {
	sdk := workersdk.NewWorkerSDK(workerID, hubAddress, "golang")
	
	worker := &GoWorker{sdk: sdk}
	worker.registerCapabilities()
	
	return worker
}

func (w *GoWorker) registerCapabilities() {
	// Hello capability
	w.sdk.AddCapability(&workersdk.Capability{
		Name:         "hello_go",
		Description:  "Returns a hello message from Go",
		InputSchema:  "{}",
		OutputSchema: `{"type":"object","properties":{"message":{"type":"string"}}}`,
		HTTPMethod:   "POST",
		AcceptsFile:  false,
	}, w.handleHello)
	
	// Calculate capability
	w.sdk.AddCapability(&workersdk.Capability{
		Name:         "calculate",
		Description:  "Performs calculations",
		InputSchema:  `{"type":"object","properties":{"operation":{"type":"string"},"a":{"type":"number"},"b":{"type":"number"}}}`,
		OutputSchema: `{"type":"object","properties":{"result":{"type":"number"}}}`,
		HTTPMethod:   "POST",
		AcceptsFile:  false,
	}, w.handleCalculate)
	
	// Composite task - calls Java worker
	w.sdk.AddCapability(&workersdk.Capability{
		Name:         "go_composite",
		Description:  "Calls Java worker for file info",
		InputSchema:  `{"type":"object","properties":{"file_path":{"type":"string"}}}`,
		OutputSchema: `{"type":"object"}`,
		HTTPMethod:   "POST",
		AcceptsFile:  false,
	}, w.handleComposite)
}

func (w *GoWorker) handleHello(params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"message":   "Hello from Go Worker! 🔵",
		"worker_id": w.sdk,
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

func (w *GoWorker) handleCalculate(params map[string]interface{}) (map[string]interface{}, error) {
	operation, _ := params["operation"].(string)
	a, _ := params["a"].(float64)
	b, _ := params["b"].(float64)
	
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
	
	return map[string]interface{}{
		"result":    result,
		"operation": operation,
		"status":    "success",
	}, nil
}

func (w *GoWorker) handleComposite(params map[string]interface{}) (map[string]interface{}, error) {
	filePath, _ := params["file_path"].(string)
	if filePath == "" {
		filePath = "/tmp/test.txt"
	}
	
	// Step 1: Do local processing
	goResult := map[string]interface{}{
		"processed_by": "golang",
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	
	// Step 2: Call Java worker
	log.Printf("  → Calling Java worker for file info...")
	javaResponse, err := w.sdk.CallWorker(
		"java-simple-worker",
		"read_file_info",
		map[string]interface{}{"filePath": filePath},
		30*time.Second,
	)
	
	if err != nil {
		// Return partial result on error
		return map[string]interface{}{
			"go_processing":    goResult,
			"java_call_error":  err.Error(),
			"combined_status":  "partial",
		}, nil
	}
	
	return map[string]interface{}{
		"go_processing":   goResult,
		"java_file_info":  javaResponse,
		"combined_status": "success",
	}, nil
}

func (w *GoWorker) Run() error {
	return w.sdk.Run()
}

func (w *GoWorker) Stop() {
	w.sdk.Stop()
}

func main() {
	// Get configuration from environment
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = "go-worker"
	}
	
	hubAddress := os.Getenv("HUB_ADDRESS")
	if hubAddress == "" {
		hubAddress = "localhost:50051"
	}
	
	// Create worker
	worker := NewGoWorker(workerID, hubAddress)
	
	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Println("\n\n✗ Shutting down...")
		worker.Stop()
		os.Exit(0)
	}()
	
	// Run worker
	if err := worker.Run(); err != nil {
		log.Fatalf("Worker error: %v", err)
	}
}
