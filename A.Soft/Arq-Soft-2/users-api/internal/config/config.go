package config

import (
	"log"
	"os"
)

type Config struct {
	AppPort       string
	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDB       string

	JWTSecret     string
	JWTExpMinutes string
}

func Load() Config {
	cfg := Config{
		AppPort:       getEnv("APP_PORT", "8081"),
		MySQLHost:     getEnv("MYSQL_HOST", "mysql"),
		MySQLPort:     getEnv("MYSQL_PORT", "3306"),
		MySQLUser:     getEnv("MYSQL_USER", "root"),
		MySQLPassword: getEnv("MYSQL_PASSWORD", "secret"),
		MySQLDB:       getEnv("MYSQL_DB", "sporthub_users"),
		JWTSecret:     getEnv("JWT_SECRET", "change_me"),
		JWTExpMinutes: getEnv("JWT_EXP_MINUTES", "60"),
	}
	log.Printf("config loaded: APP_PORT=%s MYSQL=%s:%s/%s", cfg.AppPort, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB)
	return cfg
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
