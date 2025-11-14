package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/activities-api/internal/middleware"
	"github.com/sporthub/activities-api/internal/services"
)

type enrollReq struct {
	SessionID string `json:"sessionId" binding:"required"`
}

func RegisterEnrollmentRoutes(r *gin.Engine, svc *services.EnrollmentsService, jwtSecret string) {
	// RUTA PÚBLICA: fuera del grupo protegido
	r.GET("/enrollments/by-user/:userId", func(c *gin.Context) {
		out, err := svc.ListByUser(c, c.Param("userId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, out)
	})

	// GRUPO PROTEGIDO: todas las rutas aquí requieren autenticación
	g := r.Group("/enrollments")
	g.Use(middleware.JWTAuth(jwtSecret))
	g.Use(middleware.RequireUser())

	// Protect enroll endpoint: user role required
	g.POST("", func(c *gin.Context) {
		var req enrollReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Extract userId from JWT context (security: prevent impersonation)
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
		
		// Parse sessionID from string to uint64
		sessionID, err := strconv.ParseUint(req.SessionID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
			return
		}
		
		id, err := svc.Enroll(c, sessionID, userIdStr)
		if err != nil {
			if errors.Is(err, services.ErrAlreadyEnrolled) {
				c.JSON(http.StatusConflict, gin.H{"error": "ya inscripto en esta sesión"})
				return
			}
			if errors.Is(err, services.ErrNoCupo) {
				c.JSON(http.StatusConflict, gin.H{"error": "no hay cupo disponible"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	// Cancel enrollment (owner or admin)
	g.PATCH("/:id/cancel", func(c *gin.Context) {
		idStr := c.Param("id")
		enrollmentID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid enrollment id format"})
			return
		}
		// userId stored as uint64 in context
		var requester string
		if v, ok := c.Get("userId"); ok {
			switch vv := v.(type) {
			case uint64:
				requester = fmt.Sprintf("%d", vv)
			case string:
				requester = vv
			}
		}
		role := ""
		if r, ok := c.Get("role"); ok {
			if rs, ok := r.(string); ok {
				role = rs
			}
		}
		if requester == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
			return
		}
		if err := svc.CancelEnrollment(c, enrollmentID, requester, role); err != nil {
			if errors.Is(err, services.ErrForbidden) {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
