package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/sporthub/users-api/internal/domain"
	"github.com/sporthub/users-api/internal/services"
)

func setupTestRouter(svc services.UsersService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authCtl := &AuthController{svc: svc}
	r.POST("/auth/login", authCtl.Login)
	r.POST("/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Username == "" || req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
			return
		}
		_, err := svc.Create(req.Username, req.Email, req.Password, domain.RoleUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	return r
}

type AuthController struct {
	svc interface {
		Create(username, email, password string, role domain.Role) (*domain.User, error)
		GetByID(id uint64) (*domain.User, error)
		Login(login, password string) (*domain.User, string, error)
	}
}

func (a *AuthController) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Login == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}
	u, token, err := a.svc.Login(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "userId": u.Username})
}

// mockUsersService is now defined in common.go

func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		setupMock      func(services.UsersService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful login",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "validpass",
			},
			setupMock: func(m services.UsersService) {
				m.Create("testuser", "test@example.com", "validpass", domain.RoleUser)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"token":  "test-token",
				"userId": "testuser",
			},
		},
		{
			name: "invalid credentials",
			requestBody: map[string]string{
				"login":    "testuser",
				"password": "wrongpass",
			},
			setupMock: func(m services.UsersService) {
				m.Create("testuser", "test@example.com", "validpass", domain.RoleUser)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
		{
			name: "missing login field",
			requestBody: map[string]string{
				"password": "validpass",
			},
			setupMock:      func(services.UsersService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password field",
			requestBody: map[string]string{
				"login": "testuser",
			},
			setupMock:      func(services.UsersService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := NewMockUsersService()
			tt.setupMock(mockSvc)
			router := setupTestRouter(mockSvc)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d but got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != nil {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				for k, v := range tt.expectedBody {
					if response[k] != v {
						t.Errorf("expected %s to be %v but got %v", k, v, response[k])
					}
				}
			}
		})
	}
}

func TestRegister(t *testing.T) {
	mockSvc := NewMockUsersService()
	r := setupTestRouter(mockSvc)

	tests := []struct {
		name       string
		reqBody    map[string]interface{}
		wantStatus int
	}{
		{
			name: "valid registration",
			reqBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "pass123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing username",
			reqBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "pass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			reqBody: map[string]interface{}{
				"username": "testuser",
				"password": "pass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			reqBody: map[string]interface{}{
				"username": "testuser",
				"email":    "test@example.com",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
