package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"

	"github.com/sporthub/activities-api/internal/domain"
	"github.com/sporthub/activities-api/internal/middleware"
	"github.com/sporthub/activities-api/internal/services"
)

// SessionController maneja las rutas relacionadas con las sesiones de actividades.
type SessionController struct {
	service    *services.SessionsService
	activity   *services.ActivitiesService
	rabbitChan *amqp091.Channel
}

// NewSessionController crea un nuevo controlador de sesiones.
func NewSessionController(s *services.SessionsService, a *services.ActivitiesService, ch *amqp091.Channel) *SessionController {
	return &SessionController{service: s, activity: a, rabbitChan: ch}
}

//
// ======================
// CRUD de Sesiones
// ======================
//

// POST /activities/:id/sessions
func (c *SessionController) CreateSession(ctx *gin.Context) {
	activityIDStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
		return
	}
	var s domain.Session
	if err := ctx.ShouldBindJSON(&s); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.ActivityID = activityID
	s.CreatedAt = time.Now()

	id, err := c.service.CreateSession(ctx, &s)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.ID = id

	c.publishEvent("create", s.ActivityID, s.ID)
	ctx.JSON(http.StatusCreated, s)
}

// GET /activities/:id/sessions
func (c *SessionController) GetSessionsByActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity id format"})
		return
	}
	sessions, err := c.service.GetSessionsByActivity(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, sessions)
}

// GET /sessions/:id
func (c *SessionController) GetSessionByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}
	s, err := c.service.GetSessionByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, s)
}

// PUT /sessions/:id
func (c *SessionController) UpdateSession(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}
	var update domain.Session
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.service.UpdateSession(ctx, id, &update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.publishEvent("update", update.ActivityID, id)
	ctx.JSON(http.StatusOK, update)
}

// DELETE /sessions/:id
func (c *SessionController) DeleteSession(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}
	session, err := c.service.GetSessionByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "sesión no encontrada"})
		return
	}

	if err := c.service.DeleteSession(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.publishEvent("delete", session.ActivityID, id)
	ctx.JSON(http.StatusOK, gin.H{"deleted": id})
}

//
// ======================
// Endpoint para search-api
// ======================
//

// GET /sessions/:id/search-doc
func (c *SessionController) GetSearchDoc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id format"})
		return
	}

	session, err := c.service.GetSessionByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "sesión no encontrada"})
		return
	}

	activity, err := c.activity.GetByID(ctx, session.ActivityID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "actividad no encontrada"})
		return
	}

	// Mapeamos los datos al formato que Solr espera (convierte uint64 a string para compatibilidad)
	doc := domain.SearchDoc{
		ID:         fmt.Sprintf("%d", session.ID),
		ActivityID: fmt.Sprintf("%d", activity.ID),
		SessionID:  fmt.Sprintf("%d", session.ID),
		Name:       activity.Nombre,
		Sport:      activity.Categoria, // Usamos categoria como deporte
		Site:       activity.Ubicacion,
		Instructor: activity.Instructor,
		StartAt:    session.Fecha + "T" + session.Inicio + ":00Z", // Combinamos fecha e inicio con zona horaria UTC
		EndAt:      session.Fecha + "T" + session.Fin + ":00Z",    // Combinamos fecha y fin con zona horaria UTC
		Difficulty: 1,                                            // Valor por defecto: 1 (medium), ya que no está en el dominio
		Price:      activity.PrecioBase,                          // Usamos precio base
		Tags:       []string{},                                   // Tags removido de Activity, se mantiene vacío para compatibilidad con search-api
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, doc)
}

// ======================
// Función para registrar rutas de sesiones
// ======================
func RegisterSessionRoutes(r *gin.Engine, svc *services.SessionsService, actSvc *services.ActivitiesService, rabbitChan *amqp091.Channel, jwtSecret string) {
	// Crear el controlador con todos los parámetros necesarios
	controller := NewSessionController(svc, actSvc, rabbitChan)

	// Rutas públicas
	r.GET("/sessions/:id", controller.GetSessionByID)
	r.GET("/sessions/:id/search-doc", controller.GetSearchDoc)

	// Rutas protegidas
	protected := r.Group("/sessions")
	protected.Use(middleware.JWTAuth(jwtSecret))
	protected.Use(middleware.RequireAdmin())

	protected.POST("/", controller.CreateSession)
	protected.PUT("/:id", controller.UpdateSession)
	protected.DELETE("/:id", controller.DeleteSession)

	// Rutas de actividades con sesiones
	// IMPORTANTE: Usar :id en lugar de :activityId para evitar conflicto con /activities/:id
	// Las rutas específicas con más segmentos (/activities/:id/sessions) deben registrarse
	// ANTES que las rutas genéricas (/activities/:id) para que Gin pueda distinguirlas
	activities := r.Group("/activities")
	activities.GET("/:id/sessions", controller.GetSessionsByActivity)

	protectedActivities := activities.Group("")
	protectedActivities.Use(middleware.JWTAuth(jwtSecret))
	protectedActivities.Use(middleware.RequireAdmin())
	protectedActivities.POST("/:id/sessions", controller.CreateSession)
}

// ======================
// Función auxiliar para publicar eventos
// ======================
func (c *SessionController) publishEvent(op string, activityID, sessionID uint64) {
	event := map[string]interface{}{
		"op":         op,
		"activityId": fmt.Sprintf("%d", activityID),
		"sessionId":  fmt.Sprintf("%d", sessionID),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(event)
	_ = c.rabbitChan.Publish(
		"activities.events", // exchange
		"search_sync",       // routing key
		false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

