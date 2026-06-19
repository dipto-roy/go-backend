package service_test

import (
	"context"
	"testing"

	"github.com/dip-roy/go-backend/internal/model"
	repomock "github.com/dip-roy/go-backend/internal/repository/mock"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestGetByID_Found(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	expected := &model.User{ID: id, Email: "alice@example.com", Name: "Alice"}
	userRepo.On("FindByID", mock.Anything, id).Return(expected, nil)

	user, err := svc.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, expected.Email, user.Email)
}

func TestGetByID_NotFound(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	userRepo.On("FindByID", mock.Anything, id).Return(nil, apperror.ErrNotFound)

	_, err := svc.GetByID(context.Background(), id)
	assert.True(t, apperror.Is(err, apperror.ErrNotFound))
}

func TestUpdateProfile_Success(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	existing := &model.User{ID: id, Email: "old@example.com", Name: "Old Name"}
	userRepo.On("FindByID", mock.Anything, id).Return(existing, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	updated, err := svc.UpdateProfile(context.Background(), id, service.UpdateProfileInput{
		Name:  "New Name",
		Email: "new@example.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "new@example.com", updated.Email)
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	userRepo := new(repomock.UserRepository)
	svc := service.NewUserService(userRepo)

	id := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &model.User{ID: id, PasswordHash: string(hash)}
	userRepo.On("FindByID", mock.Anything, id).Return(user, nil)

	err := svc.ChangePassword(context.Background(), id, service.ChangePasswordInput{
		CurrentPassword: "wrong",
		NewPassword:     "newpassword123",
	})

	assert.Error(t, err)
}
