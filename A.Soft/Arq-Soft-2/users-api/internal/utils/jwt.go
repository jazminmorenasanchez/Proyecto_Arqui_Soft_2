package utils

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/domain"
)

func GenerateJWT(cfg config.Config, u *domain.User) (string, error) {
	min, _ := strconv.Atoi(cfg.JWTExpMinutes)
	claims := jwt.MapClaims{
		"sub": u.ID,
		"rol": u.Role,
		"exp": time.Now().Add(time.Duration(min) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
