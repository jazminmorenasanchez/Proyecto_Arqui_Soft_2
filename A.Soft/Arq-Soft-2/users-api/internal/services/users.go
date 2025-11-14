package services

import (
	"errors"

	"github.com/sporthub/users-api/internal/config"
	"github.com/sporthub/users-api/internal/domain"
	"github.com/sporthub/users-api/internal/repository"
	"github.com/sporthub/users-api/internal/utils"
)

type UsersService interface {
	Create(username, email, password string, role domain.Role) (*domain.User, error)
	GetByID(id uint64) (*domain.User, error)
	Login(login, password string) (*domain.User, string, error)
	Delete(id uint64) error
}

type usersSvc struct {
	repo repository.UsersRepo
	cfg  config.Config
}

func NewUsersService(cfg config.Config) UsersService {
	return &usersSvc{repo: repository.NewUsersMySQL(cfg), cfg: cfg}
}

func (s *usersSvc) Create(username, email, password string, role domain.Role) (*domain.User, error) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *usersSvc) GetByID(id uint64) (*domain.User, error) {
	return s.repo.FindByID(id)
}

func (s *usersSvc) Login(login, password string) (*domain.User, string, error) {
	u, err := s.repo.FindByUsernameOrEmail(login)
	if err != nil {
		return nil, "", err
	}
	if u == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, u.PasswordHash) {
		return nil, "", errors.New("invalid credentials")
	}
	token, err := utils.GenerateJWT(s.cfg, u)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

func (s *usersSvc) Delete(id uint64) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if u == nil {
		return errors.New("not found")
	}
	if u.Role == domain.RoleAdmin {
		return ErrForbiddenDeleteAdmin
	}
	return s.repo.DeleteByID(id)
}

var ErrForbiddenDeleteAdmin = errors.New("cannot delete admin user")
