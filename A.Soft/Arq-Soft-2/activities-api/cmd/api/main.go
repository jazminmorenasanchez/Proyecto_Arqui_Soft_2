package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/activities-api/internal/clients"
	"github.com/sporthub/activities-api/internal/config"
	"github.com/sporthub/activities-api/internal/controllers"
	"github.com/sporthub/activities-api/internal/db"
	"github.com/sporthub/activities-api/internal/repository"
	"github.com/sporthub/activities-api/internal/services"
	"github.com/sporthub/activities-api/internal/utils"
)

func main() {
	cfg := config.Load()

	// Wait for RabbitMQ to be ready
	if err := utils.WaitForService("rabbitmq", 5672, 60); err != nil {
		log.Fatalf("Failed to wait for RabbitMQ: %v", err)
	}

	// DB
	mc, mdb := db.MustMongo(cfg.MongoURI, cfg.MongoDB)
	defer mc.Disconnect(cfg.Ctx)

	// RabbitMQ
	rmq := clients.MustRabbit(cfg.RabbitURL, cfg.Exchange, cfg.ExchangeType)
	defer rmq.Close()

	// Users client
	users := clients.NewUsersClient(cfg.UsersAPIBase)

	// Repos
	actRepo := repository.NewActivitiesMongo(mdb)
	sesRepo := repository.NewSessionsMongo(mdb)
	enrRepo := repository.NewEnrollmentsMongo(mdb)

	// Services
	actSvc := services.NewActivitiesService(actRepo, users, rmq, cfg)
	sesSvc := services.NewSessionsService(sesRepo, actRepo, rmq, cfg)
	enrSvc := services.NewEnrollmentsService(enrRepo, sesRepo, actRepo, rmq, cfg)

	// Router
	r := gin.Default()
	r.Use(controllers.CORSMiddleware())

	// IMPORTANTE: Registrar rutas más específicas PRIMERO
	// RegisterSessionRoutes registra /activities/:activityId/sessions
	// que debe registrarse ANTES de /activities/:id
	controllers.RegisterSessionRoutes(r, sesSvc, actSvc, rmq.GetChannel(), cfg.JWTSecret)
	controllers.RegisterActivityRoutes(r, actSvc, sesSvc, cfg)
	controllers.RegisterEnrollmentRoutes(r, enrSvc, cfg.JWTSecret)

	port := cfg.Port
	if port == "" {
		port = "8082"
	}
	log.Println("activities-api listening on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
