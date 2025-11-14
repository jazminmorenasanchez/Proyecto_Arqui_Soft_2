package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/sporthub/search-api/internal/config"
	"github.com/sporthub/search-api/internal/consumers"
	"github.com/sporthub/search-api/internal/controllers"
	"github.com/sporthub/search-api/internal/middleware"
	"github.com/sporthub/search-api/internal/repository"
	"github.com/sporthub/search-api/internal/services"
)

func main() {
	// ---- Config
	cfg := config.Load()
	log.Printf("[search-api] env=%s port=%s solr=%s memcached=%s ttl=%ds activities=%s rabbit=%s",
		cfg.Env, cfg.Port, cfg.SolrURL, cfg.MemcachedAddr, cfg.CacheTTLSeconds, cfg.ActivitiesAPI, cfg.RabbitURL)

	// ---- Dependencies
	solrRepo := repository.NewSolrRepo(cfg.SolrURL)
	local := repository.NewLocalCache(10_000)
	dist := repository.NewMemcached(cfg.MemcachedAddr)
	svc := services.NewSearchService(solrRepo, local, dist, time.Duration(cfg.CacheTTLSeconds)*time.Second)

	// ---- HTTP server (Gin)
	r := gin.Default()
	r.Use(middleware.CORS())

	search := controllers.NewSearchHandler(svc)

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.GET("/search", search.Search)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// ---- RabbitMQ consumer (async)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Retry con backoff exponencial para conectarse a RabbitMQ
		maxRetries := 10
		retryDelay := 2 * time.Second
		
		var conn *amqp.Connection
		var err error
		
		for i := 0; i < maxRetries; i++ {
			conn, err = amqp.Dial(cfg.RabbitURL)
			if err == nil {
				log.Printf("[rabbit] connected after %d attempts", i+1)
				break
			}
			
			if i < maxRetries-1 {
				log.Printf("[rabbit] connection attempt %d/%d failed (%v), retrying in %v...", i+1, maxRetries, err, retryDelay)
				time.Sleep(retryDelay)
				retryDelay *= 2 // Backoff exponencial
				if retryDelay > 30*time.Second {
					retryDelay = 30 * time.Second // Máximo 30 segundos
				}
			}
		}
		
		if err != nil {
			log.Printf("[rabbit] WARN: connection failed after %d attempts (%v). search-api will still serve /search", maxRetries, err)
			return
		}
		defer conn.Close()

		consumer := consumers.NewConsumer(solrRepo, local, dist, cfg.ActivitiesAPI)
		if err := consumer.Start(ctx, conn, cfg.RabbitQueue, cfg.RabbitExchange, cfg.RabbitRoutingKey); err != nil {
			log.Printf("[rabbit] consumer error: %v", err)
		}
	}()

	// ---- Start HTTP server
	go func() {
		log.Printf("[http] listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[http] fatal: %v", err)
		}
	}()

	// ---- Graceful shutdown (SIGINT/SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("[shutdown] signal received, shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[shutdown] http error: %v", err)
	}
	log.Printf("[shutdown] bye")
}
