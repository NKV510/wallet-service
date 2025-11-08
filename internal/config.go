package internal

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	MaxDBConns int32
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load("config.env"); err != nil {
		return nil, fmt.Errorf("error loading config.env: %w", err)
	}

	maxConns, err := strconv.Atoi(getEnv("MAX_DB_CONNS", "10"))
	if err != nil {
		maxConns = 10
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "wallet_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		MaxDBConns: int32(maxConns),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
