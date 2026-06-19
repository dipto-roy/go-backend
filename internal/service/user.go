package service

import (
	"context"

	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/internal/repository"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UpdateProfileInput struct {
	Name  string
	Email string
}

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (*model.User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, in ChangePasswordInput) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = in.Name
	user.Email = in.Email

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, in ChangePasswordInput) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.CurrentPassword)); err != nil {
		return apperror.New(400, "WRONG_PASSWORD", "current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.ErrInternal
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(ctx, user)
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.Delete(ctx, id)
}
