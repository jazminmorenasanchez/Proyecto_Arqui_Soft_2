package controllers

import (
	"net/http"
	"strconv"

	"github.com/sporthub/search-api/internal/services"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *services.Service }

func NewSearchHandler(s *services.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("query")
	sport := c.Query("sport")
	site := c.Query("site")
	date := c.Query("date") // yyyy-mm-dd
	sort := c.DefaultQuery("sort", "start_dt asc")
	page := atoi(c.DefaultQuery("page", "1"))
	size := atoi(c.DefaultQuery("size", "10"))
	
	// Validación de parámetros
	if size > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "size cannot exceed 100"})
		return
	}
	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be >= 1"})
		return
	}

	res, err := h.svc.Search(c.Request.Context(), q, sport, site, date, sort, page, size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil || i <= 0 {
		return 1
	}
	return i
}
