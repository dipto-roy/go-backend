package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dip-roy/go-backend/internal/model"
	repomock "github.com/dip-roy/go-backend/internal/repository/mock"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

const (
	testSecret        = "test-secret-key-that-is-at-least-32-chars!!"
	testAccessExpiry  = 15 * time.Minute
	testRefreshExpiry = 7 * 24 * time.Hour
)

func newAuthSvc(userRepo *repomock.UserRepository, rtRepo *repomock.RefreshTokenRepository) service.AuthService {
	return service.NewAuthService(userRepo, rtRepo, testSecret, testAccessExpiry, testRefreshExpiry)
}

func TestRegister_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
	rtRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	out, err := svc.Register(context.Background(), service.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "supersecret",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
	assert.Equal(t, "alice@example.com", out.User.Email)
	userRepo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
}

func TestRegister_Conflict(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(apperror.ErrConflict)

	_, err := svc.Register(context.Background(), service.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "supersecret",
	})

	assert.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrConflict))
}

func TestLogin_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("mypassword"), bcrypt.DefaultCost)
	userID := uuid.New()
	user := &model.User{ID: userID, Email: "bob@example.com", PasswordHash: string(hash)}

	userRepo.On("FindByEmail", mock.Anything, "bob@example.com").Return(user, nil)
	rtRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	out, err := svc.Login(context.Background(), service.LoginInput{
		Email:    "bob@example.com",
		Password: "mypassword",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	user := &model.User{ID: uuid.New(), Email: "bob@example.com", PasswordHash: string(hash)}

	userRepo.On("FindByEmail", mock.Anything, "bob@example.com").Return(user, nil)

	_, err := svc.Login(context.Background(), service.LoginInput{
		Email:    "bob@example.com",
		Password: "wrongpass",
	})

	assert.Error(t, err)
	assert.True(t, apperror.Is(err, apperror.ErrUnauthorized))
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userRepo.On("FindByEmail", mock.Anything, "ghost@example.com").Return(nil, apperror.ErrNotFound)

	_, err := svc.Login(context.Background(), service.LoginInput{
		Email:    "ghost@example.com",
		Password: "whatever",
	})

	// must return unauthorized (not not_found) to prevent email enumeration
	assert.True(t, apperror.Is(err, apperror.ErrUnauthorized))
}

func TestRefresh_InvalidToken(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	_, err := svc.Refresh(context.Background(), "not.a.valid.token")
	assert.True(t, apperror.Is(err, apperror.ErrTokenInvalid))
}

func TestLogout_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	rtRepo.On("DeleteByToken", mock.Anything, "some-refresh-token").Return(nil)

	err := svc.Logout(context.Background(), "some-refresh-token")
	assert.NoError(t, err)
	rtRepo.AssertExpectations(t)
}

func TestRegister_InternalError(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(apperror.ErrInternal)

	_, err := svc.Register(context.Background(), service.RegisterInput{
		Name: "Test", Email: "t@t.com", Password: "pass1234",
	})
	assert.Error(t, err)
}

