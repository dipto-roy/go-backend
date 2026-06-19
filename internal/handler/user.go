package handler

import (
	"time"

	"github.com/dip-roy/go-backend/internal/middleware"
	"github.com/dip-roy/go-backend/internal/model"
	"github.com/dip-roy/go-backend/internal/service"
	"github.com/dip-roy/go-backend/pkg/apperror"
	"github.com/dip-roy/go-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserResponse(u *model.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

type updateProfileRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// GetMe godoc
// @Summary Get current user profile
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=userResponse}
// @Router /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toUserResponse(user))
}

// UpdateMe godoc
// @Summary Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body updateProfileRequest true "Profile update"
// @Success 200 {object} response.Response{data=userResponse}
// @Router /users/me [put]
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, service.UpdateProfileInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toUserResponse(user))
}

// ChangePassword godoc
// @Summary Change current user password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body changePasswordRequest true "Password change"
// @Success 200 {object} response.Response
// @Router /users/me/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, formatValidationErrors(err))
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, service.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, gin.H{"message": "password changed"})
}

// DeleteMe godoc
// @Summary Delete current user account
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /users/me [delete]
func (h *Handler) DeleteMe(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	if err := h.userService.Delete(c.Request.Context(), userID); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, gin.H{"message": "account deleted"})
}

func currentUserID(c *gin.Context) (uuid.UUID, error) {
	idStr, exists := c.Get(middleware.AuthUserIDKey)
	if !exists {
		return uuid.Nil, apperror.ErrUnauthorized
	}
	id, err := uuid.Parse(idStr.(string))
	if err != nil {
		return uuid.Nil, apperror.ErrUnauthorized
	}
	return id, nil
}
