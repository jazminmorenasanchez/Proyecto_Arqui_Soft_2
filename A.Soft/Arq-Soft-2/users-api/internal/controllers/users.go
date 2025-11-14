package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/domain"
	"github.com/sporthub/users-api/internal/services"
)

type UsersController struct {
	svc services.UsersService
}

func NewUsersController() *UsersController {
	cfg := config.Load()
	return &UsersController{svc: services.NewUsersService(cfg)}
}

type createUserReq struct {
	Username string      `json:"username" binding:"required,min=3,max=50"`
	Email    string      `json:"email"    binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6,max=72"`
	Role     domain.Role `json:"rol"     binding:"required,oneof=user normal admin"`
}

func (c *UsersController) CreateUser(ctx *gin.Context) {
	var req createUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := c.svc.Create(req.Username, req.Email, req.Password, req.Role)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"id": u.ID, "username": u.Username, "email": u.Email, "role": u.Role,
	})
}

func (c *UsersController) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	u, err := c.svc.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": u.ID, "username": u.Username, "email": u.Email, "role": u.Role,
	})
}

func (c *UsersController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.svc.Delete(id); err != nil {
		if err == services.ErrForbiddenDeleteAdmin {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "cannot delete admin user"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.Status(http.StatusNoContent)
}
