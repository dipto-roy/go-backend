package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dip-roy/go-backend/internal/handler"
	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, in service.RegisterInput) (*service.AuthOutput, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthOutput), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, in service.LoginInput) (*service.AuthOutput, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthOutput), args.Error(1)
}

func (m *mockAuthService) Refresh(ctx context.Context, token string) (*service.AuthOutput, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthOutput), args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetByID(ctx context.Context, id interface{}) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserService) UpdateProfile(ctx context.Context, id interface{}, in service.UpdateProfileInput) (*model.User, error) {
	args := m.Called(ctx, id, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserService) ChangePassword(ctx context.Context, id interface{}, in service.ChangePasswordInput) error {
	args := m.Called(ctx, id, in)
	return args.Error(0)
}

func (m *mockUserService) Delete(ctx context.Context, id interface{}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupRouter(h *handler.Handler) *gin.Engine {
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)
	return r
}

func TestRegisterHandler_Success(t *testing.T) {
	authSvc := new(mockAuthService)
	h := handler.New(authSvc, nil)
	r := setupRouter(h)

	out := &service.AuthOutput{
		User:         &model.User{Email: "alice@example.com", Name: "Alice"},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}
	authSvc.On("Register", mock.Anything, service.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	}).Return(out, nil)

	body, _ := json.Marshal(map[string]string{
		"name":     "Alice",
		"email":    "alice@example.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	authSvc := new(mockAuthService)
	h := handler.New(authSvc, nil)
	r := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "not-an-email"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	authSvc := new(mockAuthService)
	h := handler.New(authSvc, nil)
	r := setupRouter(h)

	authSvc.On("Login", mock.Anything, service.LoginInput{
		Email:    "bob@example.com",
		Password: "wrong",
	}).Return(nil, apperror.ErrUnauthorized)

	body, _ := json.Marshal(map[string]string{
		"email":    "bob@example.com",
		"password": "wrong",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
