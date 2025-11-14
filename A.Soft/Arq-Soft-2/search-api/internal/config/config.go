package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// General
	Env  string
	Port string // expone el puerto del HTTP server (interno del contenedor)

	// Solr
	SolrURL string

	// Cache
	MemcachedAddr   string
	CacheTTLSeconds int

	// RabbitMQ (consumer)
	RabbitURL        string
	RabbitExchange   string
	RabbitQueue      string
	RabbitRoutingKey string

	// Upstream (para completar documento por ID)
	ActivitiesAPI string

	// Logging (opcional)
	LogLevel string
}

func Load() Config {
	return Config{
		Env:              envOr("ENV", "development"),
		Port:             envOr("SEARCH_API_PORT", "8080"),
		SolrURL:          envOr("SOLR_URL", "http://solr:8983/solr/sporthub_core"),
		MemcachedAddr:    envOr("MEMCACHED_ADDR", "memcached:11211"),
		CacheTTLSeconds:  envOrInt("CACHE_TTL_SECONDS", 60),
		RabbitURL:        envOr("RABBIT_URL", "amqp://guest:guest@rabbitmq:5672/"),
		RabbitExchange:   envOr("RABBIT_EXCHANGE", "activities.events"),
		RabbitQueue:      envOr("RABBIT_QUEUE", "search_sync"),
		RabbitRoutingKey: envOr("RABBIT_ROUTING_KEY", "#"),
		ActivitiesAPI:    envOr("ACTIVITIES_API_BASE", "http://activities-api:8080"),
		LogLevel:         envOr("LOG_LEVEL", "info"),
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		log.Printf("[config] WARN: %s must be an integer, using default %d", k, def)
	}
	return def
}
