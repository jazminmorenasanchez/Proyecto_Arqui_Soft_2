package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/activities-api/internal/config"
	"github.com/sporthub/activities-api/internal/domain"
	"github.com/sporthub/activities-api/internal/middleware"
	"github.com/sporthub/activities-api/internal/services"
	"go.mongodb.org/mongo-driver/bson"
)

func RegisterActivityRoutes(r *gin.Engine, svc *services.ActivitiesService, sesSvc *services.SessionsService, cfg *config.Config) {
	// Public listing with pagination
	pub := r.Group("/activities")
	pub.GET("", func(c *gin.Context) {
		var q PaginationQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
			return
		}
		activities, total, err := svc.List(c, q.Skip, q.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list activities"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"activities": activities, "total": total, "skip": q.Skip, "limit": q.Limit})
	})

	// Individual resource (public)
	pub.GET("/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
			return
		}
		out, err := svc.GetByID(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	// GET /activities/:id/search-doc - Endpoint para search-api
	pub.GET("/:id/search-doc", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
			return
		}

		activity, err := svc.GetByID(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "activity not found"})
			return
		}

		// Mapear actividad al formato SearchDoc para Solr
		// No incluimos sesiones: solo indexamos actividades por nombre
		doc := domain.SearchDoc{
			ID:         fmt.Sprintf("%d", activity.ID),
			ActivityID: fmt.Sprintf("%d", activity.ID),
			SessionID:  "", // Siempre vacío: no indexamos sesiones
			Name:       activity.Nombre,
			Sport:      activity.Categoria,
			Site:       activity.Ubicacion,
			Instructor: activity.Instructor,
			StartAt:    "", // Vacío: no usamos fechas de sesiones en la búsqueda
			EndAt:      "", // Vacío: no usamos fechas de sesiones en la búsqueda
			Difficulty: 1,  // Valor por defecto
			Price:      activity.PrecioBase,
			Tags:       []string{},
			UpdatedAt:  activity.UpdatedAt.Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, doc)
	})

	// Protected admin routes
	g := r.Group("/activities")
	g.Use(middleware.JWTAuth(cfg.JWTSecret))
	g.POST("", middleware.RequireAdmin(), func(c *gin.Context) {
		var req CreateActivityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Extract userId from JWT context
		userId, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user id in token"})
			return
		}

		// Convert userId to string (it's stored as uint64 in context)
		var userIdStr string
		switch v := userId.(type) {
		case uint64:
			userIdStr = fmt.Sprintf("%d", v)
		case string:
			userIdStr = v
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
			return
		}

		activity := &domain.Activity{
			OwnerUserID: userIdStr,
			Categoria:   req.Categoria,
			Nombre:      req.Nombre,
			Ubicacion:   req.Ubicacion,
			Instructor:  req.Instructor,
			PrecioBase:  req.PrecioBase,
		}
		id, err := svc.Create(c, activity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	g.PUT("/:id", middleware.RequireAdmin(), func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
			return
		}
		
		// Leer el body manualmente para verificar si instructor viene en el JSON
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}
		
		// Parsear a map para verificar campos presentes
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
			return
		}
		
		// Parsear a struct para validación
		var req CreateActivityRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Validar campos requeridos
		if req.Categoria == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "categoria is required"})
			return
		}
		if req.Nombre == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nombre is required"})
			return
		}
		if req.Ubicacion == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ubicacion is required"})
			return
		}
		if req.PrecioBase <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "precioBase is required and must be greater than 0"})
			return
		}
		
		// Construir update con campos requeridos
		update := bson.M{
			"categoria": req.Categoria,
			"nombre":    req.Nombre,
			"ubicacion": req.Ubicacion,
			"precioBase": req.PrecioBase,
		}
		
		// Solo actualizar instructor si viene en el request
		if _, exists := jsonData["instructor"]; exists {
			update["instructor"] = req.Instructor
		}
		
		if err := svc.Update(c, id, update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	g.DELETE("/:id", middleware.RequireAdmin(), func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
			return
		}
		if err := svc.Delete(c, id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Endpoint para reindexar todas las actividades en Solr
	g.POST("/reindex", middleware.RequireAdmin(), func(c *gin.Context) {
		count, err := svc.ReindexAll(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Reindexing triggered for %d activities", count), "count": count})
	})
}
