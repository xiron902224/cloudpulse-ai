package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    Port        string
    Environment string
    DatabaseURL string
    RedisURL    string
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
    }

    return &Config{
        Port:        getEnv("PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        DatabaseURL: getEnv("DATABASE_URL", "postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse?sslmode=disable"),
        RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}