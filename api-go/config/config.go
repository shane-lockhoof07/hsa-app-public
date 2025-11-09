package config

import (
    "os"
)

type Config struct {
    DatabaseURL    string
    OCRServiceURL  string
    HSADir         string
    Port           string
    ClaudeAPIKey   string
    ClaudeModel    string
}

func Load() *Config {
    return &Config{
        DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:postgres-password@localhost:5432/pg-database?sslmode=disable"),
        OCRServiceURL:  getEnv("OCR_SERVICE_URL", "http://localhost:8001"),
        HSADir:         getEnv("HSA_DIR", "/data/hsa"),
        Port:           getEnv("PORT", "8080"),
        ClaudeAPIKey:   getEnv("CLAUDE_API_KEY", "YOUR-CLAUDE-API-KEY"),
        ClaudeModel:    getEnv("CLAUDE_MODEL", "claude-3-5-haiku-20241022"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}