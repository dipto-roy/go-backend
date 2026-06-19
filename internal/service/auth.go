package service

import (
	"context"
	"time"

	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/internal/repository"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/dip-roy/go-backend/pkg/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthOutput struct {
	User         *model.User
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*AuthOutput, error)
	Login(ctx context.Context, in LoginInput) (*AuthOutput, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthOutput, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSecret        string
	accessExpiry     time.Duration
	refreshExpiry    time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtSecret string,
	accessExpiry, refreshExpiry time.Duration,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessExpiry:     accessExpiry,
		refreshExpiry:    refreshExpiry,
	}
}

func (s *authService) Register(ctx context.Context, in RegisterInput) (*AuthOutput, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	user := &model.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.buildAuthOutput(ctx, user)
}

func (s *authService) Login(ctx context.Context, in LoginInput) (*AuthOutput, error) {
	user, err := s.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		// mask not found as unauthorized to prevent email enumeration
		return nil, apperror.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, apperror.ErrUnauthorized
	}

	return s.buildAuthOutput(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, refreshTokenStr string) (*AuthOutput, error) {
	claims, err := token.Verify(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return nil, apperror.ErrTokenInvalid
	}

	rt, err := s.refreshTokenRepo.FindByToken(ctx, refreshTokenStr)
	if err != nil || rt.IsExpired() {
		return nil, apperror.ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, apperror.ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.ErrUnauthorized
	}

	if err := s.refreshTokenRepo.DeleteByToken(ctx, refreshTokenStr); err != nil {
		return nil, apperror.ErrInternal
	}

	return s.buildAuthOutput(ctx, user)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshTokenRepo.DeleteByToken(ctx, refreshToken)
}

func (s *authService) buildAuthOutput(ctx context.Context, user *model.User) (*AuthOutput, error) {
	pair, err := token.Generate(user.ID.String(), user.Email, s.jwtSecret, s.accessExpiry, s.refreshExpiry)
	if err != nil {
		return nil, apperror.ErrInternal
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     pair.RefreshToken,
		ExpiresAt: time.Now().Add(s.refreshExpiry),
	}
	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, apperror.ErrInternal
	}

	return &AuthOutput{
		User:         user,
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
