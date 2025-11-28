package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DB_SOURCE    string
	Environment  string
	RABBITMQ_URL string
}

func LoadConfig() Config {
	godotenv.Load()
	return Config{
		Port: getEnv("PORT", "8080"),
		DB_SOURCE:    getEnv("DB_SOURCE", ""),
		Environment:  getEnv("ENVIRONMENT", "development"),
		RABBITMQ_URL: getEnv("RABBITMQ_URL", ""),
	}
}


func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}