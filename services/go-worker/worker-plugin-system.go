package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/deepapp/go-worker/plugins"
	workertoworker "github.com/deepapp/go-worker/worker-to-worker"
)

type HeartbeatMessage struct {
	Timestamp string
}

type GoWorker struct {
	workerID     string
	hubAddress   string
	plugins      map[string]plugins.Plugin
	pendingCalls map[string]*PendingCall
	mu           sync.RWMutex
	connected    bool
}

type PendingCall struct {
	ResponseChan chan *Response
	Timer        *time.Timer
}

type Response struct {
	Data  map[string]interface{}
	Error error
}

func NewGoWorker(workerID, hubAddress string) *GoWorker {
	return &GoWorker{
		workerID:     workerID,
		hubAddress:   hubAddress,
		plugins:      make(map[string]plugins.Plugin),
		pendingCalls: make(map[string]*PendingCall),
		connected:    false,
	}
}

func (w *GoWorker) Initialize() error {
	log.Println("🔵 Go Worker Starting...")
	log.Printf("   Worker ID: %s", w.workerID)
	log.Printf("   Hub Address: %s", w.hubAddress)

	// Register standard plugins
	log.Println("\n📦 Loading plugins...")
	w.registerPlugin(&plugins.HelloPlugin{})
	w.registerPlugin(&plugins.HashPlugin{})
	w.registerPlugin(&plugins.Base64Plugin{})

	// Register worker-to-worker plugins
	log.Println("\n🔄 Loading worker-to-worker plugins...")
	w.registerPlugin(&workertoworker.CompositePlugin{})

	log.Printf("\n✅ Registered %d capabilities:\n", len(w.plugins))
	for name, plugin := range w.plugins {
		log.Printf("   - %s: %s", name, plugin.GetDescription())
	}

	return nil
}

func (w *GoWorker) registerPlugin(plugin plugins.Plugin) {
	name := plugin.GetName()
	w.plugins[name] = plugin
	log.Printf("   ✓ Loaded: %s", name)
}

func (w *GoWorker) GetCapabilities() []*plugins.Capability {
	caps := make([]*plugins.Capability, 0, len(w.plugins))
	for _, plugin := range w.plugins {
		caps = append(caps, plugins.ToCapability(plugin))
	}
	return caps
}

func (w *GoWorker) HandleRequest(requestID, capability, data string) (string, error) {
	log.Printf("\n📥 Request: %s (ID: %s)", capability, requestID)

	plugin, exists := w.plugins[capability]
	if !exists {
		return "", fmt.Errorf("unknown capability: %s", capability)
	}

	params, err := plugins.ParseParams(data)
	if err != nil {
		return "", fmt.Errorf("invalid request data: %v", err)
	}

	context := &plugins.ExecutionContext{
		WorkerID:   w.workerID,
		CallWorker: w.CallWorker,
	}

	result, err := plugin.Execute(params, context)
	if err != nil {
		return "", err
	}

	response, err := plugins.FormatResult(result)
	if err != nil {
		return "", fmt.Errorf("failed to format response: %v", err)
	}

	log.Printf("✅ Response sent for: %s", capability)
	return response, nil
}

func (w *GoWorker) CallWorker(targetWorkerID, capability string, params map[string]interface{}, timeout int) (map[string]interface{}, error) {
	requestID := fmt.Sprintf("%s-%d-%d", w.workerID, time.Now().UnixNano(), len(w.pendingCalls))

	responseChan := make(chan *Response, 1)
	timer := time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
		w.mu.Lock()
		delete(w.pendingCalls, requestID)
		w.mu.Unlock()
		responseChan <- &Response{Error: fmt.Errorf("timeout calling %s.%s", targetWorkerID, capability)}
	})

	w.mu.Lock()
	w.pendingCalls[requestID] = &PendingCall{
		ResponseChan: responseChan,
		Timer:        timer,
	}
	w.mu.Unlock()

	// TODO: Send WORKER_CALL message through stream
	log.Printf("  → Calling %s.%s (Request ID: %s)", targetWorkerID, capability, requestID)

	response := <-responseChan
	return response.Data, response.Error
}

func (w *GoWorker) Start() error {
	if err := w.Initialize(); err != nil {
		return err
	}

	// TODO: Connect to Hub via gRPC stream
	log.Println("\n🚀 Go Worker is running!\n")
	log.Println("⚠️  Note: Full gRPC integration pending - plugins loaded successfully")

	// Keep running
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\n👋 Shutting down gracefully...")
	return nil
}

func main() {
	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		workerID = "go-worker"
	}

	hubAddress := os.Getenv("HUB_ADDRESS")
	if hubAddress == "" {
		hubAddress = "localhost:50051"
	}

	worker := NewGoWorker(workerID, hubAddress)
	if err := worker.Start(); err != nil {
		log.Fatalf("❌ Fatal error: %v", err)
	}
}
