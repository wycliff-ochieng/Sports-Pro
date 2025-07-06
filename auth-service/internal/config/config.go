package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBPassword string
	DBName     string
	DBUser     string
	DBsslmode  string

	JWTSecret     string
	JWTExpiry     string
	RefreshSecret string
	RefreshExpiry string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning, failed to load env varivales: %v", err)
	}

	config := &Config{}

	config.DBHost = getEnv("DB_HOST", "localhost")
	config.DBPort = getEnvAsInt("DB_PORT", 5433)
	config.DBPassword = getEnv("DB_PASSWORD", "admin123")
	config.DBName = getEnv("DB_NAME", "Authentication")
	config.DBUser = getEnv("DB_USER", "admin")
	config.DBsslmode = getEnv("DB_SSLMODE", "disable")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
