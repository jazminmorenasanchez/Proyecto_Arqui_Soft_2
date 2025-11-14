package tests

import (
	"errors"

	"github.com/sporthub/users-api/internal/domain"
	"github.com/sporthub/users-api/internal/services"
)

// mockUsersService is a mock implementation of UsersService for testing
type mockUsersService struct {
	users map[string]*domain.User
}

func NewMockUsersService() services.UsersService {
	return &mockUsersService{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUsersService) Create(username, email, password string, role domain.Role) (*domain.User, error) {
	if _, exists := m.users[username]; exists {
		return nil, errors.New("username already exists")
	}
	user := &domain.User{
		ID:       uint64(len(m.users) + 1),
		Username: username,
		Email:    email,
		Role:     role,
	}
	m.users[username] = user
	return user, nil
}

func (m *mockUsersService) GetByID(id uint64) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUsersService) Login(login, password string) (*domain.User, string, error) {
	if user, exists := m.users[login]; exists && password == "validpass" {
		return user, "test-token", nil
	}
	return nil, "", errors.New("invalid credentials")
}

func (m *mockUsersService) Delete(id uint64) error {
	for username, user := range m.users {
		if user.ID == id {
			if user.Role == domain.RoleAdmin {
				return services.ErrForbiddenDeleteAdmin
			}
			delete(m.users, username)
			return nil
		}
	}
	return errors.New("not found")
}

// Helper to cast to concrete type for tests that need direct access
func toConcreteMock(svc services.UsersService) *mockUsersService {
	if m, ok := svc.(*mockUsersService); ok {
		return m
	}
	return nil
}
