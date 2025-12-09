package main

import (
	"log"
    "fmt"

	"deepapp_golang_grpc_hub/internal/config"
	"deepapp_golang_grpc_hub/internal/db"
	"deepapp_golang_grpc_hub/internal/hub"
	"deepapp_golang_grpc_hub/pkg/logger"
)

func main() {
	// Load configuration
	fmt.Println("Loading configuration...")
	cfg := config.Load()
	fmt.Printf("Config loaded: Port=%s, LogLevel=%s, DBPath=%s\n", cfg.Port, cfg.LogLevel, cfg.DBPath)

	// Initialize logger
	fmt.Println("Initializing logger...")
	logger.Init(cfg.LogLevel)
	fmt.Println("Logger initialized")

	// Initialize database
	fmt.Println("Initializing database...")
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	fmt.Println("Database initialized")

	// Create service registry with database
	fmt.Println("Creating service registry with database...")
	registry := hub.NewServiceRegistryWithDB(database)
	fmt.Println("Service registry created")

	// Start the hub server
	fmt.Println("Creating server...")
	server := hub.NewServerWithRegistry(cfg, registry)
	fmt.Printf("Server created, starting on port %s...\n", cfg.Port)
	
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}