package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dip-roy/go-backend/internal/model"
	repomock "github.com/dip-roy/go-backend/internal/repository/mock"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/dip-roy/go-backend/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRefresh_ValidToken(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userID := uuid.New()
	pair, err := token.Generate(userID.String(), "u@u.com", testSecret, testAccessExpiry, testRefreshExpiry)
	assert.NoError(t, err)

	storedRT := &model.RefreshToken{
		Token:     pair.RefreshToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(testRefreshExpiry),
	}
	user := &model.User{ID: userID, Email: "u@u.com", Name: "U"}

	rtRepo.On("FindByToken", mock.Anything, pair.RefreshToken).Return(storedRT, nil)
	userRepo.On("FindByID", mock.Anything, userID).Return(user, nil)
	rtRepo.On("DeleteByToken", mock.Anything, pair.RefreshToken).Return(nil)
	rtRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(nil)

	out, err := svc.Refresh(context.Background(), pair.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.AccessToken)
	assert.NotEmpty(t, out.RefreshToken)
}

func TestRefresh_ExpiredDBToken(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	rtRepo := new(repomock.RefreshTokenRepository)
	svc := newAuthSvc(userRepo, rtRepo)

	userID := uuid.New()
	pair, _ := token.Generate(userID.String(), "u@u.com", testSecret, testAccessExpiry, testRefreshExpiry)

	// stored token already expired
	storedRT := &model.RefreshToken{
		Token:     pair.RefreshToken,
		UserID:    userID,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	rtRepo.On("FindByToken", mock.Anything, pair.RefreshToken).Return(storedRT, nil)

	_, err := svc.Refresh(context.Background(), pair.RefreshToken)
	assert.True(t, apperror.Is(err, apperror.ErrTokenInvalid))
}

func TestChangePassword_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	user := &model.User{ID: id, PasswordHash: string(hash)}
	userRepo.On("FindByID", mock.Anything, id).Return(user, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	err := svc.ChangePassword(context.Background(), id, service.ChangePasswordInput{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass123",
	})
	assert.NoError(t, err)
}

func TestDelete_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	userRepo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)
	assert.NoError(t, err)
}
