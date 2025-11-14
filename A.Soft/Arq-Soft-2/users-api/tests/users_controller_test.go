package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/sporthub/users-api/internal/domain"
	"github.com/sporthub/users-api/internal/services"
)

func TestDeleteUser(t *testing.T) {
	mockSvc := NewMockUsersService().(*mockUsersService)

	// Create a regular user
	user, _ := mockSvc.Create("testuser", "test@example.com", "validpass", domain.RoleUser)

	// Test deleting regular user
	err := mockSvc.Delete(user.ID)
	if err != nil {
		t.Errorf("Delete() failed for regular user: %v", err)
	}

	// Verify user was deleted
	if _, exists := mockSvc.users["testuser"]; exists {
		t.Error("User should have been deleted")
	}

	// Create an admin user
	admin, _ := mockSvc.Create("admin", "admin@example.com", "validpass", domain.RoleAdmin)

	// Test deleting admin user (should fail)
	err = mockSvc.Delete(admin.ID)
	if err == nil {
		t.Error("Delete() should fail for admin user")
	}
	if err != services.ErrForbiddenDeleteAdmin {
		t.Errorf("Expected ErrForbiddenDeleteAdmin, got: %v", err)
	}

	// Verify admin still exists
	if _, exists := mockSvc.users["admin"]; !exists {
		t.Error("Admin user should not have been deleted")
	}
}

func TestCreateUser(t *testing.T) {
	_ = httptest.NewRecorder
	// This test validates the mock service works correctly
	mockSvc := NewMockUsersService()

	user, err := mockSvc.Create("testuser", "test@example.com", "pass123", domain.RoleUser)
	if err != nil {
		t.Errorf("Create() failed: %v", err)
	}
	if user == nil {
		t.Error("Create() returned nil user")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
}
