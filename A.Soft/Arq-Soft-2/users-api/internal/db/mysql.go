package db

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/domain"
)

func MustInitMySQL(cfg config.Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.MySQLUser, cfg.MySQLPassword, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB)

	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("mysql connect error: %v", err)
	}

	// Auto-migrate
	if err := gdb.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("mysql migrate error: %v", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatalf("mysql db error: %v", err)
	}
	return sqlDB
}
