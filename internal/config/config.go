package config

import (
	"os"
)

type Config struct {
	Port     string
	LogLevel string
	DBPath   string
}

func Load() *Config {
	port := getEnv("PORT", "50051")
	logLevel := getEnv("LOG_LEVEL", "info")
	dbPath := getEnv("DB_PATH", "hub.db")

	return &Config{
		Port:     port,
		LogLevel: logLevel,
		DBPath:   dbPath,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}