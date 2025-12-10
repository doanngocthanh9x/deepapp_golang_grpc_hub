package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/deepapp/go-worker/plugins"
)

func main() {
	hubAddress := os.Getenv("HUB_ADDRESS")
	if hubAddress == "" {
		hubAddress = "localhost:50051"
	}

	// Initialize GRPC worker
	worker := NewGRPCWorker("go-worker", hubAddress)

	// Register plugins
	worker.RegisterPlugin(plugins.NewHelloGoPlugin())
	worker.RegisterPlugin(plugins.NewHashTextPlugin())
	worker.RegisterPlugin(plugins.NewBase64OpsPlugin())
	worker.RegisterPlugin(plugins.NewGoCompositePlugin())

	fmt.Printf("✅ Loaded %d plugins\n", len(worker.plugins))

	// Connect to Hub
	if err := worker.Connect(); err != nil {
		log.Fatalf("Failed to connect to Hub: %v", err)
	}

	fmt.Println("✅ Worker registered with Hub")
	fmt.Println("🚀 Go Worker is running!")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Shutting down Go Worker...")
	worker.Close()
}
