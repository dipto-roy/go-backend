package repository

import (
	"context"
	"errors"

	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperror.ErrInternal
	}
	return nil
}

func (r *refreshTokenRepo) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal
	}
	return &rt, nil
}

func (r *refreshTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error; err != nil {
		return apperror.ErrInternal
	}
	return nil
}

func (r *refreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{}).Error; err != nil {
		return apperror.ErrInternal
	}
	return nil
}
