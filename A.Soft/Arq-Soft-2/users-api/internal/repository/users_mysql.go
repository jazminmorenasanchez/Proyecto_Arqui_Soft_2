package repository

import (
	"errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/domain"
)

type UsersRepo interface {
	Create(u *domain.User) error
	FindByID(id uint64) (*domain.User, error)
	FindByUsernameOrEmail(login string) (*domain.User, error)
	DeleteByID(id uint64) error
}

type usersMySQL struct{ gdb *gorm.DB }

func NewUsersMySQL(cfg config.Config) UsersRepo {
	dsn := cfg.MySQLUser + ":" + cfg.MySQLPassword + "@tcp(" + cfg.MySQLHost + ":" + cfg.MySQLPort + ")/" + cfg.MySQLDB + "?parseTime=true&charset=utf8mb4&loc=Local"
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &usersMySQL{gdb: gdb}
}

func (r *usersMySQL) Create(u *domain.User) error {
	return r.gdb.Create(u).Error
}

func (r *usersMySQL) FindByID(id uint64) (*domain.User, error) {
	var u domain.User
	if err := r.gdb.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *usersMySQL) FindByUsernameOrEmail(login string) (*domain.User, error) {
	var u domain.User
	err := r.gdb.Where("username = ? OR email = ?", login, login).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *usersMySQL) DeleteByID(id uint64) error {
	return r.gdb.Delete(&domain.User{}, id).Error
}
