package config

import (
	"context"
	"os"
	"time"
)

type Config struct {
	Port         string
	MongoURI     string
	MongoDB      string
	RabbitURL    string
	Exchange     string
	ExchangeType string
	UsersAPIBase string
	JWTSecret    string
	Ctx          context.Context
	Timeout      time.Duration
}

func Load() *Config {
	return &Config{
		Port:         getEnv("ACTIVITIES_PORT", "8082"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:      getEnv("MONGO_DB", "sporthub"),
		RabbitURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Exchange:     getEnv("RABBITMQ_EXCHANGE", "activities.events"),
		ExchangeType: getEnv("RABBITMQ_EXCHANGE_TYPE", "topic"),
		UsersAPIBase: getEnv("USERS_API_BASE_URL", "http://localhost:8081"),
		JWTSecret:    getEnv("JWT_SECRET", "change_me"),
		Ctx:          context.Background(),
		Timeout:      10 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
