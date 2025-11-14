package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/services"
)

type AuthController struct {
	svc services.UsersService
}

func NewAuthController(cfg config.Config) *AuthController {
	return &AuthController{svc: services.NewUsersService(cfg)}
}

type loginReq struct {
	Login    string `json:"login"    binding:"required"` // username o email
	Password string `json:"password" binding:"required"`
}

func (a *AuthController) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, token, err := a.svc.Login(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"role":   u.Role,
		"userId": u.ID,
	})
}
