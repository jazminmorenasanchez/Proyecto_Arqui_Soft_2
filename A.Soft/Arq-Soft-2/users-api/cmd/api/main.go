package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/controllers"
	"github.com/sporthub/users-api/internal/db"
	"github.com/sporthub/users-api/internal/middleware"
	"github.com/sporthub/users-api/internal/utils"
)

func main() {
	cfg := config.Load()

	// Wait for MySQL to be ready
	mysqlHost := os.Getenv("MYSQL_HOST")
	if mysqlHost == "" {
		mysqlHost = "mysql" // default hostname in docker-compose
	}
	if err := utils.WaitForService(mysqlHost, 3306, 60); err != nil {
		log.Fatalf("Failed to wait for MySQL: %v", err)
	}

	// DB
	sqlDB := db.MustInitMySQL(cfg)
	defer sqlDB.Close()

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())

	// Routes
	api := r.Group("/")

	authCtl := controllers.NewAuthController(cfg)
	userCtl := controllers.NewUsersController()

	api.POST("/auth/login", authCtl.Login)
	api.POST("/users", userCtl.CreateUser)
	api.GET("/users/:id", userCtl.GetByID)

	// Protected admin-only route for deleting users
	protected := r.Group("/")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.Use(middleware.AdminOnly())
	protected.DELETE("/users/:id", userCtl.Delete)

	addr := ":" + cfg.AppPort
	log.Printf("users-api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
